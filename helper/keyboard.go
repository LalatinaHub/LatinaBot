package helper

import (
	"strconv"
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

func BuildInlineKeyboardWithPage(values []string, page int) [][]echotron.InlineKeyboardButton {
	var (
		InlineKeyboard   [][]echotron.InlineKeyboardButton = [][]echotron.InlineKeyboardButton{{}}
		dimmension       int
		dimmensionButton int = 2
		pageButton       int = 8
	)

	if page < 0 {
		page = 0
	}

	newValues := values[pageButton*page:]
	for i, value := range newValues {
		if i >= pageButton {
			break
		}

		InlineKeyboard[dimmension] = append(InlineKeyboard[dimmension], echotron.InlineKeyboardButton{
			Text:         value,
			CallbackData: strings.ReplaceAll(value, " ", "_"),
		})

		if len(InlineKeyboard[dimmension]) >= dimmensionButton {
			dimmension++
			InlineKeyboard = append(InlineKeyboard, []echotron.InlineKeyboardButton{})
		}
	}

	InlineKeyboard = append(InlineKeyboard, []echotron.InlineKeyboardButton{})
	if page > 0 {
		InlineKeyboard[dimmension+1] = append(InlineKeyboard[dimmension+1], echotron.InlineKeyboardButton{
			Text:         "<< Prev",
			CallbackData: strconv.Itoa(page - 1),
		})
	}

	if page < len(values)/pageButton {
		InlineKeyboard[dimmension+1] = append(InlineKeyboard[dimmension+1], echotron.InlineKeyboardButton{
			Text:         "Next >>",
			CallbackData: strconv.Itoa(page + 1),
		})
	}

	return InlineKeyboard
}
