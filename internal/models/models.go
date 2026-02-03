package models

type Message struct {
	Id          string
	From        string
	To          string
	Contact     string
	ChatAddress string
	Content     string
	Date        string
}

type Chat struct {
	Address  string
	Name     string
	Messages []*Message
}
