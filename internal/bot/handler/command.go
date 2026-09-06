package handler

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/LalatinaHub/LatinaBot/config"
	"github.com/LalatinaHub/LatinaBot/internal/bot/template"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/LatinaBot/pkg/imageutil"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"github.com/LalatinaHub/LatinaBot/pkg/proxyutil"
	"github.com/LalatinaHub/LatinaBot/pkg/stringutil"
	"github.com/google/uuid"
	tele "gopkg.in/telebot.v3"
)

// CommandHandler handles bot commands (/start, /promote, /new_order).
type CommandHandler struct {
	cfg          *config.Config
	repos        *repository.Repositories
	tarotSvc     *tarot.Service
	trakteerSvc  *trakteer.Client
	edgeClient   *edgeserver.Client
	localOrderID string
	orderMu      sync.RWMutex
}

// NewCommandHandler creates a new CommandHandler.
func NewCommandHandler(
	cfg *config.Config,
	repos *repository.Repositories,
	tarotSvc *tarot.Service,
	trakteerSvc *trakteer.Client,
	edgeClient *edgeserver.Client,
) *CommandHandler {
	return &CommandHandler{
		cfg:         cfg,
		repos:       repos,
		tarotSvc:    tarotSvc,
		trakteerSvc: trakteerSvc,
		edgeClient:  edgeClient,
	}
}

// GetLocalOrderID returns current local order ID.
func (h *CommandHandler) GetLocalOrderID() string {
	h.orderMu.RLock()
	defer h.orderMu.RUnlock()
	return h.localOrderID
}

// HandleStart handles the /start command.
func (h *CommandHandler) HandleStart(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Get or auto-register user
	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil {
		logger.Error().Err(err).Int64("user_id", c.Sender().ID).Msg("Failed to query user")
	}

	if user == nil {
		user, err = h.repos.User.CreateUser(ctx, c.Sender().ID)
		if err != nil {
			logger.Error().Err(err).Int64("user_id", c.Sender().ID).Msg("Failed to create user")
			return c.Reply("Gagal membuat akun profil, coba lagi nanti ya.")
		}
	}

	card := h.tarotSvc.GetCard()
	donations, _ := h.trakteerSvc.GetDonations(ctx)
	servers, _ := h.repos.Server.GetAll(ctx)

	msgText := template.BuildStartMessage(c.Sender(), user, card, donations, servers)
	menu := template.BuildStartKeyboard(user)

	photo := &tele.Photo{
		File:    tele.FromURL(card.Image),
		Caption: msgText,
	}

	return c.Send(photo, menu, tele.ModeHTML)
}

// HandlePromote forwards the promo message and free public nodes to group topic.
func (h *CommandHandler) HandlePromote(c tele.Context, bot *tele.Bot) error {
	if c.Sender().ID != h.cfg.AdminID {
		return nil
	}

	h.SendPromotionalMessage(bot)
	h.SendPublicNodes(bot)
	return c.Reply("Promote executed!")
}

// HandleNewOrder generates a local UUID and renders an order image.
func (h *CommandHandler) HandleNewOrder(c tele.Context) error {
	if c.Sender().ID != h.cfg.AdminID {
		return nil
	}

	newID := uuid.New().String()
	h.orderMu.Lock()
	h.localOrderID = newID
	h.orderMu.Unlock()

	imgBytes, err := imageutil.GenerateLocalOrderImage(newID)
	if err != nil {
		return c.Reply("Gagal generate gambar order ID: " + err.Error())
	}

	photo := &tele.Photo{
		File:    tele.FromReader(imageutil.BytesReader(imgBytes)),
		Caption: fmt.Sprintf("Local Order ID: <code>%s</code>", newID),
	}

	return c.Send(photo, tele.ModeHTML)
}

// SendPublicNodes sends a random free public proxy node to group topic.
func (h *CommandHandler) SendPublicNodes(bot *tele.Bot) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	proxy, err := h.repos.Proxy.GetRandomCDNProxy(ctx)
	if err != nil || proxy == nil {
		logger.Warn().Err(err).Msg("No public proxy available for broadcast")
		return
	}

	text := fmt.Sprintf(`🌕🌖🌗🌘 <b>FREE PUBLIC PROXY</b> 🌒🌓🌔🌕

<code>ID       : </code><code>%d</code>
<code>Server   : </code><code>%s</code>
<code>Port     : </code><code>%d</code>
<code>UUID     : </code><code>%s</code>
<code>Password : </code><code>%s</code>
<code>TLS      : </code><code>%t</code>
<code>Transport: </code><code>%s</code>
<code>Host     : </code><code>%s</code>
<code>Path     : </code><code>%s</code>
<code>Insecure : </code><code>%t</code>
<code>SNI      : </code><code>%s</code>
<code>Mode     : </code><code>%s</code>
<code>Country  : </code><code>%s | %s</code>
<code>Region   : </code><code>%s</code>
<code>ORG      : </code><code>%s</code>
<code>VPN      : </code><code>%s</code>

<code>====================================</code>
<code>%s</code>
<code>====================================</code>`,
		proxy.ID,
		proxy.Server,
		proxy.ServerPort,
		proxy.UUID,
		proxy.Password,
		proxy.TLS,
		proxy.Transport,
		proxy.Host,
		proxy.Path,
		proxy.Insecure,
		proxy.SNI,
		proxy.ConnMode,
		stringutil.CountryISOToEmoji(proxy.CountryCode), proxy.CountryCode,
		proxy.Region,
		proxy.Org,
		proxy.VPN,
		proxyutil.ConvertProxyToURL(proxy),
	)

	targetChat := &tele.Chat{ID: h.cfg.GroupID}
	opts := &tele.SendOptions{
		ParseMode:       tele.ModeHTML,
		ThreadID:        h.cfg.PublicNodeThreadID,
		DisableWebPagePreview: true,
	}

	_, err = bot.Send(targetChat, text, opts)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to send public proxy node broadcast")
	}
}

// SendPromotionalMessage forwards the promotion message to thread.
func (h *CommandHandler) SendPromotionalMessage(bot *tele.Bot) {
	if h.cfg.GroupID == 0 || h.cfg.PromotionMessageID == 0 {
		return
	}

	fromChat := &tele.Chat{ID: h.cfg.GroupID}
	targetChat := &tele.Chat{ID: h.cfg.GroupID}

	msg := &tele.Message{
		ID:   h.cfg.PromotionMessageID,
		Chat: fromChat,
	}

	opts := &tele.SendOptions{
		ThreadID: h.cfg.PromotionThreadID,
	}

	_, err := bot.Forward(targetChat, msg, opts)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to forward promotional message")
	}
}
