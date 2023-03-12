package helper

import (
	"strings"

	"github.com/NicoNex/echotron/v3"
)

func BuildInlineKeyboard(values []string) [][]echotron.InlineKeyboardButton {
	var (
		InlineKeyboard [][]echotron.InlineKeyboardButton = [][]echotron.InlineKeyboardButton{{}}
		dimmension     int
	)

	for _, value := range values {
		InlineKeyboard[dimmension] = append(InlineKeyboard[dimmension], echotron.InlineKeyboardButton{
			Text:         value,
			CallbackData: strings.ReplaceAll(value, " ", "_"),
		})

		if len(InlineKeyboard[dimmension]) > 1 {
			dimmension++
			InlineKeyboard = append(InlineKeyboard, []echotron.InlineKeyboardButton{})
		}
	}

	return InlineKeyboard
}
