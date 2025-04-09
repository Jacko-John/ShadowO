package pool

import (
	"ShadowO/protocal"
	"ShadowO/utils"
	"fmt"
	"log/slog"
	"net"
	"sync/atomic"
)

type Tunnel struct {
	// TODO: implement tunnel
	conn       net.Conn
	mu         *utils.TicketLock
	newTCPChan chan *net.Conn
	headerChan chan byte
	isClosed   atomic.Bool // 原子标记是否已关闭
	pool       *Pool
	logger     *slog.Logger
}

func NewTunnel(conn net.Conn, pool *Pool, logger *slog.Logger) *Tunnel {
	t := &Tunnel{
		conn:       conn,
		mu:         utils.NewTicketLock(),
		newTCPChan: make(chan *net.Conn, 1),
		headerChan: make(chan byte, 1),
		isClosed:   atomic.Bool{},
		pool:       pool,
		logger:     logger,
	}
	go t.Do()
	return t
}

func (t *Tunnel) Do() {
	var err error
	var tcp *net.Conn
	var header byte
	defer func() {
		t.Close()
	}()
	go t.readLoop()
	for {
		select {
		case tcp = <-t.newTCPChan:
			_, err = t.conn.Write([]byte{byte(protocal.Header_NEW)})
			if err != nil {
				t.logger.Error(fmt.Sprintf("Failed to send NEW header: %v", err))
				return
			}
		case header = <-t.headerChan:
			switch protocal.Header(header) {
			case protocal.Header_PING:
				t.conn.Write([]byte{byte(protocal.Header_PONG)})
			case protocal.Header_GOT:
				if !t.handleNewRequest(tcp) {
					return
				}
			default:
				t.logger.Error(fmt.Sprintf("Unexpected header type: %v", header))
				return
			}
		}

	}
}

func (t *Tunnel) readLoop() {
	header := make([]byte, 1)
	ticket := t.mu.GetTicket()
	for {
		t.mu.Lock(ticket)
		_, err := t.conn.Read(header)
		if err != nil {
			if err == protocal.ErrEOF {
				continue
			}
			t.logger.Error(fmt.Sprintf("Failed to read header: %v", err))
			t.Close()
			t.mu.Unlock()
			return
		}
		if header[0] == byte(protocal.Header_GOT) {
			ticket = t.mu.GetSkipTicket(1)
		} else {
			ticket = t.mu.GetTicket()
		}
		t.headerChan <- header[0]
		t.mu.Unlock()
	}
}

func (t *Tunnel) handleNewRequest(tcp *net.Conn) bool {
	ticket := t.mu.GetTicket()
	t.mu.Lock(ticket)
	defer func() {
		t.mu.Unlock()
		t.pool.Put(t)
	}()
	err := protocal.NetConn(tcp, &t.conn)
	if err != nil {
		t.logger.Error(fmt.Sprintf("Something wrong with the connection: %v", err))
		return false
	}
	return true
}

func (t *Tunnel) Close() {
	if t.isClosed.CompareAndSwap(false, true) {
		if t.pool != nil {
			t.pool.GetByKey(t.conn.RemoteAddr().String())
		}
		if t.conn != nil {
			t.conn.Close()
		}
		close(t.newTCPChan)
		close(t.headerChan)
		t = nil
	}
}
