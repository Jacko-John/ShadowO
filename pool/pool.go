package pool

// package tunnel

// import (
// 	"fmt"
// 	"log/slog"
// 	"net"
// 	"sync"
// 	"sync/atomic"
// 	"time"
// )

// // 连接池配置参数
// const (
// 	MinIdleConnections  = 1  // 最小空闲连接数
// 	MaxIdleConnections  = 10 // 最大空闲连接数
// 	HealthCheckInterval = 30 * time.Second
// )

// type Pool struct {
// 	addr      string
// 	idleConns chan *Tunnel
// 	mu        sync.Mutex
// 	closed    atomic.Bool
// 	logger    *slog.Logger
// }

// // type Tunnel struct {
// // 	conn     net.Conn
// // 	mu       utils.TryMutex
// // 	signal   chan bool
// // 	Addr     string
// // 	isClosed atomic.Bool
// // 	pool     *Pool
// // }

// // var tunnelCount int32 = 0
// // var logger = slog.Default()

// func NewPool(addr string) *Pool {
// 	p := &Pool{
// 		addr:      addr,
// 		idleConns: make(chan *Tunnel, MaxIdleConnections),
// 		logger:    logger,
// 	}

// 	// 初始化最小连接数
// 	for i := 0; i < MinIdleConnections; i++ {
// 		t := p.createTunnel()
// 		go t.Do()
// 	}

// 	// 启动健康检查
// 	go p.healthCheck()
// 	return p
// }

// func (p *Pool) createTunnel() *Tunnel {
// 	t := &Tunnel{
// 		signal: make(chan bool, 1),
// 		Addr:   p.addr,
// 		pool:   p,
// 	}
// 	atomic.AddInt32(&tunnelCount, 1)
// 	return t
// }

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

// func (p *Pool) createAndStartNewTunnel() (*Tunnel, error) {
// 	t := p.createTunnel()
// 	if err := t.Connect(); err != nil {
// 		return nil, err
// 	}
// 	go t.Do()
// 	return t, nil
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

// func (p *Pool) maintainPool() {
// 	p.mu.Lock()
// 	defer p.mu.Unlock()

// 	// 维持最小连接数
// 	currentIdle := len(p.idleConns)
// 	if currentIdle < MinIdleConnections {
// 		need := MinIdleConnections - currentIdle
// 		for i := 0; i < need; i++ {
// 			select {
// 			case p.idleConns <- p.createTunnel():
// 				go p.idleConns[len(p.idleConns)-1].Do()
// 			default:
// 				break
// 			}
// 		}
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

// func (t *Tunnel) Connect() error {
// 	conn, err := net.Dial("tcp", t.Addr)
// 	if err != nil {
// 		return fmt.Errorf("connection failed: %w", err)
// 	}
// 	t.conn = conn
// 	return nil
// }

// // 其他方法保持原有优化...
