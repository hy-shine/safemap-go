package safemap

import (
	"errors"
	"sync"
	"sync/atomic"

	"golang.org/x/exp/constraints"
)

var ErrMissingHashFunc = errors.New("hash function is required")

const (
	// default buckets count
	defaultBucketCount = 1 << 5
	// max buckets count
	maxBucketCount = 1 << 10
)

type bucketMap[K comparable, V any] struct {
	sync.RWMutex
	innerMap map[K]V
}

// SafeMap is a thread-safe, generic map with configurable options.
// It uses a sharded locking mechanism to improve concurrent performance
// by reducing lock contention. The map is divided into multiple buckets,
// each with its own lock, allowing concurrent operations on different
// buckets to proceed independently.
//
// The map is designed for high-concurrency scenarios where
// thread safety and performance are important considerations.
//
// As you use this map, you must be create it with NewMap/NewStringMap/NewIntegerMap function.
type SafeMap[K comparable, V any] struct {
	count   int32
	buckets []*bucketMap[K, V]
	*options[K]
}

// NewMap creates a new thread-safe, generic map with configurable options.
//
// The function takes a variadic number of option functions that can customize
// the map's behavior. It supports different key and value types through Go's
// generics, with the constraint that the key type must be comparable.
// If the hashFunc is not provided, the function will return an ErrMissingHashFunc error.
// Parameters:
//   - options: Optional configuration functions to customize map behavior
//     (e.g., setting bucket count, custom hash functions)
//
// Returns:
//   - A SafeMap instance with the specified configuration
//
// Example:
//
//	// Create a default string-to-int safe map
//	m, err := NewMap[string, int]()
//
//	// Create a map with custom bucket count
//	m, err := NewMap[string, int](WithBuckets(8))
//
// The function initializes a map with multiple buckets to improve
// concurrent access performance by reducing lock contention.
func NewMap[K comparable, V any](options ...OptFunc[K]) (*SafeMap[K, V], error) {
	opt, err := loadOpts(options...)
	if err != nil {
		return nil, err
	}

	m := &SafeMap[K, V]{
		buckets: make([]*bucketMap[K, V], opt.bucketTotal),
		options: opt,
		count:   0,
	}

	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i] = &bucketMap[K, V]{innerMap: make(map[K]V)}
	}

	return m, nil
}

// NewStringMap returns a new string generic key SafeMap
func NewStringMap[K ~string, V any](options ...OptFunc[K]) *SafeMap[K, V] {
	options = append(options, WithHashFunc(func(k K) uint64 { return Hashstr(string(k)) }))
	m, _ := NewMap[K, V](options...)
	return m
}

// NewIntegerMap returns a new integer generic key SafeMap
func NewIntegerMap[K constraints.Integer, V any](options ...OptFunc[K]) *SafeMap[K, V] {
	options = append(options, WithHashFunc(func(k K) uint64 {
		if k < 0 {
			k = -k
		}
		return uint64(k)
	}))
	m, _ := NewMap[K, V](options...)
	return m
}

// hashIndex returns key's lock index
func (m *SafeMap[K, V]) hashIndex(key K) int {
	return int(m.hashFunc(key) & uint64(m.bucketTotal-1))
}

// allLock locks all buckets
func (m *SafeMap[K, V]) allLock() {
	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i].Lock()
	}
}

// allUnlock unlocks all buckets
func (m *SafeMap[K, V]) allUnlock() {
	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i].Unlock()
	}
}

// Get returns key's value
func (m *SafeMap[K, V]) Get(key K) (V, bool) {
	index := m.hashIndex(key)
	m.buckets[index].RLock()
	val, b := m.buckets[index].innerMap[key]
	m.buckets[index].RUnlock()
	return val, b
}

// Set sets key's value
func (m *SafeMap[K, V]) Set(key K, val V) {
	index := m.hashIndex(key)
	m.buckets[index].Lock()
	if _, b := m.buckets[index].innerMap[key]; !b {
		atomic.AddInt32(&m.count, 1)
	}
	m.buckets[index].innerMap[key] = val
	m.buckets[index].Unlock()
}

func (m *SafeMap[K, V]) Delete(key K) {
	index := m.hashIndex(key)
	m.buckets[index].Lock()
	if _, b := m.buckets[index].innerMap[key]; b {
		delete(m.buckets[index].innerMap, key)
		atomic.AddInt32(&m.count, -1)
	}
	m.buckets[index].Unlock()
}

