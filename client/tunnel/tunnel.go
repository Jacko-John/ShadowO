package tunnel

import (
	"ShadowO/client/config"
	"ShadowO/protocal"
	"ShadowO/utils"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"
	"time"

	"github.com/lxzan/gws"
)

type Tunnel struct {
	// TODO: implement tunnel
	conn       net.Conn
	mu         utils.TryMutex
	signal     chan bool
	localAddr  string
	remoteAddr string
	remotePort int32
	isClosed   atomic.Bool // 原子标记是否已关闭
	pool       *Pool
	logger     *slog.Logger
}

func (t *Tunnel) Do() {
	header := make([]byte, 1)
	var err error
	defer func() {
		t.Close()
	}()
	go t.pinger()
	for {
		err = t.conn.SetReadDeadline(time.Now().Add(150 * time.Second))
		if err != nil {
			t.logger.Error(fmt.Sprintf("SetReadDeadline error: %v", err))
			return
		}
		_, err = t.conn.Read(header)
		if err != nil {
			t.logger.Error(fmt.Sprintf("Read header failed: %v", err))
			return
		}
		switch protocal.Header(header[0]) {
		case protocal.Header_PONG:
			t.signal <- true
		case protocal.Header_NEW:
			if !t.handleNewRequest() {
				return
			}
		default:
			t.logger.Error(fmt.Sprintf("Unexpected header type: %v", header[0]))
			return
		}
	}
}

func (t *Tunnel) handleNewRequest() bool {
	if !t.mu.TryLock() {
		buf := make([]byte, 1)
		if _, err := t.conn.Read(buf); err != nil {
			t.logger.Error(fmt.Sprintf("Read follow-up header, failed: %v", err))
			return false
		}
		if protocal.Header(buf[0]) != protocal.Header_PONG {
			t.logger.Error(fmt.Sprintf("Expected PONG after failed TryLock, got: %x", buf[0]))
			return false
		}
		t.signal <- true
		t.mu.Lock()
	}
	defer t.mu.Unlock()
	if _, err := t.conn.Write([]byte{byte(protocal.Header_GOT)}); err != nil {
		t.logger.Error(fmt.Sprintf("Failed to send GOT header: %v", err))
		return false
	}
	if err := t.ConnectLocal(); err != nil {
		t.logger.Error(fmt.Sprintf("ConnectLocal failed: %v", err))
		return false
	}
	return true
}

func (t *Tunnel) pinger() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		t.Close()
		ticker.Stop()
	}()
	for range ticker.C {
		if t.isClosed.Load() {
			return
		}

		t.mu.Lock()
		if _, err := t.conn.Write([]byte{byte(protocal.Header_PING)}); err != nil {
			t.mu.Unlock()
			t.logger.Error(fmt.Sprintf("Ping send failed: %v", err))
			return
		}

		select {
		case success := <-t.signal:
			t.mu.Unlock()
			if !success {
				return
			}
		case <-time.After(30 * time.Second): // 添加超时机制
			t.mu.Unlock()
			t.logger.Error("Ping timeout")
			return
		}
	}
}

func (t *Tunnel) ConnectLocal() error {
	tcp, err := net.Dial("tcp", t.localAddr)
	if err != nil {
		return fmt.Errorf("failed to dial local service: %w", err)
	}
	err = protocal.NetConn(&tcp, &t.conn)
	if err != nil {
		return err
	}
	return nil
}

func (t *Tunnel) ConnectRemote() error {
	socket, _, err := gws.NewClient(&gws.BuiltinEventHandler{}, &gws.ClientOption{
		TlsConfig:       &tls.Config{InsecureSkipVerify: config.Get().SkipVerify},
		Addr:            t.remoteAddr,
		ParallelEnabled: true,
	})
	if err != nil {
		return fmt.Errorf("failed to connect remote service: %w", err)
	}
	conn := socket.NetConn()
	err = protocal.AuthC(&conn, t.pool.remoteAddr, t.remotePort, protocal.ID_TUNNEL)
	if err != nil {
		return err
	}
	t.conn = conn
	return nil
}
func (t *Tunnel) Close() {
	if t.isClosed.CompareAndSwap(false, true) {
		if t.pool != nil {
			t.pool.GetTunnelByLocalAddr(t.conn.LocalAddr().String())
		}
		if t.conn != nil {
			t.conn.Close()
		}
		close(t.signal)
		t = nil
	}
}
