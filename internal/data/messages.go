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

func (s *DataService) processMessage(msg *mail.Message) *models.Message {
	fromList, _ := msg.Header.AddressList("From")
	var from *mail.Address
	if len(fromList) > 0 {
		from = fromList[0]
	}

	toList, _ := msg.Header.AddressList("To")
	if len(toList) == 0 {
		toList, _ = msg.Header.AddressList("Cc")
	}
	if len(toList) == 0 {
		toList, _ = msg.Header.AddressList("Bcc")
	}
	var to *mail.Address
	if len(toList) > 0 {
		to = toList[0]
	}

	var chatAddress string
	if msg.Header.Get("Delivered-To") != "" {
		chatAddress = from.Address
	} else {
		chatAddress = to.Address
	}

	content, err := getPlainText(msg)
	if err != nil {
		content = ""
	}
	date, err := msg.Header.Date()
	id := msg.Header.Get("X-MCHAT-ID")
	if id == "" {
		id = msg.Header.Get("Message-ID")
	}

	return &models.Message{
		Id:          id,
		Contact:     from.Name,
		ChatAddress: chatAddress,
		From:        from.Address,
		To:          to.Address,
		Content:     removeQuotedText(content),
		Date:        date,
	}
}

func removeQuotedText(s string) string {
	var stopAt int
	lines := strings.SplitAfter(s, "\n")

	for i, l := range lines {
		if len(l) > 3 && l[:3] == "___" {
			// Microsoft
			stopAt = i
			break
		}
		if l != "" && l[0] == '>' {
			// Google
			stopAt = max(0, i-3)
		}
	}
	if stopAt == 0 {
		return s
	}

	res := ""
	for _, l := range lines[:stopAt] {
		res += l
	}
	return strings.TrimSpace(res)
}
