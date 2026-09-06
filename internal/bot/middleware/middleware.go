package middleware

import (
	"fmt"
	"html"
	"runtime/debug"

	"github.com/LalatinaHub/LatinaBot/pkg/logger"
	tele "gopkg.in/telebot.v3"
)

// Typing sends typing chat action.
func Typing() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			_ = c.Notify(tele.Typing)
			return next(c)
		}
	}
}

// PrivateOnly filters updates to private chats only (except for admin/threads).
func PrivateOnly() tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Chat() != nil && c.Chat().Type != tele.ChatPrivate {
				return nil
			}
			return next(c)
		}
	}
}

// AdminOnly restricts handler to admin user.
func AdminOnly(adminID int64) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			if c.Sender() == nil || c.Sender().ID != adminID {
				return nil
			}
			return next(c)
		}
	}
}

// ErrorReporter catches errors and panics, notifies the admin, and responds to the user.
func ErrorReporter(bot *tele.Bot, adminID int64) tele.MiddlewareFunc {
	return func(next tele.HandlerFunc) tele.HandlerFunc {
		return func(c tele.Context) error {
			defer func() {
				if r := recover(); r != nil {
					stack := string(debug.Stack())
					logger.Error().Interface("panic", r).Str("stack", stack).Msg("Recovered from panic in bot handler")

					sendErrorToAdminAndUser(bot, c, adminID, fmt.Sprintf("%v", r), stack)
				}
			}()

			err := next(c)
			if err != nil {
				logger.Error().Err(err).Msg("Error processing update")
				sendErrorToAdminAndUser(bot, c, adminID, err.Error(), "")
			}
			return err
		}
	}
}

func sendErrorToAdminAndUser(bot *tele.Bot, c tele.Context, adminID int64, errMsg, stack string) {
	// 1. Reply to user
	menu := &tele.ReplyMarkup{}
	btnUptime := menu.URL("Cek Status Server", "https://foolvpn.me/uptime")
	menu.Inline(menu.Row(btnUptime))

	userMsg := fmt.Sprintf(`Waduh error!
Coba bilang <a href="tg://user?id=%d">admin</a>`, adminID)

	_ = c.Reply(userMsg, menu, tele.ModeHTML)

	// 2. Alert Admin
	if adminID != 0 && bot != nil {
		senderID := int64(0)
		if c.Sender() != nil {
			senderID = c.Sender().ID
		}
		var incomingMsg string
		if c.Message() != nil {
			incomingMsg = c.Message().Text
		}

		adminReport := fmt.Sprintf("<b>Error Report</b>\n"+
			"From: %d\n"+
			"Message: %s\n"+
			"Error Message: %s\n"+
			"Error Stack:\n<code>%s</code>",
			senderID,
			html.EscapeString(incomingMsg),
			html.EscapeString(errMsg),
			html.EscapeString(stack),
		)

		_, _ = bot.Send(&tele.Chat{ID: adminID}, adminReport, tele.ModeHTML)
	}
}
