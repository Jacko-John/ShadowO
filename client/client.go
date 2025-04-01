package client

import (
	"ShadowO/client/config"
	"ShadowO/client/tunnel"
	"flag"
	"log/slog"
)

type Client struct {
	// authSecret string
	// serverUrl  string
	// localUrl   string
}

func (c *Client) Run() {
	confPath := flag.String("c", "config.yaml", "path to config file")
	flag.Parse()
	config.Init(*confPath)
	cfg := config.Get()
	pools := make([]*tunnel.Pool, len(cfg.Mappings))
	for i, mapping := range cfg.Mappings {
		rmaddr := cfg.RemoteAddr + ":" + mapping.RemotePort
		pools[i] = tunnel.NewPool(mapping.Name, rmaddr, mapping.LocalAddr, slog.Default())
	}
	select {}
}
