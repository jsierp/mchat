package models

import "net/mail"

type Message struct {
	Contact *mail.Address
	Content string
	Date    string
}

type Chat struct {
	Contact  *mail.Address
	Messages []*Message
}
