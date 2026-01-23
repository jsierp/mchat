package oxsmtp

import (
	"fmt"
	"net/smtp"
)

type Auth struct {
	User  string
	Token string
}

func (a Auth) Start(server *smtp.ServerInfo) (string, []byte, error) {
	str := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", a.User, a.Token)
	return "XOAUTH2", []byte(str), nil
}

func (a Auth) Next(fromServer []byte, more bool) (toServer []byte, err error) {
	return nil, nil
}
