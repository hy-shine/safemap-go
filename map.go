package safemap

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrMissingHashFunc = errors.New("hash function is required")

const (
	// default buckets count
	defaultBucketCount = 1 << 5
	// max buckets count
	maxBucketCount = 1 << 8
)

type SafeMap[K comparable, V any] interface {
	// Get returns key's value and exists
	Get(key K) (val V, exists bool)

	// Set sets key's value
	Set(key K, val V)

	// Delete deletes key
	Delete(key K)

	// GetAndDelete returns the existing value for the key and delete.
	// if the key exists, the loaded result is true.
	// Otherwise, it returns zero value and false.
	GetAndDelete(key K) (val V, loaded bool)

	// GetOrSet returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored.
	GetOrSet(key K, val V) (V, bool)

	// Clear clears the map
	Clear()

	// Len returns map items total
	Len() int

	// Range calls f sequentially for each key and value present in the map.
	// If f returns false, the iteration stops.
	Range(f func(k K, val V) bool)

	// IsEmpty returns true if map is empty
	IsEmpty() bool
}

type bucketMap[K comparable, V any] struct {
	sync.RWMutex
	innerMap map[K]V
}

type safeMap[K comparable, V any] struct {
	count   int32
	buckets []*bucketMap[K, V]
	*options[K]
}

// New returns a new SafeMap
func New[K comparable, V any](options ...OptFunc[K]) (SafeMap[K, V], error) {
	opt, err := loadOptfuns(options...)
	if err != nil {
		return nil, err
	}

	m := &safeMap[K, V]{
		buckets: make([]*bucketMap[K, V], opt.bucketTotal),
		options: opt,
		count:   0,
	}

	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i] = &bucketMap[K, V]{innerMap: make(map[K]V)}
	}

	return m, nil
}

// hashIndex returns key's lock index
func (m *safeMap[K, V]) hashIndex(key K) int {
	return int(m.hashFunc(key) & uint64(m.bucketTotal-1))
}

// allLock locks all buckets
func (m *safeMap[K, V]) allLock() {
	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i].Lock()
	}
}

// allUnlock unlocks all buckets
func (m *safeMap[K, V]) allUnlock() {
	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i].Unlock()
	}
}

// Get returns key's value
func (m *safeMap[K, V]) Get(key K) (V, bool) {
	index := m.hashIndex(key)
	m.buckets[index].RLock()
	val, b := m.buckets[index].innerMap[key]
	m.buckets[index].RUnlock()
	return val, b
}

// Set sets key's value
func (m *safeMap[K, V]) Set(key K, val V) {
	index := m.hashIndex(key)
	m.buckets[index].Lock()
	if _, b := m.buckets[index].innerMap[key]; !b {
		atomic.AddInt32(&m.count, 1)
	}
	m.buckets[index].innerMap[key] = val
	m.buckets[index].Unlock()
}

func (m *safeMap[K, V]) Delete(key K) {
	index := m.hashIndex(key)
	m.buckets[index].Lock()
	if _, b := m.buckets[index].innerMap[key]; b {
		delete(m.buckets[index].innerMap, key)
		atomic.AddInt32(&m.count, -1)
	}
	m.buckets[index].Unlock()
}

func (m *safeMap[K, V]) GetAndDelete(key K) (val V, loaded bool) {
	index := m.hashIndex(key)
	m.buckets[index].Lock()
	if val, b := m.buckets[index].innerMap[key]; b {
		delete(m.buckets[index].innerMap, key)
		atomic.AddInt32(&m.count, -1)
		m.buckets[index].Unlock()
		return val, true
	} else {
		m.buckets[index].Unlock()
		return val, false
	}
}

// Clear clears the map
func (m *safeMap[K, V]) Clear() {
	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i].Lock()
		// clear all keys
		// avoid make new map
		bucketLen := len(m.buckets[i].innerMap)
		for key := range m.buckets[i].innerMap {
			delete(m.buckets[i].innerMap, key)
		}
		atomic.AddInt32(&m.count, -int32(bucketLen))
		m.buckets[i].Unlock()
	}
}

// Len returns map items total
func (m *safeMap[K, V]) Len() int {
	return int(atomic.LoadInt32(&m.count))
}

// IsEmpty returns true if map is empty
func (m *safeMap[K, V]) IsEmpty() bool {
	return atomic.LoadInt32(&m.count) == 0
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *safeMap[K, V]) GetOrSet(key K, val V) (V, bool) {
	index := m.hashIndex(key)
	m.buckets[index].Lock()
	if val, b := m.buckets[index].innerMap[key]; b {
		m.buckets[index].Unlock()
		return val, true
	}

	m.buckets[index].innerMap[key] = val
	atomic.AddInt32(&m.count, 1)
	m.buckets[index].Unlock()
	return val, false
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, the iteration stops.
func (m *safeMap[K, V]) Range(f func(k K, v V) bool) {
	m.allLock()
	for i := 0; i < m.bucketTotal; i++ {
		for key, val := range m.buckets[i].innerMap {
			if !f(key, val) {
				m.allUnlock()
				return
			}
		}
	}
	m.allUnlock()
}
