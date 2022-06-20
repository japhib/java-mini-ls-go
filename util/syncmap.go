package util

import "sync"

type SyncMap[K comparable, V any] struct {
	data map[K]V
	mu   sync.RWMutex
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	return &SyncMap[K, V]{
		data: make(map[K]V),
		mu:   sync.RWMutex{},
	}
}

// Get gets a value from the map
func (sm *SyncMap[K, V]) Get(key K) (V, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	result, ok := sm.data[key]
	return result, ok
}

// Set sets a value in the map
func (sm *SyncMap[K, V]) Set(key K, val V) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = val
}
