package safemap

import (
	"strconv"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	m, err := New[string, string]()
	if err != nil {
		b := assert.ErrorIs(t, err, ErrMissingHashFunc)
		assert.True(t, b)
		if b {
			m, _ = New[string, string](WithHashFn(func(s string) uint64 { return Hashstr(s) }))
		}
	}
	assert.NotNil(t, m)
}

func BenchmarkOnlySetCSMAP(b *testing.B) {
	m, _ := New[string, string](WithHashFn(func(s string) uint64 { return Hashstr(s) }))
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

func TestCSMap(t *testing.T) {
	m, _ := New[string, string](WithHashFn(func(s string) uint64 { return Hashstr(s) }))
	n := 100000
	wg := sync.WaitGroup{}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(n int) {
			key := strconv.Itoa(n % 10050)
			m.Set(key, key)
			wg.Done()
		}(i)
	}
	wg.Wait()

	assert.Equal(t, 10050, m.Cap())
}
