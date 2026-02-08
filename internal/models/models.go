package models

import "time"

type MsgStatus int

const (
	MsgStatusSuccess MsgStatus = iota
	MsgStatusSending
	MsgStatusError
)

type Message struct {
	Id          string
	From        string
	To          string
	Contact     string
	ChatAddress string
	Content     string
	Date        time.Time
	Status      MsgStatus
}

type Chat struct {
	Address  string
	Name     string
	Messages []*Message
}
