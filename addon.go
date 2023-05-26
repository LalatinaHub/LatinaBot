package latinabot

import (
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/NicoNex/echotron/v3"
)

func (b *bot) menu(update *echotron.Update) {
	var (
		expired, password = member.GetMember(update.ChatID())
		message           = []string{fmt.Sprintf("Halo %s !", update.Message.From.FirstName)}
	)

	message = append(message, fmt.Sprintf("\nID: <code>%d</code>", update.ChatID()))
	if expired <= 0 {
		message = append(message, "Status Akun: <b>Premium</b> 👑")
		message = append(message, fmt.Sprintf("Password: <code>%s</code>", password))
		message = append(message, fmt.Sprintf("Masa Aktif: %d Day(s)", expired))

		message = append(message, "\nBatasan:")
		message = append(message, "- Tidak bisa mengambil akun VPN lebih dari 10")

		message = append(message, "\nCatatan:")
		message = append(message, "- Masa aktif akun tidak berlaku akumulasi")
		message = append(message, "- Ketahuan nakal = Premium hangus")
	} else {
		message = append(message, "Status Akun: <b>Gratis</b> 👒")
		message = append(message, fmt.Sprintf("Password: <code>%s</code>", password))
		message = append(message, "Masa Aktif: Lifetime")

		message = append(message, "\nBatasan:")
		message = append(message, "- Tidak bisa mengambil akun VPN lebih dari 3")
		message = append(message, "- Hanya bisa mengambil akun VPN dari SG dan ID")
		message = append(message, "- Hanya bisa mengambil akun VMess")

		message = append(message, "\nCara Order:")
		message = append(message, "1. Lakukan pembayaran pada e-wallet (lihat banner)")
		message = append(message, "2. Simpan bukti pembayaran berupa screenshot")
		message = append(message, "3. Kirimkan bukti pembayaran pada bot ini")
		message = append(message, "4. Duduk manis sambil menunggu aktivasi akun diproses")

		message = append(message, "\nCatatan:")
		message = append(message, "- Harga premium hanya <b>7k</b> perbulan")
	}

	message = append(message, "- Tidak ada refund")
	message = append(message, "- Dilarang membagikan password")
	message = append(message, "- Segera ganti password apabila bocor ke publik")
	message = append(message, "- Kirim /newpass untuk memperbarui password")

	message = append(message, "\nAmbil akun VPN gratis full speed dengan langkah mudah !")
	message = append(message, "\n@d_fordlalatina")

	go b.SendPhoto(echotron.NewInputFileURL("https://raw.githubusercontent.com/LalatinaHub/LatinaBot/main/assets/Banner.png"), update.ChatID(), &echotron.PhotoOptions{
		Caption:   strings.Join(message, "\n"),
		ParseMode: "HTML",
		ReplyMarkup: echotron.InlineKeyboardMarkup{
			InlineKeyboard: [][]echotron.InlineKeyboardButton{
				{
					{
						Text: "😎 Ambil Akun",
						URL:  "https://fool.azurewebsites.net/get?format=raw&region=Asia&cdn=bug.com&sni=bug.com&pass=" + password,
					},
					{
						Text: "Tutorial 🤔",
						URL:  "https://fool.azurewebsites.net/api/get.html",
					},
				},
				{
					{
						Text: "Gabung Grup",
						URL:  "https://t.me/foolvpn",
					},
				},
			},
		},
	})
}
