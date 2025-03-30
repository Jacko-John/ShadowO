package client

import (
	"context"
	"io"
	"log"
	"net"
)

func (c *Client) DialLocal() {
	log.Println("Dialing local")
	tcp, err := net.Dial("tcp", c.localUrl)
	if err != nil {
		panic(err)
	}
	defer tcp.Close()
	log.Println("Connected to local")
	conn := *c.tunnel
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_, _ = io.Copy(conn, tcp)
		cancel()
	}()
	go func() {
		_, _ = io.Copy(tcp, conn)
		cancel()
	}()
	<-ctx.Done()
	log.Println("Local connection closed")
}
