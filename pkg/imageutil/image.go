package imageutil

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"

	"github.com/skip2/go-qrcode"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

// GenerateLocalOrderImage creates a 1200x200 JPEG image with white background and orderID text.
func GenerateLocalOrderImage(orderID string) ([]byte, error) {
	const width = 1200
	const height = 200

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	white := color.RGBA{R: 255, G: 255, B: 255, A: 255}
	draw.Draw(img, img.Bounds(), &image.Uniform{C: white}, image.Point{}, draw.Src)

	// Draw text in center
	d := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{C: color.Black},
		Face: basicfont.Face7x13,
		Dot:  fixed.Point26_6{X: fixed.I(width/2 - len(orderID)*7/2), Y: fixed.I(height/2 + 6)},
	}
	d.DrawString(orderID)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GenerateQRCodePNG generates a PNG QR code byte array for the given content.
func GenerateQRCodePNG(content string, size int) ([]byte, error) {
	return qrcode.Encode(content, qrcode.Medium, size)
}

// BytesReader wraps a byte slice in a bytes.Reader.
func BytesReader(b []byte) *bytes.Reader {
	return bytes.NewReader(b)
}
