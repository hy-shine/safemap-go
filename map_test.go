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

func TestResize(t *testing.T) {
	m := NewStringMap[string, int](WithBuckets[string](2)) // Start with 4 buckets (1<<2)
	count := 1000
	for i := 0; i < count; i++ {
		m.Set(fmt.Sprintf("key-%d", i), i)
	}

	// Verify initial state
	assert.Equal(t, 4, m.bucketTotal)
	assert.Equal(t, count, m.Len())

	// Resize to 16 buckets
	m.Resize(16)

	// Verify state after resize
	assert.Equal(t, 16, m.bucketTotal)
	assert.Equal(t, count, m.Len())

	// Verify data integrity
	for i := 0; i < count; i++ {
		val, ok := m.Get(fmt.Sprintf("key-%d", i))
		assert.True(t, ok)
		assert.Equal(t, i, val)
	}
}
