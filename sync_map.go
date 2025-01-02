package safemap

import "sync"

// SyncMap is a generic wrapper around sync.Map that provides
// type-safe concurrent map operations.
// It allows storing and retrieving key-value pairs with generic types K and V,
// and implements
// the thread-safe properties of the standard library's sync.Map.
type SyncMap[K comparable, V any] struct {
	p sync.Map
}

// Get returns key's value, and exists.
//
// Same as sync.Map.Load
func (m *SyncMap[K, V]) Get(key K) (value V, exists bool) {
	_val, exists := m.p.Load(key)
	if exists {
		return _val.(V), true
	}
	return value, false
}

// Set sets key's value, same as sync.Map.Store
func (m *SyncMap[K, V]) Set(key K, value V) {
	m.p.Store(key, value)
}

// Delete deletes key, same as sync.Map.Delete
func (m *SyncMap[K, V]) Delete(key K) {
	m.p.Delete(key)
}

// GetAndDelete returns the existing value for the key and delete.
// Same as sync.Map.LoadAndDelete
func (m *SyncMap[K, V]) GetAndDelete(key K) (value V, loaded bool) {
	_val, loaded := m.p.LoadAndDelete(key)
	if loaded {
		return _val.(V), true
	}
	return value, false
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, the iteration stops.
// Same as sync.Map.Range
func (m *SyncMap[K, V]) Range(f func(K, V) bool) {
	m.Range(f)
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
//
// Same as sync.Map.LoadOrStore
func (m *SyncMap[K, V]) GetOrSet(key K, val V) (actual V, loaded bool) {
	_val, loaded := m.p.LoadOrStore(key, val)
	if loaded {
		return _val.(V), true
	}
	return actual, false
}

// Swap stores the value for the key and returns the previous value.
// Same as sync.Map.Swap
func (m *SyncMap[K, V]) Swap(key K, val V) (previous V, loaded bool) {
	_val, loaded := m.p.Swap(key, val)
	if loaded {
		return _val.(V), true
	}
	return previous, false
}

// CompareAndDelete deletes the entry for key if its value is equal to old.
// The old value must be of a comparable type.
//
// Same as sync.Map.CompareAndDelete
func (m *SyncMap[K, V]) CompareAndDelete(key K, old V) (deleted bool) {
	return m.p.CompareAndDelete(key, old)
}

// CompareAndSwap swaps the old and new values for key
// if the value stored in the map is equal to old.
// The old value must be of a comparable type.
//
// Same as sync.Map.CompareAndSwap
func (m *SyncMap[K, V]) CompareAndSwap(key K, old, new V) bool {
	return m.p.CompareAndSwap(key, old, new)
}
