package maps

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSyncMap(t *testing.T) {
	m := SyncMap[string, string]{}
	value, ok := m.Get("key")
	assert.False(t, ok)
	assert.Equal(t, "", value)
	value, ok = m.GetAndDelete("key")
	assert.False(t, ok)
	assert.Equal(t, "", value)

	m.Set("key", "value")
	value, ok = m.Get("key")
	assert.True(t, ok)
	assert.Equal(t, "value", value)

	value, ok = m.GetOrSet("key", "value1")
	assert.True(t, ok)
	assert.Equal(t, "value", value)

	value, ok = m.Swap("key", "value2")
	assert.True(t, ok)
	assert.Equal(t, "value", value)
	value, ok = m.Swap("key1", "value2")
	assert.False(t, ok)
	assert.Equal(t, "", value)
}

func TestSyncMapPointerInt(t *testing.T) {
	m := SyncMap[int, *int]{}

	value, ok := m.Get(1)
	assert.False(t, ok)
	assert.Nil(t, value)

	value, ok = m.GetAndDelete(1)
	assert.False(t, ok)
	assert.Nil(t, value)

	var val int = 1
	m.Set(1, &val)
	value, ok = m.Get(1)
	assert.True(t, ok)
	assert.Equal(t, 1, *value)
	val = 2
	value, ok = m.GetOrSet(1, &val)
	assert.True(t, ok)
	assert.Equal(t, 2, *value)
	value, ok = m.GetAndDelete(1)
	assert.True(t, ok)
	assert.Equal(t, 2, *value)
}

func TestSyncMapGet(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test getting non-existent key
	val, exists := m.Get("key1")
	if exists {
		t.Errorf("Expected non-existent key to return false, got %v", exists)
	}
	if val != 0 {
		t.Errorf("Expected zero value for non-existent key, got %v", val)
	}

	// Test getting existing key
	m.Set("key1", 42)
	val, exists = m.Get("key1")
	if !exists {
		t.Errorf("Expected key to exist")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %v", val)
	}
}

func TestSyncMapSet(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test setting a value
	m.Set("key1", 42)
	val, exists := m.Get("key1")
	if !exists {
		t.Errorf("Expected key to exist after setting")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %v", val)
	}

	// Test overwriting an existing value
	m.Set("key1", 100)
	val, exists = m.Get("key1")
	if !exists {
		t.Errorf("Expected key to exist after overwriting")
	}
	if val != 100 {
		t.Errorf("Expected value 100, got %v", val)
	}
}

func TestSyncMapDelete(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test deleting non-existent key
	m.Delete("key1") // Should not panic

	// Test deleting existing key
	m.Set("key1", 42)
	m.Delete("key1")
	val, exists := m.Get("key1")
	if exists {
		t.Errorf("Expected key to be deleted")
	}
	if val != 0 {
		t.Errorf("Expected zero value after deletion, got %v", val)
	}
}

func TestSyncMapGetAndDelete(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test getting and deleting non-existent key
	val, loaded := m.GetAndDelete("key1")
	if loaded {
		t.Errorf("Expected non-existent key to return false")
	}
	if val != 0 {
		t.Errorf("Expected zero value for non-existent key, got %v", val)
	}

	// Test getting and deleting existing key
	m.Set("key1", 42)
	val, loaded = m.GetAndDelete("key1")
	if !loaded {
		t.Errorf("Expected key to be loaded")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %v", val)
	}

	// Verify key is deleted
	_, exists := m.Get("key1")
	if exists {
		t.Errorf("Expected key to be deleted")
	}
}

func TestSyncMapRange(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Populate map
	testData := map[string]int{
		"key1": 10,
		"key2": 20,
		"key3": 30,
	}
	for k, v := range testData {
		m.Set(k, v)
	}

	// Test Range method
	foundKeys := make(map[string]int)
	m.Range(func(key string, value int) bool {
		foundKeys[key] = value
		return true
	})

	if len(foundKeys) != len(testData) {
		t.Errorf("Expected %d items, got %d", len(testData), len(foundKeys))
	}

	for k, v := range testData {
		if foundKeys[k] != v {
			t.Errorf("Mismatched value for key %s: expected %d, got %d", k, v, foundKeys[k])
		}
	}
}

func TestSyncMapGetOrSet(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test GetOrSet for non-existent key
	val, loaded := m.GetOrSet("key1", 42)
	if loaded {
		t.Errorf("Expected not loaded for new key")
	}
	if val != 42 {
		t.Errorf("Expected value 42, got %v", val)
	}

	// Test GetOrSet for existing key
	val, loaded = m.GetOrSet("key1", 100)
	if !loaded {
		t.Errorf("Expected loaded for existing key")
	}
	if val != 42 {
		t.Errorf("Expected original value 42, got %v", val)
	}
}

func TestSyncMapSwap(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test Swap for non-existent key
	prev, loaded := m.Swap("key1", 42)
	if loaded {
		t.Errorf("Expected not loaded for new key")
	}
	if prev != 0 {
		t.Errorf("Expected zero value for non-existent key, got %v", prev)
	}

	// Test Swap for existing key
	prev, loaded = m.Swap("key1", 100)
	if !loaded {
		t.Errorf("Expected loaded for existing key")
	}
	if prev != 42 {
		t.Errorf("Expected previous value 42, got %v", prev)
	}
}

func TestSyncMapCompareAndDelete(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test CompareAndDelete for non-existent key
	deleted := m.CompareAndDelete("key1", 42)
	if deleted {
		t.Errorf("Expected false for non-existent key")
	}

	// Test CompareAndDelete with incorrect value
	m.Set("key1", 42)
	deleted = m.CompareAndDelete("key1", 100)
	if deleted {
		t.Errorf("Expected false for mismatched value")
	}

	// Test successful CompareAndDelete
	deleted = m.CompareAndDelete("key1", 42)
	if !deleted {
		t.Errorf("Expected true for matching value")
	}

	// Verify key is deleted
	_, exists := m.Get("key1")
	if exists {
		t.Errorf("Expected key to be deleted")
	}
}

func TestSyncMapCompareAndSwap(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Test CompareAndSwap for non-existent key
	swapped := m.CompareAndSwap("key1", 0, 42)
	if swapped {
		t.Errorf("Expected false for non-existent key with zero value")
	}

	m.Set("key1", 42)
	// Test CompareAndSwap with incorrect old value
	swapped = m.CompareAndSwap("key1", 100, 200)
	if swapped {
		t.Errorf("Expected false for mismatched old value")
	}

	// Test successful CompareAndSwap
	swapped = m.CompareAndSwap("key1", 42, 100)
	if !swapped {
		t.Errorf("Expected true for matching old value")
	}

	// Verify new value
	val, exists := m.Get("key1")
	if !exists || val != 100 {
		t.Errorf("Expected value 100, got %v", val)
	}
}

// Concurrent Tests
func TestSyncMapConcurrentOperations(t *testing.T) {
	m := &SyncMap[string, int]{}

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Set(string(rune('a'+n%26)), n)
		}(i)
	}
	wg.Wait()

	// Verify all writes
	count := 0
	m.Range(func(key string, value int) bool {
		count++
		return true
	})
	if count != 26 { // max 26 unique keys due to a-z
		t.Errorf("Expected 26 unique keys, got %d", count)
	}

	// Concurrent reads and writes
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			m.Set(string(rune('a'+n%26)), n)
			m.Get(string(rune('a' + n%26)))
			m.Delete(string(rune('a' + n%26)))
		}(i)
	}
	wg.Wait()
}
