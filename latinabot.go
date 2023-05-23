package latinabot

import (
	"fmt"
	"os"
	"strconv"

	"strings"

	"log"

	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/NicoNex/echotron/v3"
)

type stateFn func(*echotron.Update) stateFn

type bot struct {
	chatID int64
	state  stateFn
	echotron.API
}

var (
	adminID, _ = strconv.ParseInt(os.Getenv("ADMIN_ID"), 10, 64)
	botToken   = os.Getenv("BOT_TOKEN")
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
	// Ignore not private chat
	if update.ChatID() < 0 {
		// go b.SendMessage("Please chat me in private", update.ChatID(), nil)
		return b.handleMessage
	}

	if update.Message != nil {
		if update.Message.Text == "/start" {
			go b.menu(update)
		} else if strings.HasPrefix(update.Message.Text, "/setpass") {
			var (
				values          = strings.Split(update.Message.Text, " ")
				password string = ""
			)

			if len(values) == 2 {
				password = values[1]
			} else {
				password = member.GenerateHash(strconv.FormatInt(update.ChatID(), 10))
			}

			if member.ChangePassword(update.ChatID(), password) {
				b.SendMessage("Berhasil merubah password !", update.ChatID(), nil)
				b.menu(update)
			} else {
				b.SendMessage("Gagal merubah password !", update.ChatID(), nil)
			}
		} else if strings.HasPrefix(update.Message.Text, "/member") {
			if update.ChatID() == adminID {
				values := strings.Split(update.Message.Text, " ")

				if len(values) == 3 {
					var (
						id, _   = strconv.ParseInt(values[1], 10, 64)
						subs, _ = strconv.Atoi(values[2])
					)

					b.SendMessage(fmt.Sprintf("Menambahkan %d untuk menjadi premium selama %d bulan ...", id, subs), update.ChatID(), nil)
					if member.UpdateMember(id, subs) {
						b.SendMessage("Berhasil menambahkan member premium !", update.ChatID(), nil)

						if subs > 0 {
							b.SendMessage(fmt.Sprintf("Kamu terdaftar sebagai premium selama %d bulan !", subs), id, nil)
						} else {
							b.SendMessage(fmt.Sprintf("Masa Aktif akun kamu diturunkan selama %d bulan :(", subs), id, nil)
						}
					} else {
						b.SendMessage("Gagal menambahkan member premium !", update.ChatID(), nil)
					}

					return b.handleMessage
				}
				b.SendMessage("Format pesan tidak sesuai !", update.ChatID(), nil)
			}
		} else if update.Message.Photo != nil {
			b.ForwardMessage(adminID, update.ChatID(), update.Message.ID, nil)
			b.SendMessage("Bukti pembayaran berhasil dikirimkan ke admin !\nMohon tunggu pemberitahuan dari bot", update.ChatID(), nil)
		}
	}

	return b.handleMessage
}

func Start() {
	dsp := echotron.NewDispatcher(botToken, newBot)
	log.Println(dsp.Poll())
}
