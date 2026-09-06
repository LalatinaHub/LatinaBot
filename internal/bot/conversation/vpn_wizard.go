package conversation

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/domain"
	"github.com/LalatinaHub/LatinaBot/internal/repository"
	"github.com/LalatinaHub/LatinaBot/internal/service/edgeserver"
	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	"github.com/LalatinaHub/LatinaBot/pkg/stringutil"
	"github.com/LalatinaHub/common/model"
	tele "gopkg.in/telebot.v3"
)

// VPNWizard handles the multi-step VPN creation flow.
type VPNWizard struct {
	serverRepo repository.ServerRepository
	userRepo   repository.UserRepository
	edgeClient *edgeserver.Client
	store      *FSMStore
}

// NewVPNWizard creates a new VPN wizard handler.
func NewVPNWizard(serverRepo repository.ServerRepository, userRepo repository.UserRepository, edgeClient *edgeserver.Client, store *FSMStore) *VPNWizard {
	return &VPNWizard{
		serverRepo: serverRepo,
		userRepo:   userRepo,
		edgeClient: edgeClient,
		store:      store,
	}
}

// StartWizard initializes the VPN creation dialog.
func (w *VPNWizard) StartWizard(c tele.Context) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	user, err := w.userRepo.GetUser(ctx, c.Sender().ID)
	if err != nil || user == nil {
		return c.Respond(&tele.CallbackResponse{Text: "Akun kamu belum terdaftar!", ShowAlert: true})
	}

	if user.Quota <= 10 {
		return c.Respond(&tele.CallbackResponse{
			Text:      "Kuota kamu habis kak!",
			ShowAlert: true,
		})
	}

	sess := w.store.GetSession(c.Sender().ID)
	sess.State = StateWaitingVPNProtocol
	sess.VPNData = VPNSession{}

	msg := "Pilih protokol sesuai kebutuhan:\n" +
		"• VMess: Cepat dan stabil, cocok buat streaming atau browsing.\n" +
		"• VLESS: Hemat data, ideal untuk jaringan kurang stabil.\n" +
		"• Trojan: Super aman, cocok untuk kamu yang peduli privasi.\n\n" +
		"Bebas ganti protokol kapan aja!"

	menu := &tele.ReplyMarkup{}
	btnVMess := menu.Data("VMess", "vpn_proto_vmess")
	btnVLESS := menu.Data("VLESS", "vpn_proto_vless")
	btnTrojan := menu.Data("Trojan", "vpn_proto_trojan")
	menu.Inline(menu.Row(btnVMess, btnVLESS, btnTrojan))

	_ = c.Respond()
	return c.EditCaption(msg, menu, tele.ModeHTML)
}

