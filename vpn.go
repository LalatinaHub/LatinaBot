package latinabot

import (
	"math/rand"
	"strconv"

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
				InlineKeyboard: helper.BuildInlineKeyboardWithPage(append(ccs, "Random"), 0),
			},
		})

		b.ccs = ccs
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

		if page, err := strconv.Atoi(update.CallbackQuery.Data); err == nil {
			go b.EditMessageReplyMarkup(echotron.NewMessageID(update.ChatID(), update.CallbackQuery.Message.ID), &echotron.MessageReplyMarkup{
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: helper.BuildInlineKeyboardWithPage(append(b.ccs, "Random"), page),
				},
			})

			return b.selectCC
		}

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
			newAccounts []db.DBScheme
			network     = update.CallbackQuery.Data
		)

		// Special callback data
		if network == "Done" {
			go b.EditMessageReplyMarkup(echotron.NewMessageID(update.ChatID(), update.CallbackQuery.Message.ID), nil)
			go b.SendMessage("Thanks for using our service\nPlease consider donating 🥺", update.ChatID(), &echotron.MessageOptions{
				ParseMode: "HTML",
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: [][]echotron.InlineKeyboardButton{
						{
							{
								Text:         "Main Menu",
								CallbackData: "menu",
							},
						},
					},
				},
			})

			return b.handleMessage
		}

		for _, account := range b.accounts {
			if network == "Random" || network == account.Transport {
				newAccounts = append(newAccounts, account)
			}
		}

		go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

		account := newAccounts[rand.Intn(len(newAccounts))]
		message := helper.MakeVPNMessage(account)

		go b.SendMessage(message, update.ChatID(), &echotron.MessageOptions{
			ParseMode: "HTML",
			ReplyMarkup: echotron.InlineKeyboardMarkup{
				InlineKeyboard: [][]echotron.InlineKeyboardButton{
					{
						{
							Text:         "Re-Generate",
							CallbackData: network,
						},
						{
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
