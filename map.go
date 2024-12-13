package safemap

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrMissingHashFunc = errors.New("missing hash func")

const (
	// default lock count
	defaultLockCount = 32
	// max lock count
	maxLockCount = 256
)

type SafeMap[K comparable, V any] interface {
	// Get returns key's value and exists
	Get(key K) (val V, exists bool)

	// Set sets key's value
	Set(key K, val V)

	// Delete deletes key
	Delete(key K)

	// GetAndDelete returns key's value and delete it
	// returns false if key not exists
	GetAndDelete(key K) (val V, loaded bool)

	// GetOrStore returns the existing value for the key if present.
	// Otherwise, it stores and returns the given value.
	// The loaded result is true if the value was loaded, false if stored.
	GetOrStore(key K, val V) (V, bool)

	// Clear clears the map
	Clear()

	// Len returns map items total
	Len() int

	// IsEmpty returns true if map is empty
	IsEmpty() bool
}

type unitMap[K comparable, V any] struct {
	sync.RWMutex
	innerMap map[K]V
}

type safeMap[K comparable, V any] struct {
	count      int32
	listShared []*unitMap[K, V]
	*opt[K]
}

// New returns a new SafeMap
func New[K comparable, V any](options ...OptFunc[K]) (SafeMap[K, V], error) {
	opt, err := loadOptfuns(options...)
	if err != nil {
		return nil, err
	}

	m := &safeMap[K, V]{
		opt: opt,
	}

	for range m.lock {
		m.listShared = append(m.listShared, &unitMap[K, V]{innerMap: make(map[K]V)})
	}

	return m, nil
}

// hashIndex returns key's lock index
func (m *safeMap[K, V]) hashIndex(key K) int {
	return int(m.hashFunc(key) % uint64(m.lock))
}

// Get returns key's value
func (m *safeMap[K, V]) Get(key K) (V, bool) {
	index := m.hashIndex(key)
	m.listShared[index].RLock()
	val, b := m.listShared[index].innerMap[key]
	m.listShared[index].RUnlock()
	return val, b
}

// Set sets key's value
func (m *safeMap[K, V]) Set(key K, val V) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	if _, b := m.listShared[index].innerMap[key]; !b {
		atomic.AddInt32(&m.count, 1)
	}
	m.listShared[index].innerMap[key] = val
	m.listShared[index].Unlock()
}

func (m *safeMap[K, V]) Delete(key K) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	if _, b := m.listShared[index].innerMap[key]; b {
		atomic.AddInt32(&m.count, -1)
		delete(m.listShared[index].innerMap, key)
	}
	m.listShared[index].Unlock()
}

func (m *safeMap[K, V]) GetAndDelete(key K) (val V, loaded bool) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	if val, b := m.listShared[index].innerMap[key]; b {
		delete(m.listShared[index].innerMap, key)
		m.listShared[index].Unlock()
		return val, b
	} else {
		m.listShared[index].Unlock()
		return val, false
	}
}

// Clear clears the map
func (m *safeMap[K, V]) Clear() {
	for i := 0; i < m.lock; i++ {
		m.listShared[i].Lock()
	}
	for i := 0; i < m.lock; i++ {
		m.listShared[i].innerMap = make(map[K]V)
		m.listShared[i].Unlock()
	}
	atomic.StoreInt32(&m.count, 0)
}

// Len returns map items total
func (m *safeMap[K, V]) Len() int {
	return int(atomic.LoadInt32(&m.count))
}

// IsEmpty returns true if map is empty
func (m *safeMap[K, V]) IsEmpty() bool {
	return atomic.LoadInt32(&m.count) == 0
}

// GetOrStore returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *safeMap[K, V]) GetOrStore(key K, val V) (V, bool) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	defer m.listShared[index].Unlock()
	if val, b := m.listShared[index].innerMap[key]; b {
		return val, b
	} else {
		m.listShared[index].innerMap[key] = val
		return val, false
	}
}
