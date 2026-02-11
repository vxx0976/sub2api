package payment

import (
	"encoding/base64"
	"fmt"

	qrcode "github.com/skip2/go-qrcode"
)

// GenerateQRCodeBase64 generates a QR code as base64-encoded PNG
func GenerateQRCodeBase64(content string, size int) (string, error) {
	if size <= 0 {
		size = 256
	}

	png, err := qrcode.Encode(content, qrcode.Medium, size)
	if err != nil {
		return "", fmt.Errorf("generate QR code: %w", err)
	}

	return base64.StdEncoding.EncodeToString(png), nil
}
