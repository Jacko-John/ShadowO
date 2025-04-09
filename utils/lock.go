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

// TicketLock 基于排号机制的公平锁
type TicketLock struct {
	mu         sync.Mutex    // 保护共享资源的互斥锁
	nextTicket int32         // 下一个可分配的票号
	skipQueue  *Queue[int32] // 跳过的票号
	current    atomic.Int32  // 当前允许执行的票号
	cond       *sync.Cond    // 条件变量用于等待通知
}

// NewTicketLock 创建一个新的排号锁
func NewTicketLock() *TicketLock {
	tl := &TicketLock{}
	tl.cond = sync.NewCond(&sync.Mutex{})
	tl.skipQueue = NewQueue[int32]()
	return tl
}

func (tl *TicketLock) GetTicket() int32 {
	tl.mu.Lock()

	nt, ok := tl.skipQueue.FrontSafe()
	for ok && nt == tl.nextTicket {
		tl.skipQueue.Dequeue()
		tl.nextTicket++
		nt, ok = tl.skipQueue.FrontSafe()
	}

	ticket := tl.nextTicket
	tl.nextTicket++
	tl.mu.Unlock()
	return ticket
}

func (tl *TicketLock) GetSkipTicket(skip int32) int32 {
	tl.mu.Lock()
	ticket := tl.nextTicket + skip
	tl.skipQueue.Enqueue(ticket)
	tl.mu.Unlock()
	return ticket
}

// Lock 获取锁，返回当前票号并阻塞直到轮到自己
func (tl *TicketLock) Lock(ticket int32) {
	// 等待直到当前票号等于自己的票号
	// tl.mu1.Lock()
	// defer tl.mu1.Unlock()
	for tl.current.Load() != ticket {
		tl.cond.L.Lock()
		tl.cond.Wait()
	}

}

// Unlock 释放锁，允许下一个等待者执行
func (tl *TicketLock) Unlock() {
	// tl.mu1.Lock()
	// defer tl.mu1.Unlock()
	tl.current.Add(1)
	// 通知所有等待者检查当前票号
	tl.cond.Broadcast()
}
