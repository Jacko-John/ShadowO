package server

import (
	"log"
	"net/http"

	"github.com/lxzan/gws"
)

var upgrader = gws.NewUpgrader(nil, &gws.ServerOption{
	ParallelEnabled:   true,                                 // Parallel message processing
	Recovery:          gws.Recovery,                         // Exception recovery
	PermessageDeflate: gws.PermessageDeflate{Enabled: true}, // Enable compression
})

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	// 将HTTP连接升级为WebSocket连接
	conn, err := upgrader.Upgrade(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	// 处理WebSocket连接
	s.handleClient(&AuthConn{Conn: conn.NetConn()})
}
