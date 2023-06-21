package latinabot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/NicoNex/echotron/v3"
)

type PremiumVPNInfo struct {
	VPN    string
	Domain string
	CC     string
}

type PremiumDomainInfo struct {
	Domain   string
	Populate int
	Location string
	Code     string
}

var (
	premiumVpnInfo = PremiumVPNInfo{}
	domains        = []PremiumDomainInfo{}
	relayCodes     = []string{"Tanpa Relay"}
)

func (b *bot) handlePremiumType(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		premiumVpnInfo.VPN = update.CallbackQuery.Data
		domains = []PremiumDomainInfo{}

		var (
			message, domainsCode []string
		)

		rows, err := db.New().Conn().Query("SELECT domain, location, populate, code FROM domains")
		if err != nil {
			fmt.Println(err)
		}

		for rows.Next() {
			var (
				domain, location, code sql.NullString
				populate               sql.NullInt16
			)

			rows.Scan(&domain, &location, &populate, &code)

			domainsCode = append(domainsCode, code.String)
			domains = append(domains, PremiumDomainInfo{
				Domain:   domain.String,
				Populate: int(populate.Int16),
				Location: location.String,
				Code:     code.String,
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

		return b.handlePremiumServer
	}

	return b.handlePremiumType
}

func (b *bot) handlePremiumServer(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		for _, domain := range domains {
			if domain.Code == update.CallbackQuery.Data {
				premiumVpnInfo.Domain = domain.Domain
				premiumVpnInfo.CC = domain.Location

				var buf = new(strings.Builder)
				var proxies = []db.DBScheme{}
				resp, err := apiHelper.Fetch(fmt.Sprintf("http://%s/relay", domain.Domain))
				if err != nil {
					fmt.Println(err)
				}
				defer resp.Body.Close()

				io.Copy(buf, resp.Body)
				if resp.StatusCode == 200 {
					json.Unmarshal([]byte(buf.String()), &proxies)
				}

				var rCodes = []string{}
				for _, proxy := range proxies {
					if premiumVpnInfo.CC != proxy.CountryCode && proxy.CountryCode != "" {
						isExists := func() bool {
							for _, code := range rCodes {
								if code == proxy.CountryCode {
									return true
								}
							}
							return false
						}()

						if !isExists {
							rCodes = append(rCodes, proxy.CountryCode)
						}
					}
				}

				sort.Strings(rCodes)
				relayCodes = append(relayCodes, rCodes...)

				message := []string{"Silahkan pilih relay !"}
				message = append(message, "\nSkema dengan relay:")
				message = append(message, "<code>HP -> Server Fool -> Server Relay -> Internet</code>")
				message = append(message, "\nSkema tanpa relay:")
				message = append(message, "<code>HP -> Server Fool -> Internet</code>")

				b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
				go b.SendMessage(strings.Join(message[:], "\n"), update.ChatID(), &echotron.MessageOptions{
					ParseMode: "HTML",
					ReplyMarkup: echotron.InlineKeyboardMarkup{
						InlineKeyboard: helper.BuildInlineKeyboardWithPage(relayCodes, 0),
					},
				})
				break
			}
		}

		return b.handlePremiumCreate
	}

	return b.handlePremiumServer
}

func (b *bot) handlePremiumCreate(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		var (
			message   string
			relayCode = update.CallbackQuery.Data
		)

		if len(relayCode) > 5 {
			relayCode = premiumVpnInfo.CC
		}

		if page, err := strconv.Atoi(relayCode); err == nil {
			b.EditMessageReplyMarkup(echotron.NewMessageID(update.ChatID(), update.CallbackQuery.Message.ID), &echotron.MessageReplyMarkup{
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: helper.BuildInlineKeyboardWithPage(relayCodes, page),
				},
			})

			return b.handlePremiumCreate
		}

		if member.CreatePremiumAccount(update.ChatID(), premiumVpnInfo.VPN, premiumVpnInfo.Domain, relayCode) {
			message = "Akun berhasil dibuat !"
		} else {
			message = "Akun gagal dibuat !"
		}

		b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
		b.SendMessage(message, update.ChatID(), nil)
		go b.menu(update)

		return b.handleMessage
	}

	return b.handlePremiumCreate
}