// HandleProtocolSelection handles when the user picks vmess/vless/trojan.
func (w *VPNWizard) HandleProtocolSelection(c tele.Context, proto string) error {
	sess := w.store.GetSession(c.Sender().ID)
	sess.VPNData.VPN = proto
	sess.State = StateWaitingVPNServer

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	servers, err := w.serverRepo.GetAll(ctx)
	if err != nil || len(servers) == 0 {
		return c.Respond(&tele.CallbackResponse{Text: "Tidak ada server yang tersedia saat ini.", ShowAlert: true})
	}

	// Fetch ping & org concurrently from each edge server
	type serverMetric struct {
		srv  model.Server
		info *domain.EdgeServerInfo
	}

	metrics := make([]serverMetric, len(servers))
	var wg sync.WaitGroup
	for i, s := range servers {
		wg.Add(1)
		go func(idx int, srv model.Server) {
			defer wg.Done()
			info, _ := w.edgeClient.GetInfo(ctx, srv.Domain)
			metrics[idx] = serverMetric{srv: srv, info: info}
		}(i, s)
	}
	wg.Wait()

	var b strings.Builder
	b.WriteString("Pilih server kesukaanmu!\n")
	b.WriteString("• Sorry kalo masih dikit, kalo banyak yang donasi pasti makin banyak nantinya 😉\n")
	b.WriteString("• Abaikan jika ping agak besar, karena lokasi bot ada di eropa\n\n")
	b.WriteString("<blockquote expandable>\n")
	b.WriteString("<b>Server's Stats</b>\n\n")

	menu := &tele.ReplyMarkup{}
	var serverButtons []tele.Btn

	for _, m := range metrics {
		b.WriteString(fmt.Sprintf("• <b>%s</b> [%d/%d]\n", m.srv.Code, m.srv.UsersCount, m.srv.UsersMax))
		if m.info != nil {
			b.WriteString(fmt.Sprintf("•• 📡 %d ms\n", m.info.Ping))
			b.WriteString(fmt.Sprintf("•• 🪪 %s\n", m.info.Org))
		}
		if m.srv.UsersCount >= m.srv.UsersMax {
			b.WriteString("•• 😭 Penuh\n\n")
		} else {
			b.WriteString("•• 😁 Boleh lah\n\n")
			btn := menu.Data(m.srv.Code, fmt.Sprintf("vpn_srv_%s", m.srv.Code))
			serverButtons = append(serverButtons, btn)
		}
	}
	b.WriteString("</blockquote>")

	var rows []tele.Row
	// Put up to 3 buttons per row
	for i := 0; i < len(serverButtons); i += 3 {
		end := i + 3
		if end > len(serverButtons) {
			end = len(serverButtons)
		}
		rows = append(rows, menu.Row(serverButtons[i:end]...))
	}
	menu.Inline(rows...)

	_ = c.Respond()
	return c.EditCaption(b.String(), menu, tele.ModeHTML)
}

// HandleServerSelection handles when the user picks a server code.
func (w *VPNWizard) HandleServerSelection(c tele.Context, serverCode string) error {
	sess := w.store.GetSession(c.Sender().ID)
	sess.VPNData.ServerCode = serverCode
	sess.State = StateWaitingVPNRelay
	sess.VPNData.Page = 0

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	servers, err := w.serverRepo.GetAll(ctx)
	if err != nil {
		return c.Respond(&tele.CallbackResponse{Text: "Gagal mengambil daftar server", ShowAlert: true})
	}

	var targetDomain string
	for _, s := range servers {
		if s.Code == serverCode {
			targetDomain = s.Domain
			break
		}
	}

	if targetDomain == "" {
		return c.Respond(&tele.CallbackResponse{Text: "Server tidak ditemukan", ShowAlert: true})
	}

	relays, err := w.edgeClient.GetRelays(ctx, targetDomain)
	if err != nil {
		logger.Warn().Err(err).Msg("Failed to fetch relays from server")
	}

	// Unique country codes
	ccMap := make(map[string]bool)
	for _, r := range relays {
		if r.CountryCode != "" {
			ccMap[strings.ToUpper(r.CountryCode)] = true
		}
	}

	var ccList []string
	for cc := range ccMap {
		ccList = append(ccList, cc)
	}
	sort.Strings(ccList)

	// Prepend "Tanpa Relay"
	sess.VPNData.RelaysCC = append([]string{"Tanpa Relay"}, ccList...)

	return w.renderRelayMenu(c, sess, relays)
}

// HandleRelayPagination handles pagination clicks in the relay selector.
func (w *VPNWizard) HandleRelayPagination(c tele.Context, page int) error {
	sess := w.store.GetSession(c.Sender().ID)
	sess.VPNData.Page = page
	return w.renderRelayMenu(c, sess, nil)
}

// HandleRelaySelection handles picking a relay or "Tanpa Relay".
func (w *VPNWizard) HandleRelaySelection(c tele.Context, relay string) error {
	sess := w.store.GetSession(c.Sender().ID)
	if relay == "Tanpa Relay" {
		sess.VPNData.Relay = ""
	} else {
		sess.VPNData.Relay = relay
	}
	sess.State = StateWaitingVPNConfirm

	// Render confirmation summary
	summaryData := map[string]string{
		"vpn":         sess.VPNData.VPN,
		"server_code": sess.VPNData.ServerCode,
		"relay":       sess.VPNData.Relay,
	}
	jsonBytes, _ := json.MarshalIndent(summaryData, "", "\t")

	menu := &tele.ReplyMarkup{}
	btnCancel := menu.Data("Gajadi", "cancel")
	btnConfirm := menu.Data("Yakin", "confirm")
	menu.Inline(menu.Row(btnCancel, btnConfirm))

	_ = c.Respond()
	return c.EditCaption(string(jsonBytes), menu)
}

