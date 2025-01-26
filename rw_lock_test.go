package safemap

import (
	"sync"
	"testing"
)

func TestRwLock_Get(t *testing.T) {
	lock := NewRwLock[string, int]()
	lock.Set("foo", 42)

	val, ok := lock.Get("foo")
	if !ok || val != 42 {
		t.Errorf("Get() = %v, %v, want %v, %v", val, ok, 42, true)
	}

	val, ok = lock.Get("bar")
	if ok || val != 0 {
		t.Errorf("Get() = %v, %v, want %v, %v", val, ok, 0, false)
	}
}

func TestRwLock_Set(t *testing.T) {
	lock := NewRwLock[string, int]()
	lock.Set("foo", 42)

	val, ok := lock.Get("foo")
	if !ok || val != 42 {
		t.Errorf("Set() failed, Get() = %v, %v, want %v, %v", val, ok, 42, true)
	}

	// Test overwrite
	lock.Set("foo", 100)
	val, ok = lock.Get("foo")
	if !ok || val != 100 {
		t.Errorf("Set() overwrite failed, Get() = %v, %v, want %v, %v", val, ok, 100, true)
	}
}

func TestRwLock_Delete(t *testing.T) {
	lock := NewRwLock[string, int]()
	lock.Set("foo", 42)
	lock.Delete("foo")

	val, ok := lock.Get("foo")
	if ok || val != 0 {
		t.Errorf("Delete() failed, Get() = %v, %v, want %v, %v", val, ok, 0, false)
	}

	// Delete non-existent key
	lock.Delete("bar")
}

func TestRwLock_GetAndDelete(t *testing.T) {
	lock := NewRwLock[string, int]()
	lock.Set("foo", 42)

	val, loaded := lock.GetAndDelete("foo")
	if !loaded || val != 42 {
		t.Errorf("GetAndDelete() = %v, %v, want %v, %v", val, loaded, 42, true)
	}

	val, ok := lock.Get("foo")
	if ok || val != 0 {
		t.Errorf("GetAndDelete() failed, Get() = %v, %v, want %v, %v", val, ok, 0, false)
	}

	// Test non-existent key
	val, loaded = lock.GetAndDelete("bar")
	if loaded || val != 0 {
		t.Errorf("GetAndDelete() = %v, %v, want %v, %v", val, loaded, 0, false)
	}
}

func TestRwLock_GetOrSet(t *testing.T) {
	lock := NewRwLock[string, int]()

	// Test setting new value
	val, loaded := lock.GetOrSet("foo", 42)
	if loaded || val != 42 {
		t.Errorf("GetOrSet() = %v, %v, want %v, %v", val, loaded, 42, false)
	}

	// Test getting existing value
	val, loaded = lock.GetOrSet("foo", 100)
	if !loaded || val != 42 {
		t.Errorf("GetOrSet() = %v, %v, want %v, %v", val, loaded, 42, true)
	}
}

func TestRwLock_Len(t *testing.T) {
	lock := NewRwLock[string, int]()
	if lock.Len() != 0 {
		t.Errorf("Len() = %v, want %v", lock.Len(), 0)
	}

	lock.Set("foo", 42)
	if lock.Len() != 1 {
		t.Errorf("Len() = %v, want %v", lock.Len(), 1)
	}

	lock.Delete("foo")
	if lock.Len() != 0 {
		t.Errorf("Len() = %v, want %v", lock.Len(), 0)
	}
}

func TestRwLock_Range(t *testing.T) {
	lock := NewRwLock[string, int]()
	lock.Set("foo", 42)
	lock.Set("bar", 100)

	var count int
	lock.Range(func(key string, val int) bool {
		count++
		return true
	})

	if count != 2 {
		t.Errorf("Range() visited %v items, want %v", count, 2)
	}

	// Test early exit
	count = 0
	lock.Range(func(key string, val int) bool {
		count++
		return false
	})

	if count != 1 {
		t.Errorf("Range() visited %v items, want %v", count, 1)
	}
}

func TestRwLock_Concurrent(t *testing.T) {
	lock := NewRwLock[string, int]()
	var wg sync.WaitGroup

	// Concurrent writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			lock.Set("foo", i)
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lock.Get("foo")
		}()
	}

	wg.Wait()

	// Final value should be one of the writes
	val, ok := lock.Get("foo")
	if !ok {
		t.Errorf("Concurrent Get() failed")
	}
	if val < 0 || val >= 100 {
		t.Errorf("Concurrent Set() failed, got %v, want value between 0 and 99", val)
	}
}
