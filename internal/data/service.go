package data

import (
	"cmp"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"net/mail"
	"slices"
	"strings"
	"unicode"

	"mchat/pkg/pop3"
)

// GetChats
// NewChat
// SendMessage

type Message struct {
	From    mail.Address
	Date    string
	Content string
}

type Chat struct {
	Contact  string // to be replaced by a name + email address ?
	Messages []*Message
}

func getContent(msg []byte) (string, error) {
	// non-html content starts after double breakline
	parts := strings.Split(string(msg), "\r\n\r\n")
	if len(parts) <= 1 {
		return "", errors.New("No content found")
	}
	content, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return "", err
	}
	text := string(content)
	// everything after -- is regarded as a signature
	text = strings.Split(text, "--")[0]
	// cut first line, which is a 'preview'
	firstNL := strings.IndexByte(text, '\n')
	if firstNL != -1 {
		text = text[firstNL:]
	}
	return strings.TrimFunc(text, func(r rune) bool {
		return unicode.IsSpace(r) || !unicode.IsGraphic(r) || r == '\u034f'
	}), nil
}

func processMessage(msg *mail.Message) *Message {
	bs, _ := io.ReadAll(msg.Body)
	fromList, _ := msg.Header.AddressList("From")
	var address mail.Address
	if len(fromList) > 0 {
		address = *fromList[0]
	}
	content, err := getContent(bs)
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

func getMessages() []*Message {
	p := pop3.New("localhost", "1110")
	conn, err := p.Conn()
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Auth("odoo", "odoo"); err != nil {
		log.Println(err)
	}
	defer conn.Quit()

	msginfos, err := conn.List()
	if err != nil {
		log.Println(err)
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

func GetChats() []*Chat {
	msgs := getMessages()
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
