package safemap

import (
	"fmt"
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeMap(t *testing.T) {
	_, err := NewMap[string, string]()
	assert.ErrorIs(t, err, ErrMissingHashFunc)

	m, err := NewMap[string, string](HashStrKeyFunc())
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func TestNewStringSafeMap(t *testing.T) {
	m := NewStringMap[string, int]()
	assert.NotNil(t, m)
}

func TestNewInteger(t *testing.T) {
	m := NewIntegerMap[int, string]()
	assert.NotNil(t, m)
}

func TestInteger(t *testing.T) {
	m := NewIntegerMap[int, int]()
	m.Set(-1, 1)
	val, loaded := m.GetOrSet(-1, 1)
	assert.True(t, loaded)
	assert.Equal(t, val, 1)
}

func TestSafeMapLen(t *testing.T) {
	safeMap, _ := NewMap[string, int](HashStrKeyFunc())
	n := 1000000
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := strconv.Itoa(n % 10050)
			safeMap.Set(key, n)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 10050, safeMap.Len())

	// clear
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			safeMap.Clear()
		}()
	}
	wg.Wait()

	assert.Equal(t, 0, safeMap.Len())
}

func TestGetAndDelete(t *testing.T) {
	const N = 50000
	m, _ := NewMap[string, string](HashStrKeyFunc())
	for i := 0; i < N; i++ {
		m.Set(strconv.Itoa(i), "hello")
	}

	ch := make(chan struct{ key string }, 5)
	go func() {
		for r := range ch {
			val, exists := m.Get(r.key)
			assert.False(t, exists)
			assert.Equal(t, val, "")
		}
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, ok := m.GetAndDelete(strconv.Itoa(i))
			assert.True(t, ok)
			ch <- struct{ key string }{key: strconv.Itoa(i)}
		}(i)
	}
	wg.Wait()
	close(ch)
}

func TestGetOrSet(t *testing.T) {
	m, _ := NewMap[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

	const N = 300
	for i := 0; i < N; i++ {
		m.Set(strconv.Itoa(i), i)
	}

	var wg sync.WaitGroup
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			val, exists := m.GetOrSet(strconv.Itoa(n), n)
			if n < N {
				assert.True(t, exists)
				assert.Equal(t, val, n)
			} else {
				assert.False(t, exists)
				assert.Equal(t, val, n)
			}
		}(i)
	}
	wg.Wait()
}

func TestIsEmpty(t *testing.T) {
	m, _ := NewMap[string, string](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

	// Test empty map
	assert.True(t, m.IsEmpty())

	// Add an item
	m.Set("key", "value")
	assert.False(t, m.IsEmpty())

	// Delete the item
	m.Delete("key")
	assert.True(t, m.IsEmpty())
}

func TestRange(t *testing.T) {
	m, _ := NewMap[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

	// Populate the map
	testData := map[string]int{
		"key1": 10,
		"key2": 20,
		"key3": 30,
	}
	for k, v := range testData {
		m.Set(k, v)
	}

	// Track visited keys
	visited := make(map[string]int)
	m.Range(func(k string, v int) bool {
		visited[k] = v
		return true
	})

	// Verify all keys were visited
	assert.Equal(t, testData, visited)

	// Test early termination
	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return count < 2 // Stop after 2 iterations
	})
	assert.Equal(t, 2, count)
}

func TestConcurrentOperations(t *testing.T) {
	m, _ := NewMap[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

	// Concurrent set and get operations
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := strconv.Itoa(n)
			m.Set(key, n)
			val, exists := m.Get(key)
			assert.True(t, exists)
			assert.Equal(t, n, val)
		}(i)
	}
	wg.Wait()

	// Verify final map state
	assert.True(t, m.Len() == 1000)
}

func TestClear(t *testing.T) {
	m, _ := NewMap[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))
	for i := 0; i < 1000; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	m.Clear()
	assert.Equal(t, 0, m.Len())
}

func BenchmarkSafeMapClear(b *testing.B) {
	m, _ := NewMap[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))
	for i := 0; i < 1000; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Clear()
	}
}

func TestRangeDeadlock(t *testing.T) {
	m := NewStringMap[string, int]()
	m.Set("key1", 1)
	m.Set("key2", 2)

	// This should not deadlock
	// We modify the map inside the Range callback
	m.Range(func(k string, v int) bool {
		m.Set("key3", 3)
		m.Delete("key1")
		return true
	})

	val, ok := m.Get("key3")
	assert.True(t, ok)
	assert.Equal(t, 3, val)

	_, ok = m.Get("key1")
	assert.False(t, ok)
}

func TestRangeCoverage(t *testing.T) {
	m := NewStringMap[string, int]()
	count := 100
	for i := 0; i < count; i++ {
		m.Set(string(rune(i)), i)
	}

	visited := 0
	m.Range(func(k string, v int) bool {
		visited++
		return true
	})
	assert.Equal(t, count, visited)
}

// Edge Case Tests - Empty Map Operations
func TestEmptyMapDelete(t *testing.T) {
	m := NewStringMap[string, int]()
	// Should not panic
	m.Delete("nonexistent")
	assert.Equal(t, 0, m.Len())
}

func TestEmptyMapGetAndDelete(t *testing.T) {
	m := NewStringMap[string, int]()
	val, loaded := m.GetAndDelete("nonexistent")
	assert.False(t, loaded)
	assert.Equal(t, 0, val)
	assert.Equal(t, 0, m.Len())
}

func TestEmptyMapRange(t *testing.T) {
	m := NewStringMap[string, int]()
	count := 0
	m.Range(func(k string, v int) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count)
}

