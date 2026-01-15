package data

import (
	"cmp"
	"encoding/base64"
	"io"
	"log"
	"mime"
	"mime/multipart"
	"net/mail"
	"slices"
	"strings"

	"mchat/internal/config"
	"mchat/pkg/pop3"
)

type Message struct {
	From    mail.Address
	Date    string
	Content string
}

type Chat struct {
	Contact  string // to be replaced by a name + email address ?
	Messages []*Message
}

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

func processMessage(msg *mail.Message) *Message {
	fromList, _ := msg.Header.AddressList("From")
	var address mail.Address
	if len(fromList) > 0 {
		address = *fromList[0]
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

	return &Message{
		From:    address,
		Date:    dateStr,
		Content: content,
	}
}

func getMessages(c *config.Config) []*Message {
	var p pop3.Pop3
	var conn pop3.Connection
	var err error

	if c.Google {
		p = pop3.New("pop.gmail.com", "995")
		conn, err = p.Conn(true)
	} else {
		p = pop3.New("localhost", "1110")
		conn, err = p.Conn(false)
	}
	if err != nil {
		log.Fatal(err)
	}

	if c.Google {
		err = conn.XOAuth2(c.Login, c.AccessToken)
	} else {
		err = conn.Auth("odoo", "odoo")
	}

	if err != nil {
		log.Fatal(err)
	}
	defer conn.Quit()

	msginfos, err := conn.List()
	if err != nil {
		log.Fatal(err)
	}

	var messages []*Message
	for _, m := range msginfos {
		log.Printf("Retrieving msg %d of size %d ready\n", m.Id, m.Size)
		msg, err := conn.Retr(m.Id)
		if err != nil {
			log.Println(err)
		} else {
			messages = append(messages, processMessage(msg))
		}
	}

	return messages
}

func GetChats(c *config.Config) []*Chat {
	msgs := getMessages(c)
	chats := make(map[string]*Chat)

	for _, msg := range msgs {
		from := msg.From.String()
		chat, ok := chats[from]
		if ok {
			chat.Messages = append(chat.Messages, msg)
		} else {
			chats[from] = &Chat{
				Contact:  from,
				Messages: []*Message{msg},
			}
		}
	}

	var ordChats []*Chat
	for _, chat := range chats {
		ordChats = append(ordChats, chat)
		slices.SortFunc(chat.Messages, func(a, b *Message) int {
			return cmp.Compare(a.Date, b.Date)
		})
	}

	return ordChats
}