func (m *SafeMap[K, V]) GetAndDelete(key K) (val V, loaded bool) {
	index := m.hashIndex(key)
	bucket := m.buckets[index]
	bucket.Lock()
	defer bucket.Unlock()

	if v, ok := bucket.innerMap[key]; ok {
		delete(bucket.innerMap, key)
		atomic.AddInt32(&m.count, -1)
		return v, true
	}
	return val, false // val will be the zero value for V
}

// Clear clears the map
func (m *SafeMap[K, V]) Clear() {
	// Lock all buckets to ensure atomicity of the clear operation across all shards
	// and to safely reset the global count.
	m.allLock()
	defer m.allUnlock()

	for i := 0; i < m.bucketTotal; i++ {
		m.buckets[i].innerMap = make(map[K]V)
	}
	atomic.StoreInt32(&m.count, 0)
}

// Len returns map items total
func (m *SafeMap[K, V]) Len() int {
	return int(atomic.LoadInt32(&m.count))
}

// IsEmpty returns true if map is empty
func (m *SafeMap[K, V]) IsEmpty() bool {
	return atomic.LoadInt32(&m.count) == 0
}

// GetOrSet returns the existing value for the key if present.
// Otherwise, it stores and returns the given value.
// The loaded result is true if the value was loaded, false if stored.
func (m *SafeMap[K, V]) GetOrSet(key K, val V) (V, bool) {
	index := m.hashIndex(key)
	bucket := m.buckets[index]
	bucket.Lock()
	defer bucket.Unlock()

	if existingVal, ok := bucket.innerMap[key]; ok {
		return existingVal, true
	}

	bucket.innerMap[key] = val
	atomic.AddInt32(&m.count, 1)
	return val, false
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, the iteration stops.
//
// Note: This method uses a snapshot strategy per bucket. It does not hold locks
// while calling f, preventing deadlocks if f calls other map methods.
// However, this means the iteration provides weak consistency: updates occurring
// during iteration may or may not be observed.
func (m *SafeMap[K, V]) Range(f func(k K, v V) bool) {
	for i := 0; i < m.bucketTotal; i++ {
		bucket := m.buckets[i]

		// Snapshot the bucket content
		bucket.RLock()
		keys := make([]K, 0, len(bucket.innerMap))
		vals := make([]V, 0, len(bucket.innerMap))
		for k, v := range bucket.innerMap {
			keys = append(keys, k)
			vals = append(vals, v)
		}
		bucket.RUnlock()

		// Iterate over the snapshot
		for j := range keys {
			if !f(keys[j], vals[j]) {
				return
			}
		}
	}
}

// Resize resizes the map to the new bucket count.
// This operation is expensive as it requires locking the entire map and re-hashing all keys.
func (m *SafeMap[K, V]) Resize(newBucketCount int) {
	if newBucketCount <= 0 {
		newBucketCount = defaultBucketCount
	}
	if newBucketCount > maxBucketCount {
		newBucketCount = maxBucketCount
	}

	m.allLock()
	// Do not use defer m.allUnlock() because m.buckets will change
	// and we need to unlock the OLD buckets, not the new ones.

	if m.bucketTotal == newBucketCount {
		m.allUnlock()
		return
	}

	// Create new buckets
	newBuckets := make([]*bucketMap[K, V], newBucketCount)
	for i := 0; i < newBucketCount; i++ {
		newBuckets[i] = &bucketMap[K, V]{innerMap: make(map[K]V)}
	}

	// Keep reference to old buckets and total for unlocking and moving
	oldBuckets := m.buckets
	oldBucketTotal := m.bucketTotal

	// Re-hash and move all data
	for i := 0; i < oldBucketTotal; i++ {
		oldBucket := oldBuckets[i]
		for k, v := range oldBucket.innerMap {
			// Recalculate hash index with new bucket count
			index := int(m.hashFunc(k) & uint64(newBucketCount-1))
			newBuckets[index].innerMap[k] = v
		}
	}

	// Replace buckets
	m.buckets = newBuckets
	m.bucketTotal = newBucketCount

	// Unlock old buckets
	for i := 0; i < oldBucketTotal; i++ {
		oldBuckets[i].Unlock()
	}
}
