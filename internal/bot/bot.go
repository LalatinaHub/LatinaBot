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
		Token:   cfg.AppConfig.BotToken,
		Poller:  &tele.LongPoller{Timeout: 10 * time.Second},
		Verbose: cfg.AppConfig.IsDevelopment(),
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
	b.teleBot.Use(middleware.Logger(b.cfg.IsDevelopment()))
	b.teleBot.Use(middleware.Typing())
	b.teleBot.Use(middleware.ErrorReporter(b.teleBot, b.cfg.AdminID))

	// Commands
	b.teleBot.Handle("/start", b.cmdHandler.HandleStart, middleware.PrivateOnly())
	b.teleBot.Handle("/promote", func(c tele.Context) error {
		return b.cmdHandler.HandlePromote(c, b.teleBot)
	}, middleware.AdminOnly(b.cfg.AdminID))
	b.teleBot.Handle("/new_order", b.cmdHandler.HandleNewOrder, middleware.AdminOnly(b.cfg.AdminID))

	// Direct button callback handlers (valid telebot identifiers matching [-\w]+)
	b.teleBot.Handle(&tele.Btn{Unique: "m_refresh"}, b.cbHandler.HandleRefresh)
	b.teleBot.Handle(&tele.Btn{Unique: "m_info"}, b.cbHandler.HandleServerInfo)
	b.teleBot.Handle(&tele.Btn{Unique: "c_pass"}, b.cbHandler.HandleChangePass)
	b.teleBot.Handle(&tele.Btn{Unique: "c_uuid"}, b.cbHandler.HandleChangeUUID)
	b.teleBot.Handle(&tele.Btn{Unique: "s_adblock"}, b.cbHandler.HandleToggleAdblock)
	b.teleBot.Handle(&tele.Btn{Unique: "l_wildcard"}, b.cbHandler.HandleListWildcard)
	b.teleBot.Handle(&tele.Btn{Unique: "c_wildcard"}, b.wildcardWizard.StartWizard)
	b.teleBot.Handle(&tele.Btn{Unique: "t_desclaimer"}, b.cbHandler.HandleDisclaimer)
	b.teleBot.Handle(&tele.Btn{Unique: "t_donasi"}, b.cbHandler.HandleDonasiInfo)
	b.teleBot.Handle(&tele.Btn{Unique: "m_donasi"}, b.paymentHandler.HandleMakePayment)
	b.teleBot.Handle(&tele.Btn{Unique: "c_vpn"}, b.vpnWizard.StartWizard)
	b.teleBot.Handle(&tele.Btn{Unique: "confirm"}, b.cbHandler.HandleConfirmVPN)
	b.teleBot.Handle(&tele.Btn{Unique: "cancel"}, b.cbHandler.HandleCancelVPN)

	// Comprehensive fallback & dynamic callback dispatcher
	b.teleBot.Handle(tele.OnCallback, func(c tele.Context) error {
		cb := c.Callback()
		if cb == nil {
			return nil
		}

		// Normalize callback data: strip Telebot prefix \f and extract base action
		raw := strings.TrimPrefix(cb.Data, "\f")
		action := raw
		if idx := strings.Index(raw, "|"); idx != -1 {
			action = raw[:idx]
		}

		switch {
		// Start menu callbacks (support both new _ and legacy / formats)
		case action == "c_vpn" || action == "c/vpn":
			return b.vpnWizard.StartWizard(c)
		case action == "m_refresh" || action == "m/refresh":
			return b.cbHandler.HandleRefresh(c)
		case action == "m_info" || action == "m/info":
			return b.cbHandler.HandleServerInfo(c)
		case action == "c_pass" || action == "c/pass":
			return b.cbHandler.HandleChangePass(c)
		case action == "c_uuid" || action == "c/uuid":
			return b.cbHandler.HandleChangeUUID(c)
		case action == "s_adblock" || action == "s/adblock":
			return b.cbHandler.HandleToggleAdblock(c)
		case action == "l_wildcard" || action == "l/wildcard":
			return b.cbHandler.HandleListWildcard(c)
		case action == "c_wildcard" || action == "c/wildcard":
			return b.wildcardWizard.StartWizard(c)
		case action == "t_desclaimer" || action == "t/desclaimer":
			return b.cbHandler.HandleDisclaimer(c)
		case action == "t_donasi" || action == "t/donasi":
			return b.cbHandler.HandleDonasiInfo(c)
		case action == "m_donasi" || action == "m/donasi":
			return b.paymentHandler.HandleMakePayment(c)
		case action == "confirm":
			return b.cbHandler.HandleConfirmVPN(c)
		case action == "cancel":
			return b.cbHandler.HandleCancelVPN(c)

		// Dynamic VPN Wizard Callbacks
		case strings.HasPrefix(action, "vpn_proto_"):
			proto := strings.TrimPrefix(action, "vpn_proto_")
			return b.vpnWizard.HandleProtocolSelection(c, proto)
		case strings.HasPrefix(action, "vpn_srv_"):
			srvCode := strings.TrimPrefix(action, "vpn_srv_")
			return b.vpnWizard.HandleServerSelection(c, srvCode)
		case strings.HasPrefix(action, "vpn_page_"):
			pageStr := strings.TrimPrefix(action, "vpn_page_")
			page, _ := strconv.Atoi(pageStr)
			return b.vpnWizard.HandleRelayPagination(c, page)
		case strings.HasPrefix(action, "vpn_relay_"):
			relay := strings.TrimPrefix(action, "vpn_relay_")
			return b.vpnWizard.HandleRelaySelection(c, relay)
		case action == "vpn_end":
			return c.Respond(&tele.CallbackResponse{Text: "Sudah di halaman akhir!"})

		// Dynamic Payment Check Callbacks
		case strings.HasPrefix(action, "c_donasi_") || strings.HasPrefix(action, "c/donasi_"):
			return b.paymentHandler.HandleCheckPayment(c)

		default:
			logger.Warn().Str("callback_data", cb.Data).Str("action", action).Msg("Unhandled callback query")
			return c.Respond()
		}
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
