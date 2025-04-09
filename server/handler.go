package server

import (
	"ShadowO/protocal"
	"ShadowO/server/config"
	"ShadowO/server/pool"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/lxzan/gws"
)

var upgrader = gws.NewUpgrader(nil, &gws.ServerOption{
	ParallelEnabled:   true,                                 // Parallel message processing
	Recovery:          gws.Recovery,                         // Exception recovery
	PermessageDeflate: gws.PermessageDeflate{Enabled: true}, // Enable compression
})

func (s *Server) wsHandler(w http.ResponseWriter, r *http.Request) {
	// 将HTTP连接升级为WebSocket连接
	socket, err := upgrader.Upgrade(w, r)
	if err != nil {
		log.Println(err)
		return
	}
	// 处理WebSocket连接
	conn := socket.NetConn()
	cfg := config.Get()
	_port, _, err := protocal.AuthS(&conn, cfg.Secret)
	if err != nil {
		s.logger.Error(err.Error())
		conn.Close()
		return
	}
	port := strconv.Itoa(int(_port))
	s.logger.Info(fmt.Sprintf("New connection build %s <-> %s", conn.RemoteAddr().String(), port))
	pl, ok := s.pools[port]
	if !ok {
		newPool := pool.NewPool(port, s.logger)
		s.pools[port] = newPool
		pl = newPool
		go pl.Start()
	}
	t := pool.NewTunnel(conn, pl, s.logger)
	pl.Put(t)
}
