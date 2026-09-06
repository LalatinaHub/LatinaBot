package cron

import (
	"context"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/bot/handler"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"github.com/robfig/cron/v3"
	tele "gopkg.in/telebot.v3"
)

// Scheduler manages all background jobs.
type Scheduler struct {
	cron       *cron.Cron
	repos      *repository.Repositories
	cmdHandler *handler.CommandHandler
	bot        *tele.Bot
}

// Config defines dependencies for Scheduler.
type Config struct {
	Repos      *repository.Repositories
	CmdHandler *handler.CommandHandler
	Bot        *tele.Bot
}

// NewScheduler creates a new cron scheduler configured with Asia/Jakarta timezone.
func NewScheduler(cfg Config) *Scheduler {
	loc, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		loc = time.FixedZone("WIB", 7*3600)
	}

	c := cron.New(cron.WithLocation(loc))

	s := &Scheduler{
		cron:       c,
		repos:      cfg.Repos,
		cmdHandler: cfg.CmdHandler,
		bot:        cfg.Bot,
	}

	s.registerJobs()
	return s
}

func (s *Scheduler) registerJobs() {
	// 1. Maintenance Job every 5 minutes (Clean expired users, exceeded quota, assign tenants)
	_, _ = s.cron.AddFunc("*/5 * * * *", func() {
		logger.Info().Msg("Running scheduled maintenance: clean users & sync server tenants...")
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		if _, err := s.repos.User.CleanExpiredUsers(ctx, 7); err != nil {
			logger.Error().Err(err).Msg("Cron: failed to clean expired users")
		}

		if err := s.repos.User.CleanExceededQuota(ctx, 10); err != nil {
			logger.Error().Err(err).Msg("Cron: failed to clean exceeded quota")
		}

		if err := edgeserver.AssignServerTenants(ctx, s.repos.Server, s.repos.User); err != nil {
			logger.Error().Err(err).Msg("Cron: failed to assign server tenants")
		}
	})

	// 2. Broadcast Free Public Node every 3 hours
	_, _ = s.cron.AddFunc("0 */3 * * *", func() {
		if s.cmdHandler != nil && s.bot != nil {
			logger.Info().Msg("Running scheduled public proxy node broadcast...")
			s.cmdHandler.SendPublicNodes(s.bot)
		}
	})

	// 3. Forward Promotional Message every 8 hours
	_, _ = s.cron.AddFunc("0 */8 * * *", func() {
		if s.cmdHandler != nil && s.bot != nil {
			logger.Info().Msg("Running scheduled promotional message forwarding...")
			s.cmdHandler.SendPromotionalMessage(s.bot)
		}
	})
}

// Start begins background job execution.
func (s *Scheduler) Start() {
	logger.Info().Msg("Cron scheduler started")
	s.cron.Start()
}

// Stop terminates all cron jobs gracefully.
func (s *Scheduler) Stop() context.Context {
	logger.Info().Msg("Stopping cron scheduler...")
	return s.cron.Stop()
}
