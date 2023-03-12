package latinabot

import (
	"fmt"
	"os"
	"strings"

	"log"

	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
	"github.com/NicoNex/echotron/v3"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type stateFn func(*echotron.Update) stateFn

type vpnTempData struct {
	ccs []string
}

type bot struct {
	accounts []db.DBScheme
	chatID   int64
	state    stateFn
	caser    cases.Caser
	echotron.API
	vpnTempData
}

var (
	botToken = os.Getenv("BOT_TOKEN")
)

func newBot(chatID int64) echotron.Bot {

	bot := &bot{
		chatID: chatID,
		API:    echotron.NewAPI(botToken),
		caser:  cases.Title(language.English),
	}

	bot.state = bot.handleMessage
	return bot
}

func (b *bot) Update(update *echotron.Update) {
	b.state = b.state(update)
}

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
			}),
		},
	})
}

func (b *bot) handleMessage(update *echotron.Update) stateFn {
	if update.Message != nil {
		if update.Message.Text == "/start" {
			go b.menu(update)
		}
	} else if update.CallbackQuery != nil {
		if update.CallbackQuery.Data == "menu" {
			go b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
			go b.menu(update)
		} else if update.CallbackQuery.Data == "Build_API_URL" {
			b.SendMessage("Not implemented yet !", update.ChatID(), nil)
		} else if update.CallbackQuery.Data == "Get_VPN_Account" {
			var (
				protocols      []string
				protocolsCount []string
				message        []string
			)
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
				protocolsCount = append(protocolsCount, account.VPN)
			}

			if len(protocols) <= 0 {
				go b.SendMessage("No accounts found\nPlease try again later ...", update.ChatID(), nil)
				return b.handleMessage
			}

			message = append(message, "Please select VPN protocol:")
			message = append(message, "")
			for _, protocol := range protocols {
				var count int
				for _, protocolCount := range protocolsCount {
					if protocolCount == protocol {
						count++
					}
				}
				message = append(message, fmt.Sprintf("%s : %d", b.caser.String(protocol), count))
			}

			go b.SendMessage(strings.Join(message, "\n"), update.ChatID(), &echotron.MessageOptions{
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
