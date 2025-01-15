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
	m, _ := NewSafeMap[string, string](HashStrKeyFunc())
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
	m, _ := NewSafeMap[string, string](HashStrKeyFunc())
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

func Benchmark_Concurrent_Get_SafeMap(b *testing.B) {
	m, _ := NewSafeMap[string, string](HashStrKeyFunc())
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

func Benchmark_Concurrent_Get_SyncMap(b *testing.B) {
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

func Benchmark_Concurrent_Get_SingleLock(b *testing.B) {
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

func Benchmark_Concurrent_Get_SingleRwLock(b *testing.B) {
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

func Benchmark_Concurrent_Set_SafeMap(b *testing.B) {
	m, _ := NewSafeMap[string, string](HashStrKeyFunc())
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Set(strconv.Itoa(n%5000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurrent_Set_SyncMap(b *testing.B) {
	m := sync.Map{}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Store(strconv.Itoa(n%5000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurrent_Set_SingleLock(b *testing.B) {
	m := singleLock[string, string]{m: make(map[string]string)}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Set(strconv.Itoa(n%5000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Concurrent_Set_SingleRwLock(b *testing.B) {
	m := singleRwLock[string, string]{m: make(map[string]string)}
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Set(strconv.Itoa(n%5000), data.val)
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket1_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](1))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket2_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](2))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket3_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](3))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket4_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](4))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket5_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](5))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket6_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](6))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket7_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](7))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket8_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](8))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}

func Benchmark_Bucket9_Get_SafeMap(b *testing.B) {
	m := NewSafeMapString[string, string](WithBuckets[string](9))
	ch := make(chan struct{}, b.N)
	for i := 0; i < b.N; i++ {
		go func(n int) {
			m.Get(strconv.Itoa(n % 10000))
			ch <- struct{}{}
		}(i)
	}
	for i := 0; i < b.N; i++ {
		<-ch
	}
}
