package data

import (
	"io"
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

	msginfos, err := conn.List()
	if err != nil {
		log.Println(err)
	}

	for _, m := range msginfos {
		log.Printf("Retrieving msg %d of size %d ready\n", m.Id, m.Size)
		msg, err := conn.Retr(m.Id)
		if err != nil {
			log.Println(err)
		} else {
			log.Println(msg.Header.AddressList("From"))
			bs, _ := io.ReadAll(msg.Body)
			// content comes after content type \n\n, need to decode b64 of the content only
			log.Println(string(bs))
		}
	}

	return messages
}
