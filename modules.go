package latinabot

import (
	A "github.com/LalatinaHub/LatinaApi/common/account"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/NicoNex/echotron/v3"
)

func SendVPNToTopic(chatID int64, topicID int) {
	var (
		bug     = []string{"BUG.COM"}
		account = A.PopulateBugs(A.Get("WHERE VPN != shadowsocks ORDER BY RANDOM() LIMIT 1"), bug, bug)[0]
		message = helper.MakeVPNMessage(account)
	)

	go SendMessageToTopic(message, chatID, topicID)
}

func SendMessageToTopic(message string, chatID int64, topicID int) (echotron.APIResponse, error) {
	var (
		b = &bot{
			API: echotron.NewAPI(botToken),
		}
	)

	return b.SendMessage(message, chatID, &echotron.MessageOptions{
		ParseMode:        "HTML",
		ReplyToMessageID: int(topicID),
	})
}
