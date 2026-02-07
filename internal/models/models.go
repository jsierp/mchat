package models

import "time"

type Message struct {
	Id          string
	From        string
	To          string
	Contact     string
	ChatAddress string
	Content     string
	Date        time.Time
}

type Chat struct {
	Address  string
	Name     string
	Messages []*Message
}
