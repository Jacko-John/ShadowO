package client

import (
	"crypto/sha256"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/lxzan/gws"
)

func (c *Client) DialServer(id ClientID) (*net.Conn, error) {
	// TODO: implement this
	// dialer := websocket.Dialer{
	// 	TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	// }
	socket, _, err := gws.NewClient(&gws.BuiltinEventHandler{}, &gws.ClientOption{
		TlsConfig:       &tls.Config{InsecureSkipVerify: true},
		Addr:            c.serverUrl,
		ParallelEnabled: true,
	})
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return nil, err
	}
	// fmt.Println(socket)
	secret := sha256.Sum256([]byte(c.authSecret))
	conn := socket.NetConn()
	// fmt.Println(conn)
	auth := append(secret[:], byte(id))
	conn.Write(auth)
	buf := make([]byte, len("Authentication success"))
	conn.Read(buf)
	fmt.Println(string(buf))
	if string(buf) != "Authentication success" {
		fmt.Println("Error: Authentication failed")
		return nil, err
	}
	// fmt.Println(conn)
	if id == Master {
		c.conn = &conn
	} else {
		c.tunnel = &conn
	}
	return &conn, nil
}
