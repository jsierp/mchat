package pop3

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net"
	"net/mail"
	"strings"
	"time"
)

var TIMEOUT time.Duration = 5 * time.Second

type Pop3 struct {
	host string
	port string
}

type Connection struct {
	conn   net.Conn
	reader *bufio.Reader
}

type MsgInfo struct {
	Id   int
	Size int
}

func New(host string, port string) Pop3 {
	return Pop3{
		host: host,
		port: port,
	}
}

func (p *Pop3) Conn(doTls bool) (*Connection, error) {
	log.Println("Initializing connection")
	var conn net.Conn
	var err error
	if doTls {
		conn, err = tls.Dial("tcp", net.JoinHostPort(p.host, p.port), nil)
	} else {
		conn, err = net.Dial("tcp", net.JoinHostPort(p.host, p.port))
	}
	reader := bufio.NewReader(conn)
	c := &Connection{conn: conn, reader: reader}
	if err != nil {
		return c, err
	}
	c.SetDeadline()

	if _, err := c.checkResponseOK(); err != nil {
		return nil, err
	}
	return c, err
}

func (c *Connection) SetDeadline() {
	c.conn.SetDeadline(time.Now().Add(TIMEOUT))
}

func (c *Connection) checkResponseOK() (string, error) {
	msg, err := c.reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if !strings.HasPrefix(msg, "+OK") {
		return msg, errors.New(msg)
	}
	return msg, nil
}

func GetXOAuth2String(user, token string) string {
	str := fmt.Sprintf("user=%s\x01auth=Bearer %s\x01\x01", user, token)
	return base64.StdEncoding.EncodeToString([]byte(str))
}

func (c *Connection) XOAuth2(user, token string) error {
	c.SetDeadline()
	if _, err := fmt.Fprintf(c.conn, "AUTH XOAUTH2\r\n"); err != nil {
		return err
	}
	msg, err := c.reader.ReadString('\n')
	if err != nil {
		return err
	}
	if !strings.HasPrefix(msg, "+") {
		return errors.New(msg)
	}

	authStr := GetXOAuth2String(user, token)
	log.Println("Sending XOAuth2 String")
	if _, err := fmt.Fprintf(c.conn, "%s\r\n", authStr); err != nil {
		return err
	}
	if _, err := c.checkResponseOK(); err != nil {
		return err
	}
	return nil
}

func (c *Connection) Auth(user string, pass string) error {
	c.SetDeadline()

	log.Println("Sending User")
	_, err := fmt.Fprintf(c.conn, "USER %s\r\n", user)
	if err != nil {
		return err
	}
	if _, err := c.checkResponseOK(); err != nil {
		return err
	}

	log.Println("Sending password")
	_, err = fmt.Fprintf(c.conn, "PASS %s\r\n", pass)
	if err != nil {
		return err
	}
	if _, err := c.checkResponseOK(); err != nil {
		return err
	}
	return nil
}

func (c *Connection) Quit() error {
	c.SetDeadline()
	defer c.conn.Close()

	log.Println("Quitting")
	_, err := fmt.Fprint(c.conn, "QUIT\r\n")
	if err != nil {
		return err
	}
	if _, err := c.checkResponseOK(); err != nil {
		return err
	}
	return nil
}

func (c *Connection) List() ([]MsgInfo, error) {
	c.SetDeadline()
	_, err := fmt.Fprint(c.conn, "LIST\r\n")
	if err != nil {
		return nil, err
	}
	var msg string
	if msg, err = c.checkResponseOK(); err != nil {
		return nil, err
	}

	var n int
	_, err = fmt.Sscanf(msg, "+OK %d", &n)
	if err != nil {
		return nil, errors.New("LIST size failed")
	}

	msgs := []MsgInfo{}
	for range n {
		msg, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}

		msginfo := MsgInfo{}
		_, err = fmt.Sscanf(msg, "%d %d", &msginfo.Id, &msginfo.Size)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msginfo)
	}

	return msgs, nil
}

func (c *Connection) Retr(id int) (*mail.Message, error) {
	c.SetDeadline()
	_, err := fmt.Fprintf(c.conn, "RETR %d\r\n", id)
	if err != nil {
		return nil, err
	}
	if _, err := c.checkResponseOK(); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	for {
		msg, err := c.reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		if msg == ".\r\n" {
			break
		}
		buf.WriteString(msg)
	}

	return mail.ReadMessage(bytes.NewReader(buf.Bytes()))
}
