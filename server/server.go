// server.go
package server

import (
	"ShadowO/server/config"
	"ShadowO/server/pool"
	"flag"
	"log/slog"
	"net/http"
)

type Server struct {
	pools      map[string]*pool.Pool
	httpServer *http.Server
	logger     *slog.Logger
}

func NewServer() *Server {
	pool := make(map[string]*pool.Pool)
	return &Server{
		pools:  pool,
		logger: slog.Default(),
	}
}
func (s *Server) Run() error {
	confPath := flag.String("c", "config.yaml", "path to config file")
	flag.Parse()
	config.Init(*confPath)
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.wsHandler)
	cfg := config.Get()
	s.logger.Info("listen on " + cfg.ListenAddr)
	s.logger.Info(cfg.String())
	httpServer := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: mux,
	}
	return httpServer.ListenAndServeTLS(cfg.CertFile.Cert, cfg.CertFile.Key)
}
