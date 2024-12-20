package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/skip2/go-qrcode"
	"io"
	"mime/multipart"
	"net/http"
	"os"
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

func SendFileForAnalysis(file multipart.File) (map[string]interface{}, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", "menu.jpg")
	if err != nil {
		return nil, err
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", os.Getenv("MENU_ANALYZER_URL"), body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
