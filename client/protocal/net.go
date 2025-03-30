package protocal

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net"
)

func AuthC(conn *net.Conn, secret string, id Identity) error {
	// TODO: implement authentication
	authMsg := []byte{byte(Header_AUTH)}
	secretHash := sha256.Sum256([]byte(secret))
	authMsg = append(authMsg, secretHash[:]...)
	authMsg = append(authMsg, byte(id))
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

func AuthS(conn *net.Conn, secret string) (Identity, error) {
	// TODO: implement authentication
	buf := make([]byte, 34)
	_, err := (*conn).Read(buf)
	if err != nil {
		return 0, err
	}
	if buf[0] != byte(Header_AUTH) {
		return 0, ErrAuthHeader
	}
	secretHash := sha256.Sum256([]byte(secret))
	if !bytes.Equal(secretHash[:], buf[1:33]) {
		return 0, ErrAuthFailed
	}
	id := Identity(buf[33])
	authMsg := []byte{byte(Header_AUTH_R)}
	_, err = (*conn).Write(authMsg)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func NetConn(tcp *net.Conn, O *net.Conn) error {
	// TODO: implement net copy
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
