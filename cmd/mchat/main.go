package main

import (
	"fmt"
	"io"
	"log"

	"github.com/knadh/go-pop3"
)

func main() {
	p := pop3.New(pop3.Opt{
		Host: "localhost",
		Port: 1110,
	})
	c, err := p.NewConn()
	if err != nil {
		log.Fatal("CONN ", err)
	}
	defer c.Quit()

	if err := c.Auth("odoo", "odoo"); err != nil {
		log.Fatal("AUTH ", err)
	}

	count, size, err := c.Stat()
	if err != nil {
		log.Fatal("STAT ", err)
	}
	fmt.Println("count: ", count, ", size:", size)

	msgs, _ := c.List(0)
	for _, mid := range msgs {
		m, _ := c.Retr(mid.ID)
		fmt.Println(m.Header.Get("Subject"))

		multipartReader := m.MultipartReader()
		for {
			nm, err := multipartReader.NextPart()
			if err != nil {
				log.Fatal(err)
			}
			content, _ := io.ReadAll(nm.Body)
			fmt.Println("---STARTPART---")
			fmt.Println(string(content))
			fmt.Println("----ENDPART----")
		}
	}
}
