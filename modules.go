package latinabot

import (
	A "github.com/LalatinaHub/LatinaApi/common/account"
	"github.com/LalatinaHub/LatinaBot/helper"
	"github.com/NicoNex/echotron/v3"
)

func SendVPNToTopic(chatID int64, topicID int) {
	var (
		b = &bot{
			API: echotron.NewAPI(botToken),
		}

		bug     = []string{"BUG.COM"}
		account = A.PopulateBugs(A.Get("ORDER BY RANDOM() LIMIT 1"), bug, bug)[0]
		message = helper.MakeVPNMessage(account)
	)

	go b.SendMessage(message, chatID, &echotron.MessageOptions{
		ParseMode:        "HTML",
		ReplyToMessageID: int(topicID),
	})
}
