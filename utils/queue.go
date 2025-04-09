package utils

import "sync"

// 非并发安全版本
type Queue[T any] struct {
	data  []T
	head  int // 环形缓冲区头指针
	tail  int // 环形缓冲区尾指针
	count int // 当前元素数量
}

func NewQueue[T any]() *Queue[T] {
	return &Queue[T]{
		data: make([]T, 1), // 初始容量为1
	}
}

func (q *Queue[T]) Enqueue(val T) {
	// 扩容条件：队列已满
	if q.count == len(q.data) {
		newCap := q.count * 2
		if newCap == 0 {
			newCap = 1
		}
		newData := make([]T, newCap)
		// 复制元素到新数组
		if q.head < q.tail {
			copy(newData, q.data[q.head:q.tail])
		} else {
			n := copy(newData, q.data[q.head:])
			copy(newData[n:], q.data[:q.tail])
		}
		q.data = newData
		q.head = 0
		q.tail = q.count
	}
	q.data[q.tail] = val
	q.tail = (q.tail + 1) % len(q.data)
	q.count++
}

func (q *Queue[T]) Dequeue() T {
	if q.count == 0 {
		panic("Queue is empty")
	}
	val := q.data[q.head]
	var zero T
	q.data[q.head] = zero // 避免内存泄漏（清理旧值）
	q.head = (q.head + 1) % len(q.data)
	q.count--

	// 缩容条件：元素数量小于容量的1/4，且当前容量 > 2
	if len(q.data) > 2 && q.count*4 <= len(q.data) {
		newCap := len(q.data) / 2
		newData := make([]T, newCap)
		if q.head < q.tail {
			copy(newData, q.data[q.head:q.tail])
		} else {
			n := copy(newData, q.data[q.head:])
			copy(newData[n:], q.data[:q.tail])
		}
		q.data = newData
		q.head = 0
		q.tail = q.count
	}
	return val
}

// DequeueSafe 安全出队，返回值和是否成功
func (q *Queue[T]) DequeueSafe() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	return q.Dequeue(), true
}

func (q *Queue[T]) FrontSafe() (T, bool) {
	if q.count == 0 {
		var zero T
		return zero, false
	}
	return q.data[q.head], true
}

func (q *Queue[T]) IsEmpty() bool {
	return q.count == 0
}

func (q *Queue[T]) Size() int {
	return q.count
}

// 并发安全版本（包装非并发安全队列）
type ConcurrentQueue[T any] struct {
	q  *Queue[T]
	mu sync.Mutex
}

func NewConcurrentQueue[T any]() *ConcurrentQueue[T] {
	return &ConcurrentQueue[T]{q: NewQueue[T]()}
}

func (cq *ConcurrentQueue[T]) Enqueue(val T) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	cq.q.Enqueue(val)
}

func (cq *ConcurrentQueue[T]) Dequeue() T {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return cq.q.Dequeue()
}

func (cq *ConcurrentQueue[T]) DequeueSafe() (T, bool) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return cq.q.DequeueSafe()
}

func (cq *ConcurrentQueue[T]) IsEmpty() bool {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return cq.q.IsEmpty()
}

func (cq *ConcurrentQueue[T]) Size() int {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return cq.q.Size()
}
