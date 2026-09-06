package handler

import (
	"context"
	"crypto/rand"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/bot/conversation"
	"github.com/LalatinaHub/LatinaBot/internal/bot/template"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/google/uuid"
	tele "gopkg.in/telebot.v3"
)

var domainRegex = regexp.MustCompile(`(?m)Domain:\s+([^\s]+)`)

// CallbackHandler handles inline button interactions.
type CallbackHandler struct {
	repos       *repository.Repositories
	tarotSvc    *tarot.Service
	trakteerSvc *trakteer.Client
	edgeClient  *edgeserver.Client
	fsm         *conversation.FSMStore
}

// NewCallbackHandler creates a new CallbackHandler.
func NewCallbackHandler(
	repos *repository.Repositories,
	tarotSvc *tarot.Service,
	trakteerSvc *trakteer.Client,
	edgeClient *edgeserver.Client,
	fsm *conversation.FSMStore,
) *CallbackHandler {
	return &CallbackHandler{
		repos:       repos,
		tarotSvc:    tarotSvc,
		trakteerSvc: trakteerSvc,
		edgeClient:  edgeClient,
		fsm:         fsm,
	}
}

// HandleRefresh re-renders the start menu.
func (h *CallbackHandler) HandleRefresh(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun belum terdaftar", ShowAlert: true})
	}

	card := h.tarotSvc.GetCard()
	donations, _ := h.trakteerSvc.GetDonations(ctx)
	servers, _ := h.repos.Server.GetAll(ctx)

	msgText := template.BuildStartMessage(c.Sender(), user, card, donations, servers)
	menu := template.BuildStartKeyboard(user)

	_ = c.Respond()
	return c.EditCaption(msgText, menu, tele.ModeHTML)
}

// HandleServerInfo displays an alert modal with live hardware metrics of the edge server.
func (h *CallbackHandler) HandleServerInfo(c tele.Context) error {
	var domainName string
	if c.Message() != nil {
		matches := domainRegex.FindStringSubmatch(c.Message().Caption)
		if len(matches) > 1 {
			domainName = matches[1]
		}
	}

	if domainName == "" {
		return c.Respond(&tele.CallbackResponse{
			Text:      "Domain server belum dipilih / tidak ditemukan!",
			ShowAlert: true,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status, err := h.edgeClient.GetStatus(ctx, domainName)
	if err != nil || status == nil {
		return c.Respond(&tele.CallbackResponse{
			Text:      "Gagal mengambil status server: server offline / tidak merespon",
			ShowAlert: true,
		})
	}

	cpuPercent := 0.0
	if len(status.CPU) > 0 {
		cpuPercent = status.CPU[0]
	}

	var bytesSent, bytesRecv uint64
	if len(status.NIC) > 0 {
		bytesSent = status.NIC[0].BytesSent
		bytesRecv = status.NIC[0].BytesRecv
	}

	uptimeDuration := time.Duration(status.Host.Uptime) * time.Second

	msg := fmt.Sprintf(`Server Status

Host
OS: %s %s
Disk: %.2f%%
Uptime: %s

Load
CPU: %.2f%%
RAM: %.2f%%

Network
UP: %s
DL: %s`,
		status.Host.Platform, status.Host.PlatformVersion,
		status.Disk.UsedPercent,
		formatDuration(uptimeDuration),
		cpuPercent,
		status.Mem.UsedPercent,
		formatBytes(bytesSent),
		formatBytes(bytesRecv),
	)

	return c.Respond(&tele.CallbackResponse{
		Text:      msg,
		ShowAlert: true,
	})
}

// HandleChangePass generates a new 8-character password/token.
func (h *CallbackHandler) HandleChangePass(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun tidak ditemukan", ShowAlert: true})
	}

	user.Token = generateToken(8)
	if err := h.repos.User.UpdateUser(ctx, user); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Gagal memperbarui password", ShowAlert: true})
	}

	return h.HandleRefresh(c)
}

// HandleChangeUUID generates a new UUID password and triggers server reload.
func (h *CallbackHandler) HandleChangeUUID(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun tidak ditemukan", ShowAlert: true})
	}

	user.Password = uuid.New().String()
	if err := h.repos.User.UpdateUser(ctx, user); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Gagal memperbarui UUID", ShowAlert: true})
	}

	h.reloadServersAsync()
	return h.HandleRefresh(c)
}

