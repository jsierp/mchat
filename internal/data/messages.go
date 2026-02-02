package data

import (
	"encoding/base64"
	"io"
	"mchat/internal/models"
	"mime"
	"mime/multipart"
	"net/mail"
	"strings"
)

func getPlainText(msg *mail.Message) (string, error) {
	contentType := msg.Header.Get("Content-Type")
	if contentType == "" {
		body, _ := io.ReadAll(msg.Body)
		return string(body), nil
	}
	return parsePart(msg.Body, contentType, msg.Header.Get("Content-Transfer-Encoding"))
}

func parsePart(body io.Reader, contentType string, encoding string) (string, error) {
	mediaType, params, err := mime.ParseMediaType(contentType)
	if err != nil {
		return "", err
	}

	if mediaType == "text/plain" {
		if encoding == "base64" {
			reader := base64.NewDecoder(base64.StdEncoding, body)
			content, _ := io.ReadAll(reader)
			return string(content), nil
		}
		content, _ := io.ReadAll(body)
		return string(content), nil
	}

	if strings.HasPrefix(mediaType, "multipart/") {
		mr := multipart.NewReader(body, params["boundary"])
		for {
			p, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				continue
			}
			result, _ := parsePart(p, p.Header.Get("Content-Type"), p.Header.Get("Content-Transfer-Encoding"))
			if result != "" {
				return result, nil
			}
		}
	}

	return "", nil
}

func processMessage(msg *mail.Message) *models.Message {
	fromList, _ := msg.Header.AddressList("From")
	var address *mail.Address
	if len(fromList) > 0 {
		address = fromList[0]
	}
	content, err := getPlainText(msg)
	if err != nil {
		content = ""
	}
	date, err := msg.Header.Date()
	dateStr := "error"
	if err == nil {
		dateStr = date.Format("2006-01-02 15:04")
	}
	id := msg.Header.Get("Message-ID")

	return &models.Message{
		Id:      id,
		Contact: address,
		Content: content,
		Date:    dateStr,
	}
}
