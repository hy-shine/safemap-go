package safemap

import (
	"strconv"
	"sync"
	"testing"
)

type singleLock[K comparable, V any] struct {
	sync.Mutex
	m map[K]V
}

func (l *singleLock[K, V]) Get(key K) (V, bool) {
	l.Lock()
	val, b := l.m[key]
	l.Unlock()
	return val, b
}

func (l *singleLock[K, V]) Set(key K, val V) {
	l.Lock()
	l.m[key] = val
	l.Unlock()
}

func (l *singleLock[K, V]) Delete(key K) {
	l.Lock()
	delete(l.m, key)
	l.Unlock()
}

type singleRwLock[K comparable, V any] struct {
	sync.RWMutex
	m map[K]V
}

func (l *singleRwLock[K, V]) Get(key K) (V, bool) {
	l.RLock()
	val, b := l.m[key]
	l.RUnlock()
	return val, b
}

func (l *singleRwLock[K, V]) Set(key K, val V) {
	l.Lock()
	l.m[key] = val
	l.Unlock()
}

func (l *singleRwLock[K, V]) Delete(key K) {
	l.Lock()
	delete(l.m, key)
	l.Unlock()
}

var data = struct {
	key string
	val string
}{
	key: "hello",
	val: "world",
}

func Benchmark_Single_Get_SafeMap(b *testing.B) {
	m, _ := New[string, string](HashstrKeyFunc())
	m.Set(data.key, data.val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(data.key)
	}
}

func Benchmark_Single_Get_SyncMap(b *testing.B) {
	var m sync.Map
	m.Store(data.key, data.val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Load(data.key)
	}
}

func Benchmark_Single_Get_SingleLock(b *testing.B) {
	m := singleLock[string, string]{m: make(map[string]string)}
	m.Set(data.key, data.val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(data.key)
	}
}

func Benchmark_Single_Get_SingleRwLock(b *testing.B) {
	m := singleRwLock[string, string]{m: make(map[string]string)}
	m.Set(data.key, data.val)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(data.key)
	}
}

func Benchmark_Single_Set_SafeMap(b *testing.B) {
	m, _ := New[string, string](HashstrKeyFunc())
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(data.key, data.val)
	}
}

func Benchmark_Single_Set_SyncMap(b *testing.B) {
	var m sync.Map
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Store(data.key, data.val)
	}
}

func Benchmark_Single_Set_SingleLock(b *testing.B) {
	m := singleLock[string, string]{m: make(map[string]string)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(data.key, data.val)
	}
}

func Benchmark_Single_Set_SingleRwLock(b *testing.B) {
	m := singleRwLock[string, string]{m: make(map[string]string)}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(data.key, data.val)
	}
}

func Benchmark_Concurent_Get_SafeMap(b *testing.B) {
	m, _ := New[string, string](HashstrKeyFunc())
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			m.Get(data.key)
			ch <- struct{}{}
		}()
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Get_SyncMap(b *testing.B) {
	m := sync.Map{}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			m.Load(data.key)
			ch <- struct{}{}
		}()
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Get_SingleLock(b *testing.B) {
	m := singleLock[string, string]{m: make(map[string]string)}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			m.Get(data.key)
			ch <- struct{}{}
		}()
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Get_SingleRwLock(b *testing.B) {
	m := singleRwLock[string, string]{m: make(map[string]string)}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func() {
			m.Get(data.key)
			ch <- struct{}{}
		}()
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Set_SafeMap(b *testing.B) {
	m, _ := New[string, string](HashstrKeyFunc())
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Set(strconv.Itoa(n%1000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Set_SyncMap(b *testing.B) {
	m := sync.Map{}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Store(strconv.Itoa(n%1000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Set_SingleLock(b *testing.B) {
	m := singleLock[string, string]{m: make(map[string]string)}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Set(strconv.Itoa(n%1000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurent_Set_SingleRwLock(b *testing.B) {
	m := singleRwLock[string, string]{m: make(map[string]string)}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Set(strconv.Itoa(n%1000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}
