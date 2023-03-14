package latinabot

import (
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/NicoNex/echotron/v3"
)

func (b *bot) menu(update *echotron.Update) {
	go b.SendMessage("Have a free VPN account in simple steps !\nPlease select one of the following:", update.ChatID(), &echotron.MessageOptions{
		ParseMode: "HTML",
		ReplyMarkup: echotron.InlineKeyboardMarkup{
			InlineKeyboard: append(helper.BuildInlineKeyboard([]string{"Build API URL", "Get VPN Account"}), []echotron.InlineKeyboardButton{
				{
					Text: "🌻 List of Donators",
					URL:  "https://telegra.ph/Top-Donations-11-05",
				},
				{
					Text: "Donate Me 🌱",
					URL:  "https://saweria.co/m0qa",
				},
			}, []echotron.InlineKeyboardButton{
				{
					Text: "Join Our Group",
					URL:  "https://t.me/foolvpn",
				},
			}),
		},
	})
}
