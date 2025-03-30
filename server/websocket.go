package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"net"
)

func (s *Server) handleClient(conn *AuthConn) {

	// 读取客户端认证信息
	buf := make([]byte, 33)
	conn.Read(buf)
	secretHash := sha256.Sum256([]byte(s.authSecret))
	// 验证密码
	if bytes.Equal(secretHash[:], buf[:32]) {
		n, err := conn.Write([]byte("Authentication success"))
		fmt.Println("write:", n, err)
		conn.Authed = true
		s.pool.Add(conn)
		log.Printf("Client authenticated: %s", conn.RemoteAddr())
		return
	}
	log.Printf("Client authentication failed: %s", conn.RemoteAddr())
	conn.Close()
	// 转发到nginx
	// log.Println("Forwarding to nginx")
	// nginxConn, err := net.Dial("tcp", "localhost:80")
	// if err != nil {
	// 	log.Println("Nginx connect error:", err)
	// 	return
	// }
	// defer nginxConn.Close()
	// // 转发初始数据
	// nginxConn.Write(buf)
	// _, reader, _ := conn.NextReader()
	// writer, _ := conn.NextWriter(websocket.TextMessage)
	// go io.Copy(nginxConn, reader)
	// io.Copy(writer, nginxConn)

}

func (s *Server) connectWebsocket(tcp net.Conn) {
	defer tcp.Close()
	conn := s.pool.Get()
	if conn == nil {
		return
	}
	// defer s.pool.Remove(conn)
	conn.Write([]byte("new"))
	log.Println("wrote")
	buf := make([]byte, len("Ok..."))
	conn.Read(buf)
	fmt.Println(string(buf))
	if string(buf) == "Ok..." {
		log.Println("Websocket connection copying data")
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
		conn.Write([]byte("EOF"))
	}
	fmt.Println("Nothing to do")
}
