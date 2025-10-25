package data

import (
	"log"

	"mchat/pkg/pop3"
)

type Message struct {
	Subject string
	Content string
}

func GetChats() []Message {
	messages := []Message{}
	p := pop3.New("localhost", "1110")
	conn, err := p.Conn()
	if err != nil {
		log.Fatal(err)
	}
	if err := conn.Auth("odoo", "odoo"); err != nil {
		log.Println(err)
	}
	defer conn.Quit()

	return messages
}
