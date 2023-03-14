package latinabot

import (
	A "github.com/LalatinaHub/LatinaApi/common/account"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/NicoNex/echotron/v3"
)

func Client() *bot {
	return &bot{
		API: echotron.NewAPI(botToken),
	}
}

func SendVPNToTopic(chatID int64, topicID int) {
	var (
		bug     = []string{"BUG.COM"}
		account = A.PopulateBugs(A.Get("WHERE VPN != 'shadowsocks' ORDER BY RANDOM() LIMIT 1"), bug, bug)[0]
		message = helper.MakeVPNMessage(account)
	)

	go Client().SendMessage(message, chatID, &echotron.MessageOptions{
		ParseMode:           "HTML",
		ReplyToMessageID:    int(topicID),
		DisableNotification: true,
	})
}
