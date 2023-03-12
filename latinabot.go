package latinabot

import (
	"os"

	"log"

	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/NicoNex/echotron/v3"
)

type stateFn func(*echotron.Update) stateFn

type bot struct {
	accounts []db.DBScheme
	chatID   int64
	state    stateFn
	echotron.API
}

var (
	botToken = os.Getenv("BOT_TOKEN")
)

func newBot(chatID int64) echotron.Bot {

	bot := &bot{
		chatID: chatID,
		API:    echotron.NewAPI(botToken),
	}

	bot.state = bot.handleMessage
	return bot
}

func (b *bot) Update(update *echotron.Update) {
	b.state = b.state(update)
}

func (b *bot) handleMessage(update *echotron.Update) stateFn {
	if update.Message != nil {
		if update.Message.Text == "/start" {
			b.SendMessage("Have a free VPN account in simple steps !\nPlease select one of the following:", update.ChatID(), &echotron.MessageOptions{
				ParseMode: "HTML",
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: append(helper.BuildInlineKeyboard([]string{"Build API URL", "Get VPN Account"}), []echotron.InlineKeyboardButton{
						{
							Text: "💕 List of Donators ❤️",
							URL:  "https://telegra.ph/Top-Donations-11-05",
						},
					}),
				},
			})
		}
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Data == "Build_API_URL" {
			b.SendMessage("Not implemented yet !", update.ChatID(), nil)
		} else if update.CallbackQuery.Data == "Get_VPN_Account" {
			protocols := []string{}
			b.accounts = db.New().Get("")

			go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)

			for _, account := range b.accounts {
				isExists := false
				for _, protocol := range protocols {
					if account.VPN == protocol {
						isExists = true
						break
					}
				}

				if !isExists {
					protocols = append(protocols, account.VPN)
				}
			}

			if len(protocols) <= 0 {
				go b.SendMessage("No accounts found\nPlease try again later ...", update.ChatID(), nil)
				return b.handleMessage
			}

			go b.SendMessage("Please select VPN protocol:", update.ChatID(), &echotron.MessageOptions{
				ParseMode: "HTML",
				ReplyMarkup: echotron.InlineKeyboardMarkup{
					InlineKeyboard: helper.BuildInlineKeyboard(protocols),
				},
			})

			return b.selectVPN
		}
	}

	return b.handleMessage
}

func Start() {
	dsp := echotron.NewDispatcher(botToken, newBot)
	log.Println(dsp.Poll())
}
