package pool

import (
	"ShadowO/utils"
	"errors"
	"log/slog"
	"net"
	"time"
)

// 连接池配置参数
const (
	HealthCheckInterval = 30
)

type Pool struct {
	port       string
	idleConns  *utils.SafeMap[string, *Tunnel]
	usingConns *utils.SafeMap[string, *Tunnel]
	logger     *slog.Logger
}

// type Tunnel struct {
// 	conn     net.Conn
// 	mu       utils.TryMutex
// 	signal   chan bool
// 	Addr     string
// 	isClosed atomic.Bool
// 	pool     *Pool
// }

// var tunnelCount int32 = 0
// var logger = slog.Default()

func NewPool(port string, logger *slog.Logger) *Pool {
	p := &Pool{
		port:       port,
		idleConns:  utils.NewSafeMap[string, *Tunnel](),
		usingConns: utils.NewSafeMap[string, *Tunnel](),
		logger:     logger,
	}
	return p
}

func (p *Pool) Start() error {
	tcpListener, err := net.Listen("tcp", ":"+p.port)
	if err != nil {
		return err
	}
	defer tcpListener.Close()
	go p.healthCheck()
	for {
		conn, err := tcpListener.Accept()
		if err != nil {
			return err
		}
		t, err := p.Get()
		if err != nil {
			conn.Close()
			continue
		}
		t.newTCPChan <- &conn
	}
}

func (p *Pool) healthCheck() {
	ticker := time.NewTicker(HealthCheckInterval * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		p.idleConns.Range(func(key string, value *Tunnel) bool {
			if value == nil || value.isClosed.Load() {
				p.idleConns.InnerDelete(key)

			}
			return true
		})
		p.usingConns.Range(func(key string, value *Tunnel) bool {
			if value == nil || value.isClosed.Load() {
				p.usingConns.InnerDelete(key)
			}
			return true
		})
	}
}

func (p *Pool) Get() (*Tunnel, error) {
	var t *Tunnel
	var k string
	p.idleConns.Range(func(key string, value *Tunnel) bool {
		if value.isClosed.Load() {
			p.idleConns.InnerDelete(key)
			return true
		}
		k, t = key, value
		return false
	})
	if t == nil {
		return nil, errors.New("no idle tunnels")
	}
	p.usingConns.Set(k, t)
	return t, nil
}

func (p *Pool) Put(t *Tunnel) {
	p.usingConns.Delete(t.conn.RemoteAddr().String())
	if t.isClosed.Load() {
		return
	}
	p.idleConns.Set(t.conn.RemoteAddr().String(), t)
}

func (p *Pool) GetByKey(key string) (*Tunnel, error) {
	t := p.idleConns.Get(key)
	if t == nil {
		return nil, errors.New("no such idle tunnels")
	}
	return t, nil
}

// func (p *Pool) Get() (*Tunnel, error) {
// 	select {
// 	case t := <-p.idleConns:
// 		if t.isClosed.Load() {
// 			return p.createAndStartNewTunnel()
// 		}
// 		return t, nil
// 	default:
// 		return p.createAndStartNewTunnel()
// 	}
// }

// func (p *Pool) Put(t *Tunnel) {
// 	if t.isClosed.Load() || p.closed.Load() {
// 		t.Close()
// 		return
// 	}

// 	select {
// 	case p.idleConns <- t:
// 		p.maintainPool()
// 	default:
// 		t.Close()
// 	}
// }

// func (p *Pool) healthCheck() {
// 	ticker := time.NewTicker(HealthCheckInterval)
// 	defer ticker.Stop()

// 	for range ticker.C {
// 		if p.closed.Load() {
// 			return
// 		}

// 		p.mu.Lock()
// 		size := len(p.idleConns)
// 		for i := 0; i < size; i++ {
// 			select {
// 			case t := <-p.idleConns:
// 				if t.isClosed.Load() {
// 					t.Close()
// 				} else {
// 					p.idleConns <- t
// 				}
// 			default:
// 				break
// 			}
// 		}
// 		p.maintainPool()
// 		p.mu.Unlock()
// 	}
// }

// func (p *Pool) Close() {
// 	p.closed.Store(true)
// 	close(p.idleConns)
// 	for t := range p.idleConns {
// 		t.Close()
// 	}
// }

// 其他方法保持原有优化...
