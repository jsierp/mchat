package data

import (
	"fmt"
	"io"
	"log"

	"github.com/knadh/go-pop3"
)

type Message struct {
	Subject string
	Content string
}

func GetData() []Message {
	messages := []Message{}
	p := pop3.New(pop3.Opt{
		Host: "localhost",
		Port: 1110,
	})
	c, err := p.NewConn()
	if err != nil {
		log.Fatal("CONN ", err)
	}
	defer c.Quit()

	if err := c.Auth("odoo", "odoo"); err != nil {
		log.Fatal("AUTH ", err)
	}

	count, size, err := c.Stat()
	if err != nil {
		log.Fatal("STAT ", err)
	}
	fmt.Println("Messages count: ", count, ", messages size:", size, "\n")

	msgs, _ := c.List(0)
	for _, mid := range msgs {
		m, _ := c.Retr(mid.ID)

		multipartReader := m.MultipartReader()
		// first part is plain text
		nm, err := multipartReader.NextPart()
		if err != nil {
			log.Fatal(err)
		}
		content, _ := io.ReadAll(nm.Body)
		// fmt.Println("Content", string(content))
		messages = append(messages, Message{
			Subject: m.Header.Get("Subject"),
			Content: string(content),
		})
	}

	return messages
}
