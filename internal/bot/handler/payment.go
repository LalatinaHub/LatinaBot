package handler

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaBot/config"
	"github.com/LalatinaHub/LatinaBot/internal/bot/template"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/internal/service/payment"
	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/LatinaBot/pkg/imageutil"
	tele "gopkg.in/telebot.v3"
)

// PaymentHandler manages Xendit QR payments and verifications.
type PaymentHandler struct {
	cfg         *config.Config
	repos       *repository.Repositories
	paymentSvc  *payment.Service
	tarotSvc    *tarot.Service
	trakteerSvc *trakteer.Client
	edgeClient  *edgeserver.Client
}

// NewPaymentHandler creates a new PaymentHandler.
func NewPaymentHandler(
	cfg *config.Config,
	repos *repository.Repositories,
	paymentSvc *payment.Service,
	tarotSvc *tarot.Service,
	trakteerSvc *trakteer.Client,
	edgeClient *edgeserver.Client,
) *PaymentHandler {
	return &PaymentHandler{
		cfg:         cfg,
		repos:       repos,
		paymentSvc:  paymentSvc,
		tarotSvc:    tarotSvc,
		trakteerSvc: trakteerSvc,
		edgeClient:  edgeClient,
	}
}

// HandleMakePayment creates a dynamic QR code for 5000 IDR.
func (h *PaymentHandler) HandleMakePayment(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	res := h.paymentSvc.MakePayment(ctx, 5000)
	if res.Error {
		_ = c.Respond()
		return c.Reply(fmt.Sprintf("Gagal membuat pembayaran!\n%s", res.Message), tele.ModeHTML)
	}

	qrBytes, err := imageutil.GenerateQRCodePNG(res.QRString, 500)
	if err != nil {
		_ = c.Respond()
		return c.Reply("Gagal generate QR Code: " + err.Error())
	}

	menu := &tele.ReplyMarkup{}
	btnRefresh := menu.Data("Refresh", fmt.Sprintf("c/donasi_%s", res.Message))
	menu.Inline(menu.Row(btnRefresh))

	photo := &tele.Photo{
		File:    tele.FromReader(imageutil.BytesReader(qrBytes)),
		Caption: "Lakukan pembayaran pada qrcode di atas!",
	}

	_ = c.Respond()
	return c.Reply(photo, menu)
}

// HandleCheckPayment checks status of QR payment and grants quota.
func (h *PaymentHandler) HandleCheckPayment(c tele.Context) error {
	callbackData := c.Callback().Data
	token := strings.TrimPrefix(callbackData, "c/donasi_")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	exists, err := h.repos.Donation.Exists(ctx, token)
	if err != nil || exists {
		return c.Respond(&tele.CallbackResponse{Text: "Token Expired!", ShowAlert: true})
	}

	checkRes := h.paymentSvc.CheckPayment(ctx, token)
	if checkRes.Error {
		return c.Respond(&tele.CallbackResponse{Text: checkRes.Message, ShowAlert: true})
	}

	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun tidak ditemukan", ShowAlert: true})
	}

	now := time.Now()
	expired := user.Expired
	if now.After(expired) {
		expired = now
	}
	user.Expired = expired.AddDate(0, 0, 30)

	if user.Quota > 0 {
		user.Quota += h.cfg.QuotaPerDonation
	} else {
		user.Quota = h.cfg.QuotaPerDonation
	}

	if err := h.repos.User.UpdateUser(ctx, user); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Gagal update akun pengguna", ShowAlert: true})
	}

	_ = h.repos.Donation.Create(ctx, token)

	// Async reload
	go func() {
		rCtx, rCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer rCancel()
		apiToken, _ := h.repos.KV.Get(rCtx, "apiToken")
		servers, _ := h.repos.Server.GetAll(rCtx)
		h.edgeClient.ReloadServers(rCtx, servers, apiToken)
	}()

	_ = c.Delete()

	card := h.tarotSvc.GetCard()
	donations, _ := h.trakteerSvc.GetDonations(ctx)
	servers, _ := h.repos.Server.GetAll(ctx)

	msgText := template.BuildStartMessage(c.Sender(), user, card, donations, servers)
	startMenu := template.BuildStartKeyboard(user)

	photo := &tele.Photo{
		File:    tele.FromURL(card.Image),
		Caption: msgText,
	}

	return c.Send(photo, startMenu, tele.ModeHTML)
}
