package client

import (
	"ShadowO/client/config"
	"ShadowO/client/tunnel"
	"flag"
	"log/slog"
)

func Run() {
	confPath := flag.String("c", "config.yaml", "path to config file")
	flag.Parse()
	config.Init(*confPath)
	cfg := config.Get()
	logger := slog.Default()
	logger.Info("ShadowO client started")
	logger.Info("config: " + cfg.String())
	pools := make([]*tunnel.Pool, len(cfg.Mappings))
	for i, mapping := range cfg.Mappings {
		pools[i] = tunnel.NewPool(mapping.Name, cfg.RemoteAddr, mapping.RemotePort, mapping.LocalAddr, logger)
	}
	select {}
}
