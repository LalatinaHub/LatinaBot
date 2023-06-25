package latinabot

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"

	apiHelper "github.com/LalatinaHub/LatinaApi/api/helper"
	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/LalatinaHub/LatinaSub-go/ipapi"
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
	relayCountries = []string{}
	password       = os.Getenv("PASSWORD")
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
			message = append(message, fmt.Sprintf("<a href='%s'>%s</a>: %d", domain.Domain, domain.Code, domain.Populate))
		}
		message = append(message, "\nSilahkan pilih lokasi akun:")

		b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
		go b.SendMessage(strings.Join(message[:], "\n"), update.ChatID(), &echotron.MessageOptions{
			ParseMode:             "HTML",
			DisableWebPagePreview: true,
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

				for _, proxy := range proxies {
					if premiumVpnInfo.CC != proxy.CountryCode && proxy.CountryCode != "" {
						for _, country := range ipapi.CountryList {
							if country.Code == proxy.CountryCode {
								isExists := func() bool {
									for _, countryName := range relayCountries {
										if countryName == country.Name {
											return true
										}
									}
									return false
								}()

								if !isExists {
									relayCountries = append(relayCountries, country.Name)
								}
							}
						}
					}
				}

				sort.Strings(relayCountries)
				relayCountries = append([]string{"Tanpa Relay"}, relayCountries...)

				message := []string{"Silahkan pilih relay !"}
				message = append(message, "\nSkema dengan relay:")
				message = append(message, "<code>HP -> Server Fool -> Server Relay -> Internet</code>")
				message = append(message, "\nSkema tanpa relay:")
				message = append(message, "<code>HP -> Server Fool -> Internet</code>")

				b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
				go b.SendMessage(strings.Join(message[:], "\n"), update.ChatID(), &echotron.MessageOptions{
					ParseMode: "HTML",
					ReplyMarkup: echotron.InlineKeyboardMarkup{
						InlineKeyboard: helper.BuildInlineKeyboardWithPage(relayCountries, 0),
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
			relayCode = strings.ReplaceAll(update.CallbackQuery.Data, "_", " ")
		)

		if relayCode == relayCountries[0] {
			relayCode = ""
		}

		if page, err := strconv.Atoi(relayCode); err == nil {
			b.EditMessageReplyMarkup(echotron.NewMessageID(update.ChatID(), update.CallbackQuery.Message.ID), &echotron.MessageReplyMarkup{
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: helper.BuildInlineKeyboardWithPage(relayCountries, page),
				},
			})

			return b.handlePremiumCreate
		}

		for _, country := range ipapi.CountryList {
			if country.Name == relayCode {
				relayCode = country.Code
				break
			}
		}

		if member.CreatePremiumAccount(update.ChatID(), premiumVpnInfo.VPN, premiumVpnInfo.Domain, relayCode) {
			for _, domain := range domains {
				apiHelper.Fetch("http://" + domain.Domain + "/" + password)
			}
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
