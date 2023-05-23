package latinabot

import (
	"fmt"
	"strings"

	"github.com/LalatinaHub/LatinaApi/common/member"
	"github.com/NicoNex/echotron/v3"
)

func (b *bot) menu(update *echotron.Update) {
	message := []string{fmt.Sprintf("Halo %s", update.Message.From.FirstName)}

	if expired, password := member.GetMember(update.Message.From.ID); expired <= 0 {
		message = append(message, "\nStatus Akun: <b>Premium</b>")
		message = append(message, fmt.Sprintf("Password: <code>%s</code>", password))
		message = append(message, fmt.Sprintf("Masa Aktif: %d Day(s)", expired))

		message = append(message, "\nBatasan:")
		message = append(message, "- Tidak bisa mengambil akun VPN lebih dari 10")

		message = append(message, "\nCatatan:")
		message = append(message, "- Tolong jangan membagikan password demi kebaikan bersama :)")
		message = append(message, "- Segera ganti password kamu apabila bocor ke publik")
		message = append(message, "- Masa aktif akun tidak berlaku akumulasi")
		message = append(message, "- Kirim <code>/setpass PASSWORD</code> untuk merubah password")
	} else {
		message = append(message, "\nStatus Akun: <b>Gratis</b>")
		message = append(message, "Masa Aktif: Lifetime")

		message = append(message, "\nBatasan:")
		message = append(message, "- Tidak bisa mengambil akun VPN lebih dari 3")
		message = append(message, "- Hanya bisa mengambil akun VPN dari SG dan ID")

		message = append(message, "\nCatatan:")
		message = append(message, "- Pembelian premium tidak berlaku refund")
		message = append(message, "- Lihat banner atau hubungi admin untuk order premium")
		message = append(message, "- Harga premium hanya <b>5k</b> perbulan")

		message = append(message, "\nCara Order:")
		message = append(message, "1. Lakukan pembayaran pada e-wallet yang tertera di banner")
		message = append(message, "2. Simpan bukti pembayaran berupa screenshot")
		message = append(message, "3. Kirimkan bukti pembayaran pada bot ini")
		message = append(message, "4. Duduk manis sambil menunggu aktivasi akun diproses")
	}

	message = append(message, "\nAmbil akun VPN gratis full speed dengan langkah mudah !")
	message = append(message, "\n@d_fordlalatina")

	b.SendPhoto(echotron.NewInputFilePath("./assets/Banner.png"), update.ChatID(), &echotron.PhotoOptions{
		Caption:   strings.Join(message, "\n"),
		ParseMode: "HTML",
		ReplyMarkup: echotron.InlineKeyboardMarkup{
			InlineKeyboard: [][]echotron.InlineKeyboardButton{
				{
					{
						Text: "Join Group",
						URL:  "https://t.me/foolvpn",
					},
				},
			},
		},
	})
}
