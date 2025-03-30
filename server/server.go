// server.go
package server

import (
	"log"
	"net"
	"net/http"
	"sync"
)

type Server struct {
	pool       *ConnectionPool
	authSecret string
	httpServer *http.Server
	mu         sync.Mutex
}

func NewServer(authSecret string) *Server {
	pool := &ConnectionPool{
		conns: make(map[string]*AuthConn),
	}
	return &Server{
		pool:       pool,
		authSecret: authSecret,
	}
}

func (s *Server) Listen() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/tunnel", s.wsHandler)
	s.httpServer = &http.Server{
		Addr:    ":8081",
		Handler: mux,
	}
	return s.httpServer.ListenAndServeTLS("cert.pem", "key.pem")
	// handle request
}

func (s *Server) ServeContent() error {
	tcpConn, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	for {
		conn, err := tcpConn.Accept()
		if err != nil {
			return err
		}
		log.Println("new connection from", conn.RemoteAddr())
		s.connectWebsocket(conn)
	}
}
