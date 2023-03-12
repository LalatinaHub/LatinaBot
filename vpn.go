package latinabot

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/LalatinaHub/LatinaApi/common/account/converter"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/NicoNex/echotron/v3"
)

func (b *bot) selectVPN(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		var (
			ccs         []string
			newAccounts []db.DBScheme
			protocol    = update.CallbackQuery.Data
		)

		go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

		for _, account := range b.accounts {
			if account.VPN == protocol {
				isExists := false
				for _, cc := range ccs {
					if cc == account.CountryCode {
						isExists = true
						break
					}
				}

				newAccounts = append(newAccounts, account)

				if !isExists {
					ccs = append(ccs, account.CountryCode)
				}
			}
		}

		go b.SendMessage("Select country code:", update.ChatID(), &echotron.MessageOptions{
			ParseMode: "HTML",
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: helper.BuildInlineKeyboard(append(ccs, "Random")),
			},
		})

		b.accounts = newAccounts
		return b.selectCC
	}

	return b.handleMessage
}

func (b *bot) selectCC(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		var (
			modes       []string
			newAccounts []db.DBScheme
			cc          = update.CallbackQuery.Data
		)

		go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

		for _, account := range b.accounts {
			if cc == "Random" || cc == account.CountryCode {
				isExists := false
				for _, mode := range modes {
					if mode == account.ConnMode {
						isExists = true
						break
					}
				}

				newAccounts = append(newAccounts, account)

				if !isExists {
					modes = append(modes, account.ConnMode)
				}
			}
		}

		go b.SendMessage("Select connection mode:", update.ChatID(), &echotron.MessageOptions{
			ParseMode: "HTML",
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: helper.BuildInlineKeyboard(append(modes, "Random")),
			},
		})

		b.accounts = newAccounts
		return b.selectMode
	}

	return b.handleMessage
}

func (b *bot) selectMode(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		var (
			tlss        []string
			newAccounts []db.DBScheme
			mode        = update.CallbackQuery.Data
		)

		go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

		for _, account := range b.accounts {
			if mode == "Random" || mode == account.ConnMode {
				isExists := false
				tlsstr := "False"

				if account.TLS {
					tlsstr = "True"
				}

				for _, tls := range tlss {
					if tls == tlsstr {
						isExists = true
						break
					}
				}

				newAccounts = append(newAccounts, account)

				if !isExists {
					tlss = append(tlss, tlsstr)
				}
			}
		}

		go b.SendMessage("Select TLS:", update.ChatID(), &echotron.MessageOptions{
			ParseMode: "HTML",
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: helper.BuildInlineKeyboard(append(tlss, "Random")),
			},
		})

		b.accounts = newAccounts
		return b.selectTLS
	}

	return b.handleMessage
}

func (b *bot) selectTLS(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		var (
			networks    []string
			newAccounts []db.DBScheme
			tls         = update.CallbackQuery.Data
		)

		go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

		for _, account := range b.accounts {
			var (
				isExists = false
				isMatch  = false
			)

			if tls == "Random" {
				newAccounts = append(newAccounts, account)
				isMatch = true
			} else if tls == "True" {
				if account.TLS {
					newAccounts = append(newAccounts, account)
					isMatch = true
				}
			} else {
				if !account.TLS {
					newAccounts = append(newAccounts, account)
					isMatch = true
				}
			}

			if isMatch {
				for _, network := range networks {
					if network == account.Transport {
						isExists = true
						break
					}
				}
			}

			if !isExists && isMatch {
				networks = append(networks, account.Transport)
			}
		}

		go b.SendMessage("Select network/transport:", update.ChatID(), &echotron.MessageOptions{
			ParseMode: "HTML",
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: helper.BuildInlineKeyboard(append(networks, "Random")),
			},
		})

		b.accounts = newAccounts
		return b.finalVPN
	}

	return b.handleMessage
}

func (b *bot) finalVPN(update *echotron.Update) stateFn {
	if update.CallbackQuery != nil {
		var (
			message     []string
			newAccounts []db.DBScheme
			network     = update.CallbackQuery.Data
		)

		// Ignore special callback data
		switch network {
		case "Get_Another":
			newAccounts = b.accounts
		case "Done":
			go b.SendMessage("Thanks for using our service\nPlease consider donating 🥺", update.ChatID(), &echotron.MessageOptions{
				ParseMode: "HTML",
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: [][]echotron.InlineKeyboardButton{
						{
							echotron.InlineKeyboardButton{
								Text: "Donate Me 💕",
								URL:  "https://saweria.co/m0qa",
							},
						},
					},
				},
			})

			return b.handleMessage
		default:
			for _, account := range b.accounts {
				if network == "Random" || network == account.Transport {
					newAccounts = append(newAccounts, account)
				}
			}
			b.accounts = newAccounts
		}

		go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

		account := newAccounts[rand.Intn(len(newAccounts))]

		message = append(message, fmt.Sprintf("<code>REMARKS       : %s</code>", account.Remark))
		message = append(message, fmt.Sprintf("<code>SERVER        : %s</code>", account.Server))
		message = append(message, fmt.Sprintf("<code>HOST          : %s</code>", account.Host))
		message = append(message, fmt.Sprintf("<code>PORT          : %d</code>", account.ServerPort))
		message = append(message, fmt.Sprintf("<code>UUID          : %s</code>", account.UUID))
		message = append(message, fmt.Sprintf("<code>PASSWORD      : %s</code>", account.Password))
		message = append(message, fmt.Sprintf("<code>ISP           : %s</code>", account.Org))
		message = append(message, fmt.Sprintf("<code>Protocol      : %s</code>", account.VPN))
		message = append(message, fmt.Sprintf("<code>CC            : %s</code>", account.CountryCode))
		message = append(message, fmt.Sprintf("<code>Mode          : %s</code>", account.ConnMode))
		message = append(message, fmt.Sprintf("<code>TLS           : %t</code>", account.TLS))
		message = append(message, fmt.Sprintf("<code>Network       : %s</code>", account.Transport))
		message = append(message, fmt.Sprintf("<code>PATH          : %s</code>", account.Path))
		message = append(message, fmt.Sprintf("<code>SERVICE NAME  : %s</code>", account.ServiceName))
		message = append(message, fmt.Sprintf("<code>%s</code>", "---------------"))
		message = append(message, fmt.Sprintf("<code>%s</code>", converter.ToRaw([]db.DBScheme{account})))
		message = append(message, fmt.Sprintf("<code>%s</code>", "---------------"))

		go b.SendMessage(strings.Join(message, "\n"), update.ChatID(), &echotron.MessageOptions{
			ParseMode: "HTML",
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: [][]echotron.InlineKeyboardButton{
					{
						echotron.InlineKeyboardButton{
							Text:         "Get Another",
							CallbackData: "Get_Another",
						},
						echotron.InlineKeyboardButton{
							Text:         "Done",
							CallbackData: "Done",
						},
					},
				},
			},
		})

		return b.finalVPN
	}

	go b.SendMessage("Send /start for start using this bot again !", update.ChatID(), nil)
	return b.handleMessage
}
