package safemap

import (
	"fmt"
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

func TestMap(t *testing.T) {
	var m sync.Map
	// m.Store("key", "value")
	value, ok := m.Swap("key", "value1")
	fmt.Println(value, ok)
}
