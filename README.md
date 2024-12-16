# safemap-go

## Overview

safemap-go is a high-performance, concurrent-safe generic map implementation in Go. It inspires by [sync.Map](https://pkg.go.dev/sync#Map). It provides a thread-safe alternative to standard Go maps with enhanced features and optimized performance.

## Features

- 🔒 Thread-safe operations
- 🚀 High-performance concurrent access
- 🧩 Generic type support
- 🔍 Flexible key hashing
- 📊 Efficient locks mechanism

## Installation

**go get**

```bash
go get github.com/hy-shine/safemap-go
```

**import package**

```go
import "github.com/hy-shine/safemap-go"
```

## Usage

### Basic Usage

```go
import "github.com/hy-shine/safemap-go"

// Create a new SafeMap with string keys and integer values
m, err := safemap.New[string, int](
    safemap.HashstrKeyFunc(),
)
if err != nil {
    fmt.Println(err)
    return
}

// Set a value
m.Set("key", 42)

// Get a value
val, exists := m.Get("key")
if exists {
    fmt.Println(val) // Prints: 42
}

// Delete a key
m.Delete("key")
```

### Advanced Usage

```go
// Custom hash function
customHashFunc := safemap.WithHashFunc(func(key string) uint64 {
    // Implement custom hash logic
    return customHash(key)
})

m, err := safemap.New[string, int](
    customHashFunc,
    safemap.WithCap(6), // Set lock capacity: 1<<6
)

var keys []string
var vals []int
// Iterate over map
m.Range(func(key string, value int) bool {
    keys = append(keys, key)
    vals = append(vals, value)
    return true // continue iteration
})

fmt.Printf("Keys: %v\n", keys)
fmt.Printf("Vals: %v\n", vals)

// Clear the map
m.Clear()
```

## Methods

- `Get(key K) (val V, exists bool)`: Retrieve a value
- `Set(key K, val V)`: Set a value
- `Delete(key K)`: Remove a key
- `GetAndDelete(key K) (val V, loaded bool)`: Get and remove a value
- `GetOrSet(key K, val V) (V, bool)`: Get existing or set new value
- `Clear()`: Remove all entries
- `Len() int`: Get number of entries
- `IsEmpty() bool`: Check if map is empty
- `Range(f func(k K, val V) bool)`: Iterate over entries

## License

See [License](./LICENSE)
