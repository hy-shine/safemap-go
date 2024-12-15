package safemap

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	_, err := New[string, string]()
	assert.ErrorIs(t, err, ErrMissingHashFunc)

	fn := WithHashFunc(func(s string) uint64 { return Hashstr(s) })
	m, err := New[string, string](fn)
	assert.Nil(t, err)
	assert.NotNil(t, m)
}

func BenchmarkOnlySetCSMAP(b *testing.B) {
	m, _ := New[string, string](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))
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

func TestCSMapLen(t *testing.T) {
	safeMap, _ := New[string, int](
		WithHashFunc(func(s string) uint64 { return Hashstr(s) }))
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
	m, _ := New[string, string](WithHashFunc(func(s string) uint64 { return Hashstr(s) }))
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
