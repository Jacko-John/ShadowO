package tunnel

import (
	"ShadowO/utils"
	"fmt"
	"log/slog"
	"strconv"
	"sync/atomic"
	"time"
)

// 连接池配置参数
const (
	MinIdleConnections  = 1  // 最小空闲连接数
	MaxIdleConnections  = 5  // 最大空闲连接数
	MaxRetryCount       = 3  // 最大重试次数
	HealthCheckInterval = 30 // 健康检查间隔
)

type Pool struct {
	name       string
	remoteAddr string
	loaclAddr  string
	remotePort string
	idleConns  *utils.SafeMap[string, *Tunnel]
	closed     atomic.Bool
	logger     *slog.Logger
	idleCount  atomic.Int32
	retryCount atomic.Int32
}

// type Tunnel struct {
// 	conn     net.Conn
// 	mu       utils.TryMutex
// 	signal   chan bool
// 	Addr     string
// 	isClosed atomic.Bool
// 	pool     *Pool
// }

// var logger = slog.Default()

func NewPool(name, remoteAddr, remotePort, localAddr string, logger *slog.Logger) *Pool {
	p := &Pool{
		name:       name,
		remoteAddr: remoteAddr,
		loaclAddr:  localAddr,
		remotePort: remotePort,
		idleConns:  utils.NewSafeMap[string, *Tunnel](),
		logger:     logger,
	}

	// 初始化最小连接数
	for range MinIdleConnections {
		t, err := p.createAndStartNewTunnel()
		if err != nil {
			p.logger.Error(fmt.Sprintf("%s :create tunnel failed, %v", p.name, err))
			continue
		}
		p.idleConns.Set(t.conn.LocalAddr().String(), t)
		p.idleCount.Add(1)
	}

	go p.healthCheck()
	return p
}

func (p *Pool) createTunnel() (*Tunnel, error) {
	rport, err := strconv.Atoi(p.remotePort)
	if err != nil {
		return nil, err
	}
	t := &Tunnel{
		signal:     make(chan bool, 1),
		localAddr:  p.loaclAddr,
		remoteAddr: p.remoteAddr,
		remotePort: int32(rport),
		pool:       p,
		logger:     p.logger,
	}
	return t, nil
}

func (p *Pool) createAndStartNewTunnel() (*Tunnel, error) {
	t, err := p.createTunnel()
	if err != nil {
		return nil, err
	}
	if err = t.ConnectRemote(); err != nil {
		return nil, err
	}
	go t.Do()
	return t, nil
}

func (p *Pool) healthCheck() {
	ticker := time.NewTicker(time.Second * HealthCheckInterval)
	defer ticker.Stop()

	for range ticker.C {
		if p.closed.Load() {
			return
		}
		p.idleConns.Range(func(key string, value *Tunnel) bool {
			if value == nil || value.isClosed.Load() {
				p.idleCount.Add(-1)
				p.idleConns.InnerDelete(key)
			}
			return true
		})
		p.maintainPool()
	}
}

func (p *Pool) maintainPool() {
	// 维持最小连接数
	for p.idleCount.Load() < MinIdleConnections && !p.closed.Load() {
		t, err := p.createAndStartNewTunnel()
		if err != nil {
			p.retryCount.Add(1)
			p.logger.Error(fmt.Sprintf("%s :create tunnel failed, %v", p.name, err))
			if p.retryCount.Load() >= MaxRetryCount {
				p.logger.Error(fmt.Sprintf("%s :create tunnel failed after %d retries, %v", p.name, p.retryCount.Load(), err))
				p.logger.Error("closing " + p.name)
				p.Close()
				return
			}
			continue
		}
		p.idleCount.Add(1)
		p.idleConns.Set(t.conn.LocalAddr().String(), t)
	}
}

func (p *Pool) PutTunnel(t *Tunnel) {
	if p.closed.Load() || p.idleCount.Load() >= MaxIdleConnections {
		t.Close()
		return
	}
	p.idleCount.Add(1)
	p.idleConns.Set(t.conn.LocalAddr().String(), t)
}

func (p *Pool) GetTunnelByLocalAddr(localAddr string) *Tunnel {
	if p.closed.Load() {
		return nil
	}
	t := p.idleConns.Delete(localAddr)
	if t != nil {
		p.idleCount.Add(-1)
		p.maintainPool()
	}
	return t
}

func (p *Pool) Close() {
	if p.closed.CompareAndSwap(false, true) {
		p.idleConns.Range(func(key string, value *Tunnel) bool {
			value.Close()
			return true
		})
		p = nil
	}
}

// 其他方法保持原有优化...
