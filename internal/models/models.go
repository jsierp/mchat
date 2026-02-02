package models

import "net/mail"

type Message struct {
	Id      string
	Contact *mail.Address
	Content string
	Date    string
}

type Chat struct {
	Contact  *mail.Address // TODO do not use a pointer
	Messages []*Message    // TODO use slice of structs
}
