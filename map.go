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
	Get(key K) (val V, exists bool)
	Set(key K, val V)
	Delete(key K)
	GetAndDelete(key K) (val V, loaded bool)
	// Clear()
	Cap() int
	IsEmpty() bool
}

type unitMap[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

type safeMap[K comparable, V any] struct {
	count      int32
	listShared []*unitMap[K, V]
	*opt[K]
}

func New[K comparable, V any](options ...OptFunc[K]) (SafeMap[K, V], error) {
	opt, err := loadOptfuns(options...)
	if err != nil {
		return nil, err
	}

	m := &safeMap[K, V]{
		opt: opt,
	}

	for range m.lock {
		m.listShared = append(m.listShared, &unitMap[K, V]{m: make(map[K]V)})
	}

	return m, nil
}

// hashIndex returns key's index
func (m *safeMap[K, V]) hashIndex(key K) int {
	return int(m.hashFunc(key) % uint64(m.lock))
}

func (m *safeMap[K, V]) Get(key K) (V, bool) {
	index := m.hashIndex(key)
	m.listShared[index].RLock()
	val, b := m.listShared[index].m[key]
	m.listShared[index].RUnlock()
	return val, b
}

func (m *safeMap[K, V]) Set(key K, val V) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	if _, b := m.listShared[index].m[key]; !b {
		atomic.AddInt32(&m.count, 1)
	}
	m.listShared[index].m[key] = val
	m.listShared[index].Unlock()
}

func (m *safeMap[K, V]) Delete(key K) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	if _, b := m.listShared[index].m[key]; b {
		atomic.AddInt32(&m.count, -1)
		delete(m.listShared[index].m, key)
	}
	m.listShared[index].Unlock()
}

func (m *safeMap[K, V]) GetAndDelete(key K) (val V, loaded bool) {
	index := m.hashIndex(key)
	m.listShared[index].Lock()
	if val, b := m.listShared[index].m[key]; b {
		delete(m.listShared[index].m, key)
		m.listShared[index].Unlock()
		return val, b
	} else {
		m.listShared[index].Unlock()
		return val, false
	}
}

// Cap returns map items total
func (m *safeMap[K, V]) Cap() int {
	return int(atomic.LoadInt32(&m.count))
}

// IsEmpty returns true if map is empty
func (m *safeMap[K, V]) IsEmpty() bool {
	return atomic.LoadInt32(&m.count) == 0
}