// Edge Case Tests - Idempotent Operations
func TestRepeatedDelete(t *testing.T) {
	m := NewStringMap[string, int]()
	m.Set("key", 1)
	assert.Equal(t, 1, m.Len())

	m.Delete("key")
	assert.Equal(t, 0, m.Len())

	// Repeated delete should be safe
	m.Delete("key")
	assert.Equal(t, 0, m.Len())
}

func TestRepeatedSet(t *testing.T) {
	m := NewStringMap[string, int]()
	m.Set("key", 1)
	assert.Equal(t, 1, m.Len())

	// Repeated set should not increase count
	m.Set("key", 2)
	assert.Equal(t, 1, m.Len())

	val, ok := m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, 2, val)
}

func TestRepeatedClear(t *testing.T) {
	m := NewStringMap[string, int]()
	m.Set("key", 1)

	m.Clear()
	assert.Equal(t, 0, m.Len())

	// Repeated clear should be safe
	m.Clear()
	assert.Equal(t, 0, m.Len())
}

// Edge Case Tests - Range Special Scenarios
func TestRangeDeleteCurrentKey(t *testing.T) {
	m := NewStringMap[string, int]()
	for i := 0; i < 10; i++ {
		m.Set(fmt.Sprintf("key-%d", i), i)
	}

	// Delete keys during range (snapshot strategy should handle this)
	m.Range(func(k string, v int) bool {
		m.Delete(k)
		return true
	})

	// All keys should be deleted
	assert.Equal(t, 0, m.Len())
}

func TestRangeDeleteAllKeys(t *testing.T) {
	m := NewStringMap[string, int]()
	m.Set("key1", 1)
	m.Set("key2", 2)
	m.Set("key3", 3)

	visited := 0
	m.Range(func(k string, v int) bool {
		visited++
		m.Clear()
		return true
	})

	assert.True(t, visited > 0)
	assert.Equal(t, 0, m.Len())
}

// Concurrent Edge Case Tests
func TestConcurrentGetOrSet(t *testing.T) {
	m := NewStringMap[string, int]()
	const numGoroutines = 100
	const key = "shared-key"

	wg := sync.WaitGroup{}
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			val, loaded := m.GetOrSet(key, n)
			// Either we set it, or someone else did
			assert.True(t, !loaded || val >= 0)
		}(i)
	}
	wg.Wait()

	// Key should exist and have exactly one value
	assert.Equal(t, 1, m.Len())
	val, ok := m.Get(key)
	assert.True(t, ok)
	assert.True(t, val >= 0 && val < numGoroutines)
}

func TestConcurrentDeleteAndGet(t *testing.T) {
	m := NewStringMap[string, int]()
	const numKeys = 1000

	// Populate map
	for i := 0; i < numKeys; i++ {
		m.Set(fmt.Sprintf("key-%d", i), i)
	}

	wg := sync.WaitGroup{}
	// Half goroutines delete, half read
	for i := 0; i < numKeys; i++ {
		wg.Add(2)
		key := fmt.Sprintf("key-%d", i)

		go func(k string) {
			defer wg.Done()
			m.Delete(k)
		}(key)

		go func(k string) {
			defer wg.Done()
			// May or may not find the key
			m.Get(k)
		}(key)
	}
	wg.Wait()

	assert.Equal(t, 0, m.Len())
}

func TestConcurrentRangeAndModify(t *testing.T) {
	m := NewStringMap[string, int]()
	for i := 0; i < 100; i++ {
		m.Set(fmt.Sprintf("key-%d", i), i)
	}

	wg := sync.WaitGroup{}

	// Start range in one goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.Range(func(k string, v int) bool {
			return true
		})
	}()

	// Modify map concurrently
	for i := 100; i < 200; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Set(fmt.Sprintf("key-%d", n), n)
		}(i)
	}

	wg.Wait()
	// Should not panic or deadlock
	assert.True(t, m.Len() >= 100)
}

// Options Edge Case Tests
func TestWithBucketsZero(t *testing.T) {
	m := NewStringMap[string, int](WithBuckets[string](0))
	// Should use default
	assert.Equal(t, defaultBucketCount, m.bucketTotal)
}

func TestWithBucketsExceedMax(t *testing.T) {
	m := NewStringMap[string, int](WithBuckets[string](20))
	// Should be capped
	assert.Equal(t, maxBucketCount, m.bucketTotal)
}

func TestMultipleWithBuckets(t *testing.T) {
	m := NewStringMap[string, int](
		WithBuckets[string](3),
		WithBuckets[string](5), // Should override
	)
	assert.Equal(t, 1<<5, m.bucketTotal)
}

// Extreme Scenario Tests
func TestExtremeHighConcurrency(t *testing.T) {
	m := NewStringMap[string, int]()
	const numGoroutines = 10000

	wg := sync.WaitGroup{}
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := fmt.Sprintf("key-%d", n%1000)
			m.Set(key, n)
			m.Get(key)
		}(i)
	}
	wg.Wait()

	assert.True(t, m.Len() <= 1000)
}

func TestHotKeyScenario(t *testing.T) {
	m := NewStringMap[string, int]()
	const numGoroutines = 1000
	const hotKey = "hot-key"

	wg := sync.WaitGroup{}
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Set(hotKey, n)
			m.Get(hotKey)
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 1, m.Len())
	val, ok := m.Get(hotKey)
	assert.True(t, ok)
	assert.True(t, val >= 0 && val < numGoroutines)
}

func TestLargeValueSize(t *testing.T) {
	m := NewStringMap[string, []byte]()
	largeValue := make([]byte, 1024*1024) // 1MB

	m.Set("large", largeValue)
	val, ok := m.Get("large")
	assert.True(t, ok)
	assert.Equal(t, len(largeValue), len(val))
}
