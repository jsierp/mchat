package pop3

import (
	"bufio"
	"bytes"
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
	conn net.Conn
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

func (p *Pop3) Conn() (Connection, error) {
	log.Println("Initializing connection")
	conn, err := net.Dial("tcp", net.JoinHostPort(p.host, p.port))
	c := Connection{conn: conn}
	if err != nil {
		return c, err
	}
	c.SetDeadline()
	reader := bufio.NewReader(conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return c, err
	}
	if !strings.HasPrefix(msg, "+OK") {
		return c, errors.New("Connection failed")
	}
	log.Print(msg)
	return c, err
}

func (c *Connection) SetDeadline() {
	c.conn.SetDeadline(time.Now().Add(TIMEOUT))
}

func (c *Connection) Auth(user string, pass string) error {
	log.Println("Sending User")
	c.SetDeadline()
	_, err := fmt.Fprintf(c.conn, "USER %s\r\n", user)
	if err != nil {
		return err
	}
	c.SetDeadline()
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	log.Print(msg)
	if !strings.HasPrefix(msg, "+OK") {
		return errors.New("USER command failed")
	}

	log.Println("Sending password")
	c.SetDeadline()
	_, err = fmt.Fprintf(c.conn, "PASS %s\r\n", pass)
	if err != nil {
		return err
	}
	c.SetDeadline()
	msg, err = reader.ReadString('\n')
	if err != nil {
		return err
	}
	log.Print(msg)
	if !strings.HasPrefix(msg, "+OK") {
		return errors.New("PASS command failed")
	}
	return nil
}

func (c *Connection) Quit() error {
	defer c.conn.Close()
	log.Println("Quitting")
	c.SetDeadline()
	_, err := fmt.Fprint(c.conn, "QUIT\r\n")
	if err != nil {
		return err
	}

	c.SetDeadline()
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return err
	}
	log.Print(msg)
	if !strings.HasPrefix(msg, "+OK") {
		return errors.New("QUIT command failed")
	}

	return nil
}

func (c *Connection) List() ([]MsgInfo, error) {
	c.SetDeadline()
	_, err := fmt.Fprint(c.conn, "LIST\r\n")
	if err != nil {
		return nil, err
	}

	c.SetDeadline()
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Print(msg)
	if !strings.HasPrefix(msg, "+OK") {
		return nil, errors.New("LIST command failed")
	}

	var n int
	_, err = fmt.Sscanf(msg, "+OK %d", &n)
	if err != nil {
		return nil, errors.New("LIST size failed")
	}

	msgs := []MsgInfo{}
	for range n {
		msg, err := reader.ReadString('\n')
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
	reader := bufio.NewReader(c.conn)
	msg, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}
	log.Print(msg)
	if !strings.HasPrefix(msg, "+OK") {
		return nil, errors.New("RETR failed")
	}

	var buf bytes.Buffer

	for {
		msg, err = reader.ReadString('\n')
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
