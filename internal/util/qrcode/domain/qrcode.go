package domain

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image/png"

	"github.com/skip2/go-qrcode"
)

type QRCodeGenerator interface {
	GenerateQRCode(data string) (string, error)
}

type qrCodeGenerator struct {
	size int
}

func NewQRCodeGenerator(size int) QRCodeGenerator {
	return &qrCodeGenerator{
		size: size,
	}
}

func (g *qrCodeGenerator) GenerateQRCode(data string) (string, error) {
	qr, err := qrcode.New(data, qrcode.Medium)
	if err != nil {
		return "", &QRCodeGenerationError{Message: err.Error()}
	}

	qr.DisableBorder = false
	img := qr.Image(g.size)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", &QRCodeGenerationError{Message: err.Error()}
	}

	base64Str := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("data:image/png;base64,%s", base64Str), nil
}
