# safemap-go

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

## Overview

`safemap-go` is a **high-performance**, **thread-safe** generic map implementation for Go. It uses a **sharded locking mechanism** to achieve superior concurrent performance compared to `sync.Map` and single-lock solutions.

### Why SafeMap?

- 🚀 **6.1% faster** than `sync.Map` in concurrent write scenarios
- 💾 **61% less memory allocation** compared to `sync.Map`
- 🔒 **Thread-safe** with no data races
- 🧩 **Generic** type support (Go 1.18+)
- ⚡ **Optimized** for read-write mixed and write-intensive workloads

## Features

- ✅ Thread-safe concurrent operations
- ✅ High-performance sharded locking
- ✅ Generic type support
- ✅ Flexible custom hash functions
- ✅ Comprehensive API (Get, Set, Delete, Range, etc.)

## Installation

```bash
go get github.com/hy-shine/safemap-go
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/hy-shine/safemap-go"
)

func main() {
    // Create a string-to-int map
    m := safemap.NewStringMap[string, int]()
    
    // Set values
    m.Set("answer", 42)
    m.Set("pi", 3)
    
    // Get a value
    if val, exists := m.Get("answer"); exists {
        fmt.Println(val) // Output: 42
    }
    
    // Iterate over all entries
    m.Range(func(key string, val int) bool {
        fmt.Printf("%s: %d\n", key, val)
        return true // continue iteration
    })
    
    // Delete a key
    m.Delete("pi")
    
    // Check map size
    fmt.Println("Size:", m.Len()) // Output: Size: 1
}
```

## Usage Guide

### Creating Maps

#### String Keys (Most Common)

```go
// Simplest way - uses default hash function
m := safemap.NewStringMap[string, int]()
```

#### Integer Keys

```go
// For integer keys
m := safemap.NewIntegerMap[int, string]()
```

#### Custom Types with Custom Hash Function

```go
// Define custom hash function
customHash := safemap.WithHashFunc(func(key MyType) uint64 {
    // Your custom hash logic
    return hash(key)
})

m, err := safemap.NewMap[MyType, ValueType](customHash)
if err != nil {
    log.Fatal(err)
}
```

### Configuring Bucket Count

```go
// Create map with 128 buckets (1<<7)
m := safemap.NewStringMap[string, int](
    safemap.WithBuckets[string](7), // 2^7 = 128 buckets
)
```

**Bucket Selection Guide:**

- Default (32 buckets): Good for most use cases
- 64-128 buckets: High concurrency scenarios
- 256+ buckets: Extreme concurrency (10,000+ goroutines)

### Basic Operations

```go
m := safemap.NewStringMap[string, int]()

// Set a value
m.Set("key", 100)

// Get a value
val, exists := m.Get("key")

// Delete a key
m.Delete("key")

// Get and delete atomically
val, loaded := m.GetAndDelete("key")

// Get existing value or set new one
val, loaded := m.GetOrSet("key", 200)

// Clear all entries
m.Clear()

// Check if empty
if m.IsEmpty() {
    fmt.Println("Map is empty")
}

// Get size
size := m.Len()
```

### Iteration with Range

```go
m := safemap.NewStringMap[string, int]()
m.Set("a", 1)
m.Set("b", 2)
m.Set("c", 3)

// Iterate over all entries
m.Range(func(key string, val int) bool {
    fmt.Printf("%s: %d\n", key, val)
    return true // return false to stop iteration
})
```

## Performance

### Benchmark Results

```
goos: darwin
goarch: arm64
cpu: Apple M1 Pro

# Concurrent Read Performance
Benchmark_Concurrent_Get_SafeMap-8      8167600    442.0 ns/op    24 B/op    1 allocs/op
Benchmark_Concurrent_Get_SyncMap-8      8375133    436.7 ns/op    24 B/op    1 allocs/op

# Concurrent Write Performance (SafeMap wins!)
Benchmark_Concurrent_Set_SafeMap-8      7014656    510.7 ns/op    51 B/op    2 allocs/op
Benchmark_Concurrent_Set_SyncMap-8      6650262    543.9 ns/op   131 B/op    5 allocs/op
```

### Performance Comparison

| Metric | SafeMap | sync.Map | Improvement |
|--------|---------|----------|-------------|
| Concurrent Write Speed | 510.7 ns/op | 543.9 ns/op | **6.1% faster** |
| Memory per Write | 51 B/op | 131 B/op | **61% less** |
| Allocations per Write | 2 allocs/op | 5 allocs/op | **60% fewer** |
| Concurrent Read Speed | 442.0 ns/op | 436.7 ns/op | ~1% slower |

### When to Use SafeMap

- Read-write mixed workloads (50/50 or more writes)
- Write-intensive scenarios (>30% writes)
- Memory-sensitive applications
- Predictable performance requirements

## Design Principles

### Sharded Locking Mechanism

SafeMap divides the map into multiple **buckets** (default: 32), each with its own lock:

```
┌─────────────────────────────────────┐
│           SafeMap                   │
├──────────┬──────────┬──────────────┤
│ Bucket 0 │ Bucket 1 │ ... Bucket N │
│  🔒      │  🔒      │     🔒       │
│ {k1:v1}  │ {k2:v2}  │   {kN:vN}    │
└──────────┴──────────┴──────────────┘
```

**Benefits:**

- Operations on different buckets can proceed **concurrently**
- Reduces lock contention compared to single-lock solutions
- Better scalability with more CPU cores

## API Reference

### Core Methods

| Method | Description |
|--------|-------------|
| `Get(key K) (V, bool)` | Retrieve a value by key |
| `Set(key K, val V)` | Set a key-value pair |
| `Delete(key K)` | Remove a key |
| `GetAndDelete(key K) (V, bool)` | Atomically get and remove |
| `GetOrSet(key K, val V) (V, bool)` | Get existing or set new value |
| `Clear()` | Remove all entries |
| `Len() int` | Get number of entries |
| `IsEmpty() bool` | Check if map is empty |
| `Range(f func(K, V) bool)` | Iterate over entries |

## Other Concurrent Maps

This package also provides two additional concurrent map implementations:

### SyncMap

Generic wrapper around Go's `sync.Map`:

```go
import "github.com/hy-shine/safemap-go/maps"

m := maps.NewSyncMap[string, int]()
m.Set("key", 42)
val, _ := m.Get("key")
```

### RwMap

Generic map protected by a single `sync.RWMutex`:

```go
import "github.com/hy-shine/safemap-go/maps"

m := maps.NewRwMap[string, int]()
m.Set("key", 42)
val, _ := m.Get("key")
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

See [LICENSE](./LICENSE)
