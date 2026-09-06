package bot

import (
	"strconv"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaBot/config"
	"github.com/LalatinaHub/LatinaBot/internal/bot/conversation"
	"github.com/LalatinaHub/LatinaBot/internal/bot/handler"
	"github.com/LalatinaHub/LatinaBot/internal/bot/middleware"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/cloudflare"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/internal/service/ocr"
	"github.com/LalatinaHub/LatinaBot/internal/service/payment"
	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	tele "gopkg.in/telebot.v3"
)

// Bot manages the Telegram bot instance and handler wiring.
type Bot struct {
	teleBot        *tele.Bot
	cfg            *config.Config
	cmdHandler     *handler.CommandHandler
	cbHandler      *handler.CallbackHandler
	paymentHandler *handler.PaymentHandler
	photoHandler   *handler.PhotoHandler
	vpnWizard      *conversation.VPNWizard
	wildcardWizard *conversation.WildcardWizard
	fsmStore       *conversation.FSMStore
}

// Config holds dependencies for initializing the Bot.
type Config struct {
	AppConfig   *config.Config
	Repos       *repository.Repositories
	TarotSvc    *tarot.Service
	TrakteerSvc *trakteer.Client
	PaymentSvc  *payment.Service
	EdgeClient  *edgeserver.Client
	CFClient    *cloudflare.Client
	OCRScanner  ocr.Scanner
}

