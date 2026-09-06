package ocr

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractOrderID(t *testing.T) {
	scanner := &VisionScanner{}

	t.Run("Valid single UUID", func(t *testing.T) {
		text := "Donasi Berhasil!\nID Transaksi: 12345678-abcd-1234-abcd-1234567890ab\nTerima kasih!"
		orderID := scanner.ExtractOrderID(text)
		assert.Equal(t, "12345678-abcd-1234-abcd-1234567890ab", orderID)
	})

	t.Run("Multiple UUIDs returns last", func(t *testing.T) {
		text := "Prev: 00000000-0000-0000-0000-000000000000\nOrder: aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
		orderID := scanner.ExtractOrderID(text)
		assert.Equal(t, "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee", orderID)
	})

	t.Run("No UUID", func(t *testing.T) {
		text := "Pembayaran sukses tanpa order id."
		orderID := scanner.ExtractOrderID(text)
		assert.Empty(t, orderID)
	})
}
