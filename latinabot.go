package latinabot

import (
	"database/sql"
	"fmt"
	"os"
	"strconv"

	"strings"

	"log"

	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/LalatinaHub/LatinaSub-go/db"
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
			if !member.IsExists(update.ChatID()) {
				member.UpdateMember(update.ChatID(), -1)
			}
			go b.menu(update)
		} else if strings.HasPrefix(update.Message.Text, "/newpass") {
			if member.ChangePassword(update.ChatID()) {
				go b.SendMessage("Password berhasil diperbarui", update.ChatID(), nil)
				go b.menu(update)
			} else {
				go b.SendMessage("Password gagal diperbarui", update.ChatID(), nil)
			}
		} else if strings.HasPrefix(update.Message.Text, "/resetpass") {
			premiumData := member.GetPremiumAccount(update.ChatID())
			if premiumData.Domain != "" {
				if member.UpdatePremiumPassword(update.ChatID(), premiumData.Domain) {
					go b.SendMessage("Password berhasil diperbarui", update.ChatID(), nil)
					go b.menu(update)
				} else {
					go b.SendMessage("Password gagal diperbarui", update.ChatID(), nil)
				}
			}
		} else if strings.HasPrefix(update.Message.Text, "/dbq") {
			if update.ChatID() == adminID {
				_, e := b.SendDocument(echotron.NewInputFilePath("DB_QUERY.txt"), adminID, nil)
				if e != nil {
					go b.SendMessage(e.Error(), adminID, nil)
				}
			}
		} else if strings.HasPrefix(update.Message.Text, "/member") {
			if update.ChatID() == adminID {
				values := strings.Split(update.Message.Text, " ")

				if len(values) == 3 {
					var (
						id, _   = strconv.ParseInt(values[1], 10, 64)
						subs, _ = strconv.Atoi(values[2])
					)

					go b.SendMessage(fmt.Sprintf("Menambahkan %d untuk menjadi premium selama %d bulan ...", id, subs), update.ChatID(), nil)
					if member.UpdateMember(id, subs) {
						go b.SendMessage("Berhasil menambahkan member premium !", update.ChatID(), nil)

						if subs > 0 {
							go b.SendMessage(fmt.Sprintf("Masa aktif premium kamu ditambahkan selama %d bulan !\n\nNb: Hubungi admin jika jumlah quota tidak bertambah", subs), id, nil)
						} else if subs < 0 {
							go b.SendMessage(fmt.Sprintf("Masa aktif premium kamu diturunkan selama %d bulan :(", subs), id, nil)
						}
					} else {
						go b.SendMessage("Gagal menambahkan member premium !", update.ChatID(), nil)
					}

					return b.handleMessage
				}
				go b.SendMessage("Format pesan tidak sesuai !", update.ChatID(), nil)
			}
		} else if update.Message.Photo != nil {
			res, _ := b.ForwardMessage(adminID, update.ChatID(), update.Message.ID, nil)
			go b.SendMessage(fmt.Sprintf("Bukti pembayaran dari <code>%d</code>", update.ChatID()), adminID, &echotron.MessageOptions{
				ReplyToMessageID: res.Result.ID,
				ParseMode:        "HTML",
			})
			go b.SendMessage("Bukti pembayaran berhasil dikirimkan ke admin !\nMohon tunggu pemberitahuan dari bot", update.ChatID(), nil)
		}
	} else if update.CallbackQuery != nil {
		switch update.CallbackQuery.Data {
		case "create_account":
			if premium, _ := member.GetMember(update.ChatID()); premium < 1 {
				var (
					vpnTypes []string
				)

				rows, _ := db.New().Conn().Query("SELECT protocol FROM protocols")

				defer rows.Close()

				for rows.Next() {
					var vpn sql.NullString

					rows.Scan(&vpn)
					vpnTypes = append(vpnTypes, vpn.String)
				}

				b.DeleteMessage(update.ChatID(), update.CallbackQuery.Message.ID)
				go b.SendMessage("Silahkan pilih protokol:", update.ChatID(), &echotron.MessageOptions{
					ReplyMarkup: echotron.InlineKeyboardMarkup{
						InlineKeyboard: helper.BuildInlineKeyboard(vpnTypes),
					},
				})

				return b.handlePremiumType
			}
		}
	}

	return b.handleMessage
}

func Start() {
	dsp := echotron.NewDispatcher(botToken, newBot)
	log.Println(dsp.Poll())
}
