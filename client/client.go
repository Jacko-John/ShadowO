package client

import (
	"log"
	"net"
)

type ClientID byte

const (
	Master ClientID = ClientID(0)
	Tunnel ClientID = ClientID(1)
)

type Client struct {
	authSecret string
	serverUrl  string
	localUrl   string
	conn       *net.Conn
	tunnel     *net.Conn
}

func NewClient(authSecret, serverUrl, localUrl string) *Client {
	return &Client{
		authSecret: authSecret,
		serverUrl:  serverUrl,
		localUrl:   localUrl,
	}
}

func (c *Client) Run() {

	conn, err := c.DialServer(Tunnel)
	if err != nil {
		log.Println(err)
		return
	}
	// conn := *(c.tunnel)
	// log.Println("Connected to server")
	// log.Println(conn)
	for {
		buf := make([]byte, len("new"))
		(*conn).Read(buf)
		log.Println(string(buf))
		if string(buf) != "new" {
			continue
		}
		(*conn).Write([]byte("Ok..."))
		c.DialLocal()
	}
}
