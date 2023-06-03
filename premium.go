package latinabot

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/NicoNex/echotron/v3"
)

type PremiumVPNInfo struct {
	VPN string
}

type PremiumDomainInfo struct {
	Domain   string
	Code     string
	Populate int
}

var (
	premiumVpnInfo = PremiumVPNInfo{}
	domains        = []PremiumDomainInfo{}
)

func (b *bot) handlePremiumType(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		premiumVpnInfo.VPN = update.CallbackQuery.Data
		domains = []PremiumDomainInfo{}

		var (
			message, domainsCode []string
		)

		rows, err := db.New().Conn().Query("SELECT domain, location, populate FROM domains")
		if err != nil {
			fmt.Println(err)
		}

		for rows.Next() {
			var (
				domain, location sql.NullString
				populate         sql.NullInt16
			)

			rows.Scan(&domain, &location, &populate)

			domainsCode = append(domainsCode, location.String)
			domains = append(domains, PremiumDomainInfo{
				Domain:   domain.String,
				Code:     location.String,
				Populate: int(populate.Int16),
			})
		}

		message = append(message, "Daftar Pengguna/Populasi Server")
		for _, domain := range domains {
			message = append(message, fmt.Sprintf("%s: %d", domain.Code, domain.Populate))
		}
		message = append(message, "\nSilahkan pilih lokasi akun:")

		b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
		go b.SendMessage(strings.Join(message[:], "\n"), update.ChatID(), &echotron.MessageOptions{
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: helper.BuildInlineKeyboard(domainsCode),
			},
		})

		return b.handlePremiumCreate
	}

	return b.handlePremiumType
}

func (b *bot) handlePremiumCreate(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		for _, domain := range domains {
			if domain.Code == update.CallbackQuery.Data {
				var message string
				if member.CreatePremiumAccount(update.ChatID(), premiumVpnInfo.VPN, domain.Domain) {
					message = "Akun berhasil dibuat !"
				} else {
					message = "Akun gagal dibuat !"
				}

				b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
				b.SendMessage(message, update.ChatID(), nil)
				go b.menu(update)

				break
			}
		}

		return b.handleMessage
	}

	return b.handlePremiumCreate
}
