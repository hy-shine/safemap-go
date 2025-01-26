package safemap

import "sync"

type RwLock[T comparable, V any] struct {
	m  map[T]V
	mu sync.RWMutex
}

// Get returns the value for the key if present.
// The second return value bool is true if the value was found, or false if not.
func (l *RwLock[T, V]) Get(key T) (V, bool) {
	l.mu.RLock()
	val, b := l.m[key]
	l.mu.RUnlock()
	return val, b
}

// Set stores the given value for the specified key in the map.
// If the key already exists, its value will be overwritten.
// The operation is protected by a write lock to ensure thread safety.
func (l *RwLock[T, V]) Set(key T, val V) {
	l.mu.Lock()
	l.m[key] = val
	l.mu.Unlock()
}

// Delete removes the key-value pair from the map.
// The operation is protected by a write lock to ensure thread safety.
func (l *RwLock[T, V]) Delete(key T) {
	l.mu.Lock()
	delete(l.m, key)
	l.mu.Unlock()
}

// GetAndDelete retrieves and removes the value associated with the specified key.
// It returns the value and a boolean indicating whether the key was found and deleted.
// The operation is protected by a write lock to ensure thread safety.
func (l *RwLock[T, V]) GetAndDelete(key T) (val V, loaded bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if val, b := l.m[key]; b {
		delete(l.m, key)
		return val, true
	} else {
		return val, false
	}
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (l *RwLock[T, V]) GetOrSet(key T, val V) (V, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if val, b := l.m[key]; b {
		return val, true
	}
	l.m[key] = val
	return val, false
}

// Len returns the number of key-value pairs in the map.
// The operation is protected by a read lock to ensure thread safety.
func (l *RwLock[T, V]) Len() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.m)
}

// Range iterates over the map and calls the provided function for each key-value pair.
// The operation is protected by a read lock to ensure thread safety.
func (l *RwLock[T, V]) Range(f func(key T, val V) bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	for key, val := range l.m {
		if !f(key, val) {
			break
		}
	}
}

// NewRwLock returns a new initialized RwLock.
func NewRwLock[T comparable, V any]() *RwLock[T, V] {
	return &RwLock[T, V]{
		m: make(map[T]V),
	}
}
