package pop3

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
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
