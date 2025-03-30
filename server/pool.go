package server

import (
	"net"
	"sync"
)

type AuthConn struct {
	net.Conn
	Authed bool // 标记是否通过验证
}

type ConnectionPool struct {
	mu    sync.Mutex
	conns map[string]*AuthConn
}

func (p *ConnectionPool) Add(conn *AuthConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.conns[conn.RemoteAddr().String()] = conn
}

func (p *ConnectionPool) Get() *AuthConn {
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, conn := range p.conns {
		if conn.Authed {
			return conn
		} else {
			delete(p.conns, conn.RemoteAddr().String())
		}
	}
	return nil
}

func (p *ConnectionPool) Remove(conn *AuthConn) {
	p.mu.Lock()
	defer p.mu.Unlock()
	conn.Close()
	delete(p.conns, conn.RemoteAddr().String())
}
