package utils

import "sync"

type SafeMap[K comparable, V any] struct {
	rw sync.RWMutex
	m  map[K]V
}

func NewSafeMap[K comparable, V any]() *SafeMap[K, V] {
	return &SafeMap[K, V]{
		m: make(map[K]V),
	}
}

func (s *SafeMap[K, V]) Get(key K) V {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return s.m[key]
}

func (s *SafeMap[K, V]) Set(key K, val V) {
	s.rw.Lock()
	defer s.rw.Unlock()
	s.m[key] = val
}

func (s *SafeMap[K, V]) Delete(key K) V {
	s.rw.Lock()
	defer s.rw.Unlock()
	val, ok := s.m[key]
	if ok {
		delete(s.m, key)
	}
	return val
}

func (s *SafeMap[K, V]) Len() int {
	s.rw.RLock()
	defer s.rw.RUnlock()
	return len(s.m)
}

func (s *SafeMap[K, V]) Range(f func(K, V) bool) {
	s.rw.Lock()
	defer s.rw.Unlock()
	for k, v := range s.m {
		if !f(k, v) {
			break
		}
	}
}

func (s *SafeMap[K, V]) InnerSet(key K, val V) {
	s.m[key] = val
}

func (s *SafeMap[K, V]) InnerDelete(key K) V {
	val, ok := s.m[key]
	if ok {
		delete(s.m, key)
	}
	return val
}

func (s *SafeMap[K, V]) InnerLen() int {
	return len(s.m)
}
