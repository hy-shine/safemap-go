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
