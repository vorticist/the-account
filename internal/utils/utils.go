package utils

import (
	"encoding/base64"
	"github.com/skip2/go-qrcode"
)

func GenerateQRCodeBase64(content string) (string, error) {
	// Generate the QR code as PNG data
	pngData, err := qrcode.Encode(content, qrcode.Medium, 256)
	if err != nil {
		return "", err
	}

	// Encode the PNG data as a Base64 string
	base64Data := base64.StdEncoding.EncodeToString(pngData)

	// Return the data with a proper data URI scheme for embedding in HTML or other contexts
	return base64Data, nil
}
