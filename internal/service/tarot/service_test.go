package tarot

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTarotService_GetCard(t *testing.T) {
	svc := NewService()
	assert.NotEmpty(t, svc.cards)

	for i := 0; i < 20; i++ {
		card := svc.GetCard()
		assert.NotEmpty(t, card.Name)
		assert.NotEmpty(t, card.Message)
		assert.True(t, strings.HasPrefix(card.Image, baseImageURL))
	}
}
