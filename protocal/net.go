package protocal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"net"
)

func AuthC(conn *net.Conn, secret string, port int32, id Identity) error {
	// TODO: implement authentication
	authMsg := make([]byte, 38)
	authMsg[0] = byte(Header_AUTH)
	secretHash := sha256.Sum256([]byte(secret))
	copy(authMsg[1:33], secretHash[:])
	portBytes := authMsg[33:37]
	portBytes[0] = byte(port >> 24)
	portBytes[1] = byte(port >> 16)
	portBytes[2] = byte(port >> 8)
	portBytes[3] = byte(port)
	authMsg[37] = byte(id)
	_, err := (*conn).Write(authMsg)
	if err != nil {
		return err
	}
	buf := make([]byte, 1)
	_, err = (*conn).Read(buf)
	if err != nil {
		return err
	}
	if buf[0] != byte(Header_AUTH_R) {
		return ErrAuthFailed
	}
	return nil
}

func AuthS(conn *net.Conn, secret string) (int32, Identity, error) {
	// TODO: implement authentication
	buf := make([]byte, 38)
	_, err := (*conn).Read(buf)
	if err != nil {
		return 0, 0, err
	}
	if buf[0] != byte(Header_AUTH) {
		return 0, 0, ErrAuthHeader
	}
	secretHash := sha256.Sum256([]byte(secret))
	if !bytes.Equal(secretHash[:], buf[1:33]) {
		fmt.Printf("c: %x\n", buf[1:33])
		fmt.Printf("s: %x\n", secretHash[:])
		return 0, 0, ErrAuthFailed
	}
	port := int32(buf[33])<<24 | int32(buf[34])<<16 | int32(buf[35])<<8 | int32(buf[36])
	id := Identity(buf[37])
	authMsg := []byte{byte(Header_AUTH_R)}
	_, err = (*conn).Write(authMsg)
	if err != nil {
		return 0, 0, err
	}
	return port, id, nil
}

func NetConn(tcp *net.Conn, O *net.Conn) error {
	// TODO: implement net copy
	defer (*tcp).Close()
	var err [2]error
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		_, err[0] = netWriteToO(*tcp, *O)
		cancel()
	}()
	go func() {
		_, err[1] = netReadFromO(*O, *tcp)
		cancel()
	}()
	<-ctx.Done()
	if err[0] == nil {
		return err[1]
	}
	return err[0]
}

func HeaderPing(conn *net.Conn) (Header, error) {
	// TODO: implement ping
	pingMsg := []byte{byte(Header_PING)}
	_, err := (*conn).Write(pingMsg)
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 1)
	_, err = (*conn).Read(buf)
	if err != nil {
		return 0, err
	}
	if Header(buf[0]) != Header_PONG {
		return Header(buf[0]), ErrPong
	}
	return Header(buf[0]), nil
}

func HeaderPong(conn *net.Conn) (Header, error) {
	// TODO: implement pong
	buf := make([]byte, 1)
	_, err := (*conn).Read(buf)
	if err != nil {
		return 0, err
	}
	if Header(buf[0]) != Header_PING {
		return Header(buf[0]), ErrPing
	}
	_, err = (*conn).Write([]byte{byte(Header_PONG)})
	if err != nil {
		return 0, err
	}
	return Header(buf[0]), nil
}

func HeaderNew(Conn *net.Conn) (Header, error) {
	// TODO: implement new header
	_, err := (*Conn).Write([]byte{byte(Header_NEW)})
	if err != nil {
		return 0, err
	}
	buf := make([]byte, 1)
	_, err = (*Conn).Read(buf)
	if err != nil {
		return 0, err
	}
	if Header(buf[0]) != Header_GOT {
		return Header(buf[0]), ErrGot
	}
	return Header(buf[0]), nil
}

func HeaderGot(Conn *net.Conn) (Header, error) {
	// TODO: implement got header
	buf := make([]byte, 1)
	_, err := (*Conn).Read(buf)
	if err != nil {
		return 0, err
	}
	if Header(buf[0]) != Header_NEW {
		return Header(buf[0]), ErrNew
	}
	_, err = (*Conn).Write([]byte{byte(Header_GOT)})
	if err != nil {
		return 0, err
	}
	return Header(buf[0]), nil
}
