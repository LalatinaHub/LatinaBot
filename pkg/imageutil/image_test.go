package imageutil

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateLocalOrderImage(t *testing.T) {
	data, err := GenerateLocalOrderImage("12345678-1234-1234-1234-123456789abc")
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	img, format, err := image.Decode(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, "jpeg", format)
	assert.Equal(t, 1200, img.Bounds().Dx())
	assert.Equal(t, 200, img.Bounds().Dy())
}

func TestGenerateQRCodePNG(t *testing.T) {
	data, err := GenerateQRCodePNG("https://foolvpn.web.id", 256)
	assert.NoError(t, err)
	assert.NotEmpty(t, data)

	img, format, err := image.Decode(bytes.NewReader(data))
	assert.NoError(t, err)
	assert.Equal(t, "png", format)
	assert.Equal(t, 256, img.Bounds().Dx())
}