// NewBot initializes the Telebot engine with all handlers and middlewares.
func NewBot(cfg Config) (*Bot, error) {
	pref := tele.Settings{
		Token:  cfg.AppConfig.BotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := tele.NewBot(pref)
	if err != nil {
		return nil, err
	}

	fsmStore := conversation.NewFSMStore(10 * time.Minute)

	cmdHandler := handler.NewCommandHandler(
		cfg.AppConfig,
		cfg.Repos,
		cfg.TarotSvc,
		cfg.TrakteerSvc,
		cfg.EdgeClient,
	)

	cbHandler := handler.NewCallbackHandler(
		cfg.Repos,
		cfg.TarotSvc,
		cfg.TrakteerSvc,
		cfg.EdgeClient,
		fsmStore,
	)

	paymentHandler := handler.NewPaymentHandler(
		cfg.AppConfig,
		cfg.Repos,
		cfg.PaymentSvc,
		cfg.TarotSvc,
		cfg.TrakteerSvc,
		cfg.EdgeClient,
	)

	photoHandler := handler.NewPhotoHandler(
		cfg.AppConfig,
		cfg.Repos,
		cfg.OCRScanner,
		cmdHandler,
		cfg.TrakteerSvc,
		cfg.TarotSvc,
		cfg.EdgeClient,
	)

	vpnWizard := conversation.NewVPNWizard(
		cfg.Repos.Server,
		cfg.Repos.User,
		cfg.EdgeClient,
		fsmStore,
	)

	wildcardWizard := conversation.NewWildcardWizard(
		cfg.Repos.User,
		cfg.Repos.Server,
		cfg.Repos.Wildcard,
		cfg.CFClient,
		fsmStore,
	)

	botInstance := &Bot{
		teleBot:        b,
		cfg:            cfg.AppConfig,
		cmdHandler:     cmdHandler,
		cbHandler:      cbHandler,
		paymentHandler: paymentHandler,
		photoHandler:   photoHandler,
		vpnWizard:      vpnWizard,
		wildcardWizard: wildcardWizard,
		fsmStore:       fsmStore,
	}

	botInstance.registerRoutes()
	return botInstance, nil
}

func (b *Bot) registerRoutes() {
	// Middlewares
	b.teleBot.Use(middleware.Typing())
	b.teleBot.Use(middleware.ErrorReporter(b.teleBot, b.cfg.AdminID))

	// Commands
	b.teleBot.Handle("/start", b.cmdHandler.HandleStart, middleware.PrivateOnly())
	b.teleBot.Handle("/promote", func(c tele.Context) error {
		return b.cmdHandler.HandlePromote(c, b.teleBot)
	}, middleware.AdminOnly(b.cfg.AdminID))
	b.teleBot.Handle("/new_order", b.cmdHandler.HandleNewOrder, middleware.AdminOnly(b.cfg.AdminID))

	// Callbacks
	b.teleBot.Handle(&tele.Btn{Unique: "m/refresh"}, b.cbHandler.HandleRefresh)
	b.teleBot.Handle(&tele.Btn{Unique: "m/info"}, b.cbHandler.HandleServerInfo)
	b.teleBot.Handle(&tele.Btn{Unique: "c/pass"}, b.cbHandler.HandleChangePass)
	b.teleBot.Handle(&tele.Btn{Unique: "c/uuid"}, b.cbHandler.HandleChangeUUID)
	b.teleBot.Handle(&tele.Btn{Unique: "s/adblock"}, b.cbHandler.HandleToggleAdblock)
	b.teleBot.Handle(&tele.Btn{Unique: "l/wildcard"}, b.cbHandler.HandleListWildcard)
	b.teleBot.Handle(&tele.Btn{Unique: "c/wildcard"}, b.wildcardWizard.StartWizard)
	b.teleBot.Handle(&tele.Btn{Unique: "t/desclaimer"}, b.cbHandler.HandleDisclaimer)
	b.teleBot.Handle(&tele.Btn{Unique: "t/donasi"}, b.cbHandler.HandleDonasiInfo)

	// Payment callbacks
	b.teleBot.Handle(&tele.Btn{Unique: "m/donasi"}, b.paymentHandler.HandleMakePayment)

	// VPN Wizard
	b.teleBot.Handle(&tele.Btn{Unique: "c/vpn"}, b.vpnWizard.StartWizard)
	b.teleBot.Handle(&tele.Btn{Unique: "confirm"}, b.cbHandler.HandleConfirmVPN)
	b.teleBot.Handle(&tele.Btn{Unique: "cancel"}, b.cbHandler.HandleCancelVPN)

	// Dynamic Callbacks via OnCallback
	b.teleBot.Handle(tele.OnCallback, func(c tele.Context) error {
		data := c.Callback().Data

		if strings.HasPrefix(data, "vpn_proto_") {
			proto := strings.TrimPrefix(data, "vpn_proto_")
			return b.vpnWizard.HandleProtocolSelection(c, proto)
		}

		if strings.HasPrefix(data, "vpn_srv_") {
			srvCode := strings.TrimPrefix(data, "vpn_srv_")
			return b.vpnWizard.HandleServerSelection(c, srvCode)
		}

		if strings.HasPrefix(data, "vpn_page_") {
			pageStr := strings.TrimPrefix(data, "vpn_page_")
			page, _ := strconv.Atoi(pageStr)
			return b.vpnWizard.HandleRelayPagination(c, page)
		}

		if strings.HasPrefix(data, "vpn_relay_") {
			relay := strings.TrimPrefix(data, "vpn_relay_")
			return b.vpnWizard.HandleRelaySelection(c, relay)
		}

		if data == "vpn_end" {
			return c.Respond(&tele.CallbackResponse{Text: "Sudah di halaman akhir!"})
		}

		if strings.HasPrefix(data, "c/donasi_") {
			return b.paymentHandler.HandleCheckPayment(c)
		}

		return nil
	})

	// Photo Handler (Donation OCR)
	b.teleBot.Handle(tele.OnPhoto, func(c tele.Context) error {
		return b.photoHandler.HandlePhoto(c, b.teleBot)
	}, middleware.PrivateOnly())

	// Text Handler (Wildcard domain input)
	b.teleBot.Handle(tele.OnText, func(c tele.Context) error {
		sess := b.fsmStore.GetSession(c.Sender().ID)
		if sess.State == conversation.StateWaitingWildcardDomain {
			return b.wildcardWizard.HandleTextMessage(c)
		}
		return nil
	}, middleware.PrivateOnly())
}

// Start runs the Telegram bot poller.
func (b *Bot) Start() {
	logger.Info().Msg("Telegram bot polling started")
	b.teleBot.Start()
}

// Stop terminates the Telegram bot gracefully.
func (b *Bot) Stop() {
	logger.Info().Msg("Stopping Telegram bot...")
	b.teleBot.Stop()
}

// TeleBot returns the underlying telebot instance.
func (b *Bot) TeleBot() *tele.Bot {
	return b.teleBot
}

// CmdHandler returns the CommandHandler instance.
func (b *Bot) CmdHandler() *handler.CommandHandler {
	return b.cmdHandler
}
