package safemap

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSafeMap(t *testing.T) {
	_, err := New[string, string]()
	assert.ErrorIs(t, err, ErrMissingHashFunc)

	m, err := New[string, string](HashstrKeyFunc())
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func BenchmarkOnlySetSafeMap(b *testing.B) {
	m, _ := New[string, string](HashstrKeyFunc())
	m.Set("hello", "world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set("hello", "world")
	}
}

func BenchmarkOnlySetSyncMap(b *testing.B) {
	var m sync.Map
	m.Store("hello", "world")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store("hello", "world")
	}
}

func BenchmarkOnlyGetSafeMap(b *testing.B) {
	m, _ := New[string, string](HashstrKeyFunc())
	m.Set("hello", "Hello World!!!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get("hello")
	}
}

func BenchmarkOnlyGetSyncMap(b *testing.B) {
	var m sync.Map
	m.Store("hello", "Hello World!!!")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Load("hello")
	}
}

func TestSafeMapLen(t *testing.T) {
	safeMap, _ := New[string, int](HashstrKeyFunc())
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
	m, _ := New[string, string](HashstrKeyFunc())
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
	m, _ := New[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

	// Test getting a non-existent key
	val, loaded := m.GetOrSet("key1", 42)
	assert.Equal(t, 42, val)
	assert.False(t, loaded)

	// Test getting an existing key
	val, loaded = m.GetOrSet("key1", 100)
	assert.Equal(t, 42, val)
	assert.True(t, loaded)
}

func TestIsEmpty(t *testing.T) {
	m, _ := New[string, string](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

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
	m, _ := New[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

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
	m, _ := New[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))

	// Concurrent set and get operations
	wg := sync.WaitGroup{}
	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := strconv.Itoa(n % 100)
			m.Set(key, n)
			val, exists := m.Get(key)
			assert.True(t, exists)
			assert.Equal(t, n, val)
		}(i)
	}
	wg.Wait()

	// Verify final map state
	assert.True(t, m.Len() == 100)
}

func BenchmarkSafeMapClear(b *testing.B) {
	m, _ := New[string, int](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))
	for i := 0; i < 1000; i++ {
		m.Set(strconv.Itoa(i), i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Clear()
	}
}
