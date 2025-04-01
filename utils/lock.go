package utils

import (
	"sync"
	"sync/atomic"
)

type TryMutex struct {
	mu    sync.Mutex
	state int32 // 0: 未锁定, 1: 已锁定
}

func (m *TryMutex) TryLock() bool {
	if atomic.CompareAndSwapInt32(&m.state, 0, 1) {
		m.mu.Lock()
		return true
	}
	return false
}

func (m *TryMutex) Lock() {
	m.mu.Lock()
	atomic.StoreInt32(&m.state, 1)
}

func (m *TryMutex) Unlock() {
	atomic.StoreInt32(&m.state, 0)
	m.mu.Unlock()
}
