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
	db       db.DB
	chatID   int64
	state    stateFn
	echotron.API
}

var (
	botToken = os.Getenv("BOT_TOKEN")
)

func newBot(chatID int64) echotron.Bot {
	bot := &bot{
		db:     *db.New(),
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
					InlineKeyboard: helper.BuildInlineKeyboard([]string{"Build API URL", "Get VPN Account"}),
				},
			})
		}
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Data == "Build_API_URL" {
			b.SendMessage("Not implemented yet !", update.ChatID(), nil)
		} else if update.CallbackQuery.Data == "Get_VPN_Account" {
			var (
				protocols []string
			)
			b.accounts = b.db.Get("")

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