// HandleToggleAdblock toggles the adblock setting and triggers server reload.
func (h *CallbackHandler) HandleToggleAdblock(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun tidak ditemukan", ShowAlert: true})
	}

	user.Adblock = !user.Adblock
	if err := h.repos.User.UpdateUser(ctx, user); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Gagal memperbarui adblock", ShowAlert: true})
	}

	h.reloadServersAsync()
	return h.HandleRefresh(c)
}

// HandleListWildcard displays all registered wildcard domains.
func (h *CallbackHandler) HandleListWildcard(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	servers, _ := h.repos.Server.GetAll(ctx)
	wildcards, _ := h.repos.Wildcard.GetAll(ctx)

	var b strings.Builder
	b.WriteString("Daftar Wildcard:\n")
	b.WriteString("<blockquote expandable>\n")
	for _, w := range wildcards {
		b.WriteString(fmt.Sprintf("• <code>%s</code>\n", w.Domain))
	}
	b.WriteString("</blockquote>\n\n")

	if len(wildcards) > 0 && len(servers) > 0 {
		b.WriteString("Contoh:\n")
		b.WriteString(fmt.Sprintf("<code>%s.%s</code>", wildcards[0].Domain, servers[0].Domain))
	}

	menu := &tele.ReplyMarkup{}
	btnRegister := menu.Data("Tambah Wildcard", "c_wildcard")
	menu.Inline(menu.Row(btnRegister))

	_ = c.Respond()
	return c.Reply(b.String(), menu, tele.ModeHTML)
}

// HandleDisclaimer shows the disclaimer alert modal.
func (h *CallbackHandler) HandleDisclaimer(c tele.Context) error {
	msg := "Semua akun yang disediakan oleh API, merupakan akun yang tersedia secara bebas di Internet.\n" +
		"Layanan ini tidak melakukan aktivitas tracing, logging, atau semacamnya!."

	return c.Respond(&tele.CallbackResponse{
		Text:      msg,
		ShowAlert: true,
	})
}

// HandleDonasiInfo sends donation tutorial with image.
func (h *CallbackHandler) HandleDonasiInfo(c tele.Context) error {
	msg := "<b>Cara Melakukan Donasi</b>\n" +
		"1. Buka halaman donasi\n" +
		"2. Selesaikan donasi\n" +
		"3. Simpan bukti donasi\n" +
		"4. Kirimkan bukti donasi ke bot\n" +
		"5. Crop fotonya biar lebih jelas (optional)\n\n\n" +
		"↓ Tombol Donasi"

	photo := &tele.Photo{
		File:    tele.FromURL("https://okzpqehvbvtzrjzohjtw.supabase.co/storage/v1/object/public/assets/donasi_1.png"),
		Caption: "↑ Contoh bukti donasi ↑\n\n" + msg,
	}

	menu := &tele.ReplyMarkup{}
	btnDonasiQRIS := menu.Data("Donasi QRIS", "m_donasi")
	menu.Inline(menu.Row(btnDonasiQRIS))

	_ = c.Respond()
	return c.Send(photo, menu, tele.ModeHTML)
}

// HandleConfirmVPN saves the configured VPN settings to user profile.
func (h *CallbackHandler) HandleConfirmVPN(c tele.Context) error {
	sess := h.fsm.GetSession(c.Sender().ID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := h.repos.User.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun tidak ditemukan", ShowAlert: true})
	}

	user.ServerCode = sess.VPNData.ServerCode
	user.Relay = sess.VPNData.Relay
	user.VPN = sess.VPNData.VPN

	if err := h.repos.User.UpdateUser(ctx, user); err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Gagal menyimpan konfigurasi", ShowAlert: true})
	}

	h.fsm.Reset(c.Sender().ID)
	h.reloadServersAsync()

	return h.HandleRefresh(c)
}

// HandleCancelVPN cancels the current VPN wizard and resets state.
func (h *CallbackHandler) HandleCancelVPN(c tele.Context) error {
	h.fsm.Reset(c.Sender().ID)
	return h.HandleRefresh(c)
}

func (h *CallbackHandler) reloadServersAsync() {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		apiToken, err := h.repos.KV.Get(ctx, "apiToken")
		if err != nil || apiToken == "" {
			return
		}

		servers, err := h.repos.Server.GetAll(ctx)
		if err != nil {
			return
		}

		time.Sleep(1 * time.Second)
		h.edgeClient.ReloadServers(ctx, servers, apiToken)
	}()
}

func generateToken(n int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	randBytes := make([]byte, n)
	_, _ = rand.Read(randBytes)
	for i := range b {
		b[i] = charset[int(randBytes[i])%len(charset)]
	}
	return string(b)
}

func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	}
	return fmt.Sprintf("%dm", minutes)
}
