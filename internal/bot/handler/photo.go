package handler

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaBot/config"
	"github.com/LalatinaHub/LatinaBot/internal/bot/template"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/internal/service/ocr"
	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	tele "gopkg.in/telebot.v3"
)

// PhotoHandler processes incoming photo messages (donation receipts).
type PhotoHandler struct {
	cfg         *config.Config
	repos       *repository.Repositories
	ocrScanner  ocr.Scanner
	cmdHandler  *CommandHandler
	trakteerSvc *trakteer.Client
	tarotSvc    *tarot.Service
	edgeClient  *edgeserver.Client
}

// NewPhotoHandler creates a new PhotoHandler.
func NewPhotoHandler(
	cfg *config.Config,
	repos *repository.Repositories,
	ocrScanner ocr.Scanner,
	cmdHandler *CommandHandler,
	trakteerSvc *trakteer.Client,
	tarotSvc *tarot.Service,
	edgeClient *edgeserver.Client,
) *PhotoHandler {
	return &PhotoHandler{
		cfg:         cfg,
		repos:       repos,
		ocrScanner:  ocrScanner,
		cmdHandler:  cmdHandler,
		trakteerSvc: trakteerSvc,
		tarotSvc:    tarotSvc,
		edgeClient:  edgeClient,
	}
}

// HandlePhoto processes an uploaded photo to verify donation receipts via OCR.
func (h *PhotoHandler) HandlePhoto(c tele.Context, bot *tele.Bot) error {
	photo := c.Message().Photo
	if photo == nil {
		return nil
	}

	reader, err := bot.File(&photo.File)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to download telegram photo")
		return c.Reply("Gagal mengunduh foto!")
	}
	defer reader.Close()

	imgBytes, err := io.ReadAll(reader)
	if err != nil {
		logger.Error().Err(err).Msg("Failed to read photo bytes")
		return c.Reply("Gagal memproses file foto!")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	ocrText, err := h.ocrScanner.ScanImage(ctx, imgBytes)
	if err != nil {
		logger.Error().Err(err).Msg("OCR scanning failed")
		return c.Reply("Gagal membaca Order ID!\nCoba crop fotonya biar lebih jelas.")
	}

	orderID := h.ocrScanner.ExtractOrderID(ocrText)
	if orderID == "" {
		return c.Reply("Gagal membaca Order ID!\nCoba crop fotonya biar lebih jelas.")
	}

	// 1. Check if already claimed
	claimed, err := h.repos.Donation.Exists(ctx, orderID)
	if err == nil && claimed {
		return c.Reply("Order ID Expired!")
	}

	// 2. Validate against local order ID or Trakteer
	isVerified := false
	if h.cmdHandler != nil && strings.EqualFold(h.cmdHandler.GetLocalOrderID(), orderID) {
		isVerified = true
	} else {
		donations, err := h.trakteerSvc.GetDonations(ctx)
		if err == nil && donations != nil {
			for _, d := range donations.Result.Data {
				if strings.EqualFold(d.OrderID, orderID) {
					isVerified = true
					break
				}
			}
		}
	}

	if !isVerified {
		return c.Reply("Order ID Invalid!")
	}

	// 3. Update User Quota & Expiry (+30 days, +1,000,000 MB)
	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		user, _ = h.repos.User.CreateUser(ctx, c.Sender().ID)
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
		return c.Reply("Gagal memperbarui kuota akun!")
	}

	_ = h.repos.Donation.Create(ctx, orderID)

	// Reload servers
	go func() {
		rCtx, rCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer rCancel()
		apiToken, _ := h.repos.KV.Get(rCtx, "apiToken")
		servers, _ := h.repos.Server.GetAll(rCtx)
		h.edgeClient.ReloadServers(rCtx, servers, apiToken)
	}()

	_ = c.Reply(fmt.Sprintf("Donasi terverifikasi! Kuota +%d MB, Masa aktif +30 hari.", h.cfg.QuotaPerDonation))

	card := h.tarotSvc.GetCard()
	donations, _ := h.trakteerSvc.GetDonations(ctx)
	servers, _ := h.repos.Server.GetAll(ctx)

	msgText := template.BuildStartMessage(c.Sender(), user, card, donations, servers)
	menu := template.BuildStartKeyboard(user)

	startPhoto := &tele.Photo{
		File:    tele.FromURL(card.Image),
		Caption: msgText,
	}

	return c.Send(startPhoto, menu, tele.ModeHTML)
}
