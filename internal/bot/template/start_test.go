package template

import (
	"testing"
	"time"

	"github.com/LalatinaHub/LatinaBot/internal/service/tarot"
	"github.com/LalatinaHub/LatinaBot/internal/service/trakteer"
	"github.com/LalatinaHub/common/model"
	"github.com/stretchr/testify/assert"
	tele "gopkg.in/telebot.v3"
)

func TestBuildStartMessage(t *testing.T) {
	sender := &tele.User{ID: 123456, FirstName: "Alice"}
	user := &model.User{
		ID:         123456,
		Token:      "tok12345",
		Password:   "uuid-pass",
		Expired:    time.Now().Add(24 * time.Hour),
		ServerCode: "sg1",
		Quota:      10000,
		Relay:      "SG",
		Adblock:    true,
		VPN:        "vless",
	}
	card := tarot.Card{Name: "The Magician", Message: "create a new reality"}
	donations := &trakteer.Response{}
	donations.Result.Data = []trakteer.SupportItem{
		{SupporterName: "Bob", Quantity: 5, SupportMessage: "Keren!"},
	}
	servers := []model.Server{
		{Code: "sg1", Domain: "sg1.foolvpn.me"},
	}

	msg := BuildStartMessage(sender, user, card, donations, servers)

	assert.Contains(t, msg, "The Magician on Service")
	assert.Contains(t, msg, "create a new reality")
	assert.Contains(t, msg, "ID: <code>123456</code>")
	assert.Contains(t, msg, "Nama: Alice")
	assert.Contains(t, msg, "Password: <code>tok12345</code>")
	assert.Contains(t, msg, "Status: <b>Donator</b>")
	assert.Contains(t, msg, "Tipe: vless")
	assert.Contains(t, msg, "Server Code: sg1")
	assert.Contains(t, msg, "Domain: sg1.foolvpn.me")
	assert.Contains(t, msg, "Bob [5]")
}

func TestBuildStartKeyboard(t *testing.T) {
	user := &model.User{
		Token:   "secret-token",
		Adblock: true,
	}

	markup := BuildStartKeyboard(user)
	assert.NotNil(t, markup)
	assert.NotEmpty(t, markup.InlineKeyboard)

	// Verify all data buttons have valid Telebot identifiers (no slashes)
	for _, row := range markup.InlineKeyboard {
		for _, btn := range row {
			if btn.Unique != "" {
				assert.NotContains(t, btn.Unique, "/", "Button unique %q should not contain '/'", btn.Unique)
			}
		}
	}
}
