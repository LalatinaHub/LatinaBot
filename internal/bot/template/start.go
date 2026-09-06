package template

import (
	"fmt"
	"strings"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/common/model"
	tele "gopkg.in/telebot.v3"
)

// BuildStartMessage constructs the HTML formatted text caption for the /start menu.
func BuildStartMessage(from *tele.User, user *model.User, card tarot.Card, donations *trakteer.Response, servers []model.Server) string {
	now := time.Now()
	isDonator := !now.After(user.Expired)

	var b strings.Builder

	b.WriteString(fmt.Sprintf("❡ <b>%s on Service</b> ❡\n", card.Name))
	b.WriteString(fmt.Sprintf("<i><u>%s</u></i>\n\n", card.Message))

	// Profil Kamu
	b.WriteString("<blockquote expandable>")
	b.WriteString("❂ <b>Profil Kamu</b> ❂\n")
	b.WriteString("---------------\n")
	if from != nil {
		b.WriteString(fmt.Sprintf("ID: <code>%d</code>\n", from.ID))
		b.WriteString(fmt.Sprintf("Nama: %s\n", from.FirstName))
	}
	b.WriteString(fmt.Sprintf("Password: <code>%s</code>\n", user.Token))
	statusStr := "Penikmat"
	if isDonator {
		statusStr = "Donator"
	}
	b.WriteString(fmt.Sprintf("Status: <b>%s</b>\n", statusStr))
	b.WriteString(fmt.Sprintf("Expired: %s\n", user.Expired.Format("2006-01-02")))
	b.WriteString("</blockquote>\n\n")

	// Akun VPN
	b.WriteString("<blockquote>")
	b.WriteString("❁ <b>Akun VPN</b> ❁\n")
	b.WriteString("---------------\n")
	b.WriteString(fmt.Sprintf("Tipe: %s\n", user.VPN))
	b.WriteString(fmt.Sprintf("UUID: <tg-spoiler>%s</tg-spoiler>\n", user.Password))
	if user.ServerCode != "" {
		b.WriteString(fmt.Sprintf("Server Code: %s\n", user.ServerCode))
		var srvDomain string
		for _, s := range servers {
			if s.Code == user.ServerCode {
				srvDomain = s.Domain
				break
			}
		}
		if srvDomain != "" {
			b.WriteString(fmt.Sprintf("Domain: %s\n", srvDomain))
		}
	}
	b.WriteString(fmt.Sprintf("Relay: %s\n", user.Relay))
	b.WriteString(fmt.Sprintf("Quota: %d MB\n", user.Quota))
	adblockStr := "Mati"
	if user.Adblock {
		adblockStr = "Hidup"
	}
	b.WriteString(fmt.Sprintf("Adblock: %s\n", adblockStr))
	b.WriteString("</blockquote>\n\n")

	// Informasi
	b.WriteString("<blockquote>")
	b.WriteString("✤ <b>Informasi</b> ✤\n")
	b.WriteString("---------------\n")
	b.WriteString("☾ Maksimal 10 akun API\n")
	b.WriteString("☾ 1x donasi untuk 29 hari premium, berapapun jumlahnya")
	b.WriteString("</blockquote>\n\n")

	// Para Supreme Being
	b.WriteString("<blockquote expandable>")
	b.WriteString("※ <b>Para Supreme Being</b> ※\n")
	b.WriteString("---------------\n")
	if donations != nil && len(donations.Result.Data) > 0 {
		for _, don := range donations.Result.Data {
			b.WriteString(fmt.Sprintf("<b>%s [%d]</b>\n", don.SupporterName, don.Quantity))
			b.WriteString(fmt.Sprintf("<tg-spoiler>%s</tg-spoiler>\n\n", don.SupportMessage))
		}
	}
	b.WriteString("</blockquote>\n\n\n")

	b.WriteString(time.Now().Format(time.RFC1123))

	return b.String()
}

// BuildStartKeyboard generates the inline keyboard markup for the start screen.
func BuildStartKeyboard(user *model.User) *tele.ReplyMarkup {
	menu := &tele.ReplyMarkup{}

	subURL := fmt.Sprintf(
		"https://api.foolvpn.web.id/sub?format=raw&cdn=104.18.2.2&sni=google.com&mode=cdn,sni&region=Asia&vpn=vmess,vless,trojan&pass=%s",
		user.Token,
	)

	btnAmbilAkun := menu.URL("Ambil Akun", subURL)
	btnBuatAkun := menu.Data("Buat Akun", "c_vpn")

	btnWebsite := menu.URL("Website", "https://foolvpn.web.id")
	btnGrup := menu.URL("Grup", "https://t.me/foolvpn")
	btnConverter := menu.URL("Converter", "https://t.me/subxfm_bot")

	btnGantiPass := menu.Data("Ganti Password", "c_pass")
	btnGantiUUID := menu.Data("Ganti UUID", "c_uuid")

	adblockLabel := "Hidupkan Adblock"
	if user.Adblock {
		adblockLabel = "Matikan Adblock"
	}
	btnToggleAdblock := menu.Data(adblockLabel, "s_adblock")

	btnListWildcard := menu.Data("List Wildcard", "l_wildcard")
	btnDisclaimer := menu.Data("❗️ Desclaimer ❗️", "t_desclaimer")

	btnCaraDonasi := menu.Data("Cara Donasi", "t_donasi")
	btnTrakteer := menu.URL("Trakteer", "https://trakteer.id/dickymuliafiqri/tip")

	btnRefresh := menu.Data("🔄", "m_refresh")
	btnUptime := menu.URL("ℹ️", "https://foolvpn.me/uptime")

	menu.Inline(
		menu.Row(btnAmbilAkun, btnBuatAkun),
		menu.Row(btnWebsite, btnGrup, btnConverter),
		menu.Row(btnGantiPass, btnGantiUUID),
		menu.Row(btnToggleAdblock),
		menu.Row(btnListWildcard),
		menu.Row(btnDisclaimer),
		menu.Row(btnCaraDonasi, btnTrakteer),
		menu.Row(btnRefresh, btnUptime),
	)

	return menu
}
