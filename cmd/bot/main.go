package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LalatinaHub/LatinaBot/config"
	"github.com/LalatinaHub/LatinaBot/internal/api"
	"github.com/LalatinaHub/LatinaBot/internal/bot"
	"github.com/LalatinaHub/LatinaBot/internal/cron"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/cloudflare"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/internal/service/ocr"
	"github.com/LalatinaHub/LatinaBot/internal/service/payment"
	"github.com/LalatinaHub/LatinaBot/internal/service/proxycheck"
	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"github.com/LalatinaHub/common/database"
	tele "gopkg.in/telebot.v3"
)

func main() {
	// 1. Load application configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// 2. Setup structured logging
	logLevel := "info"
	if cfg.IsDevelopment() {
		logLevel = "debug"
	}
	logger.SetupLogger(logLevel, cfg.IsProduction())
	logger.Info().
		Str("app_env", cfg.AppEnv).
		Str("port", cfg.Port).
		Int64("admin_id", cfg.AdminID).
		Bool("development", cfg.IsDevelopment()).
		Msg("Starting LatinaBot (Go 1.24+)...")

	// Ensure temp directory exists
	_ = os.MkdirAll("./temp", 0755)

	// 3. Connect to Turso LibSQL Database pool from common
	initCtx, initCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer initCancel()

	db, err := database.GetDB()
	if err != nil {
		logger.Warn().Err(err).Msg("Turso database connection not established, running in degraded mode")
	} else {
		logger.Info().Msg("Connected to Turso LibSQL database pool")
		if err := database.InitIndexes(initCtx); err != nil {
			logger.Warn().Err(err).Msg("Failed to verify/create database indexes")
		} else {
			logger.Info().Msg("Database indexes verified successfully")
		}
	}

	// 4. Wire repository layer
	repos := &repository.Repositories{}
	if db != nil {
		repos.User = repository.NewUserRepository(db)
		repos.Server = repository.NewServerRepository(db)
		repos.Proxy = repository.NewProxyRepository(db)
		repos.Wildcard = repository.NewWildcardRepository(db)
		repos.Donation = repository.NewDonationRepository(db)
		repos.KV = repository.NewKVRepository(db)
	}

	// 5. Wire domain services
	tarotSvc := tarot.NewService()
	trakteerSvc := trakteer.NewClient(cfg.TrakteerToken)
	paymentSvc := payment.NewService(cfg.GatewayPaymentKey)
	edgeClient := edgeserver.NewClient()
	cfClient := cloudflare.NewClient(
		cfg.CloudflareEmail,
		cfg.CloudflareAPIKey,
		cfg.CloudflareAccountID,
		cfg.CloudflareZoneID,
		cfg.WorkerServiceName,
		cfg.WorkerEnvironment,
	)
	proxyChecker := proxycheck.NewChecker()

	visionCtx, visionCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer visionCancel()
	ocrScanner, err := ocr.NewVisionScanner(visionCtx, cfg.ServiceAccountURL)
	if err != nil {
		logger.Warn().Err(err).Msg("Vision OCR scanner operating in degraded mode")
	}

	// 6. Initialize Telegram Bot
	telegramBot, err := bot.NewBot(bot.Config{
		AppConfig:   cfg,
		Repos:       repos,
		TarotSvc:    tarotSvc,
		TrakteerSvc: trakteerSvc,
		PaymentSvc:  paymentSvc,
		EdgeClient:  edgeClient,
		CFClient:    cfClient,
		OCRScanner:  ocrScanner,
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize Telegram Bot")
	}

	// 7. Start Telegram Bot in background
	go func() {
		telegramBot.Start()
	}()

	// 8. Initialize and start HTTP server (for /api/v1/proxy/check)
	httpServer := api.NewServer(cfg.HTTPAddress(), proxyChecker)
	go func() {
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal().Err(err).Msg("HTTP server failed to listen")
		}
	}()

	// 9. Initialize and start Cron scheduler
	cronScheduler := cron.NewScheduler(cron.Config{
		Repos:      repos,
		CmdHandler: telegramBot.CmdHandler(),
		Bot:        telegramBot.TeleBot(),
	})
	cronScheduler.Start()

	// Notify Admin that Bot is ready
	if cfg.AdminID != 0 {
		_, _ = telegramBot.TeleBot().Send(&tele.Chat{ID: cfg.AdminID}, "Bot ready! (Go 1.24+)")
	}
	logger.Info().Msg("Bot ready!")

	// 10. Wait for graceful shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logger.Info().Msg("Shutting down LatinaBot gracefully...")

	// Stop Bot & Cron
	telegramBot.Stop()
	<-cronScheduler.Stop().Done()

	// Shutdown HTTP Server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("HTTP server forced to shutdown")
	}

	// Close database pool cleanly
	if err := database.Close(); err != nil {
		logger.Warn().Err(err).Msg("Failed to close database pool cleanly")
	}

	logger.Info().Msg("LatinaBot service exited cleanly")
}