func (w *VPNWizard) renderRelayMenu(c tele.Context, sess *Session, relays []model.ProxyNode) error {
	totalRelays := len(sess.VPNData.RelaysCC)
	totalPages := (totalRelays + 5) / 6
	if totalPages == 0 {
		totalPages = 1
	}

	if sess.VPNData.Page < 0 {
		sess.VPNData.Page = 0
	}
	if sess.VPNData.Page >= totalPages {
		sess.VPNData.Page = totalPages - 1
	}

	var b strings.Builder
	b.WriteString("Fool VPN menyediakan jalur relay dengan skema:\n")
	b.WriteString("Kamu ➡️ Server Fool VPN ➡️ Server Relay ➡️ Internet\n\n")
	b.WriteString("Keuntungan relay:\n")
	b.WriteString("• Privasi Lebih Terjaga: Alamat IP asli kamu gak kelihatan.\n")
	b.WriteString("• Anti Blokir: Bebas akses meski ada pembatasan jaringan.\n")
	b.WriteString("• Koneksi Stabil: Relay bikin koneksi tetap lancar meski di jaringan ketat.\n\n")

	if len(relays) > 0 {
		orgMap := make(map[string]bool)
		for _, r := range relays {
			if r.Org != "" {
				orgMap[r.Org] = true
			}
		}
		b.WriteString("Daftar Server:\n")
		count := 0
		for org := range orgMap {
			if count < 10 {
				b.WriteString(fmt.Sprintf("• %s\n", org))
				count++
			}
		}
		b.WriteString("• ...\n\n")
	}

	b.WriteString("Catatan\n")
	b.WriteString("• Relay nambahin latensi/ping\n")
	b.WriteString("• Relay gak bikin lemot (tergantung server relay)\n\n")
	b.WriteString(fmt.Sprintf("%d/%d", sess.VPNData.Page+1, totalPages))

	menu := &tele.ReplyMarkup{}

	startIdx := sess.VPNData.Page * 6
	endIdx := startIdx + 6
	if endIdx > totalRelays {
		endIdx = totalRelays
	}

	var countryButtons []tele.Btn
	for i := startIdx; i < endIdx; i++ {
		cc := sess.VPNData.RelaysCC[i]
		label := cc
		if cc != "Tanpa Relay" {
			label = fmt.Sprintf("%s %s", stringutil.CountryISOToEmoji(cc), cc)
		}
		btn := menu.Data(label, fmt.Sprintf("vpn_relay_%s", cc))
		countryButtons = append(countryButtons, btn)
	}

	var rows []tele.Row
	for i := 0; i < len(countryButtons); i += 3 {
		end := i + 3
		if end > len(countryButtons) {
			end = len(countryButtons)
		}
		rows = append(rows, menu.Row(countryButtons[i:end]...))
	}

	// Navigation Row
	var navRow []tele.Btn
	if sess.VPNData.Page > 0 {
		navRow = append(navRow, menu.Data("〈 Prev", fmt.Sprintf("vpn_page_%d", sess.VPNData.Page-1)))
	} else {
		navRow = append(navRow, menu.Data("⛔️ End", "vpn_end"))
	}

	if sess.VPNData.Page < totalPages-1 {
		navRow = append(navRow, menu.Data("Next 〉", fmt.Sprintf("vpn_page_%d", sess.VPNData.Page+1)))
	} else {
		navRow = append(navRow, menu.Data("End ⛔️", "vpn_end"))
	}
	rows = append(rows, menu.Row(navRow...))

	menu.Inline(rows...)

	_ = c.Respond()
	return c.EditCaption(b.String(), menu, tele.ModeHTML)
}
