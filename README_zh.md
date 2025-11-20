# safemap-go

[![Go Version](https://img.shields.io/badge/Go-%3E%3D%201.18-blue)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

[English](README.md) | [简体中文](README_zh.md)

## 概述

`safemap-go` 是一个**高性能**、**线程安全**的 Go 语言泛型 Map 实现。它使用**分片锁机制**,相比 `sync.Map` 和单锁方案实现了卓越的并发性能。

### 为什么选择 SafeMap?

- 🚀 并发写场景比 `sync.Map` **快 6.1%**
- 💾 内存分配比 `sync.Map` **减少 61%**
- 🔒 **线程安全**,无数据竞态
- 🧩 **泛型**类型支持 (Go 1.18+)
- ⚡ 针对**读写混合**和**写密集**场景优化

## 特性

- ✅ 线程安全的并发操作
- ✅ 高性能分片锁机制
- ✅ 泛型类型支持
- ✅ 灵活的自定义哈希函数
- ✅ 完善的 API (Get, Set, Delete, Range 等)

## 安装

```bash
go get github.com/hy-shine/safemap-go
```

## 快速开始

```go
package main

import (
    "fmt"
    "github.com/hy-shine/safemap-go"
)

func main() {
    // 创建一个 string 到 int 的 map
    m := safemap.NewStringMap[string, int]()
    
    // 设置值
    m.Set("answer", 42)
    m.Set("pi", 3)
    
    // 获取值
    if val, exists := m.Get("answer"); exists {
        fmt.Println(val) // 输出: 42
    }
    
    // 遍历所有条目
    m.Range(func(key string, val int) bool {
        fmt.Printf("%s: %d\n", key, val)
        return true // 继续遍历
    })
    
    // 删除键
    m.Delete("pi")
    
    // 检查 map 大小
    fmt.Println("Size:", m.Len()) // 输出: Size: 1
}
```

## 使用指南

### 创建 Map

#### String 类型键 (最常用)

```go
// 最简单的方式 - 使用默认哈希函数
m := safemap.NewStringMap[string, int]()
```

#### Integer 类型键

```go
// 用于整数键
m := safemap.NewIntegerMap[int, string]()
```

#### 自定义类型与自定义哈希函数

```go
// 定义自定义哈希函数
customHash := safemap.WithHashFunc(func(key MyType) uint64 {
    // 你的自定义哈希逻辑
    return hash(key)
})

m, err := safemap.NewMap[MyType, ValueType](customHash)
if err != nil {
    log.Fatal(err)
}
```

### 配置 Bucket 数量

```go
// 创建拥有 128 个 bucket 的 map (1<<7)
m := safemap.NewStringMap[string, int](
    safemap.WithBuckets[string](7), // 2^7 = 128 buckets
)
```

**Bucket 选择指南:**

- 默认 (32 buckets): 适用于大多数场景
- 64-128 buckets: 高并发场景
- 256+ buckets: 极端并发场景 (10,000+ goroutines)

### 基本操作

```go
m := safemap.NewStringMap[string, int]()

// 设置值
m.Set("key", 100)

// 获取值
val, exists := m.Get("key")

// 删除键
m.Delete("key")

// 原子性地获取并删除
val, loaded := m.GetAndDelete("key")

// 获取已存在的值或设置新值
val, loaded := m.GetOrSet("key", 200)

// 清空所有条目
m.Clear()

// 检查是否为空
if m.IsEmpty() {
    fmt.Println("Map 为空")
}

// 获取大小
size := m.Len()
```

### 使用 Range 遍历

```go
m := safemap.NewStringMap[string, int]()
m.Set("a", 1)
m.Set("b", 2)
m.Set("c", 3)

// 遍历所有条目
m.Range(func(key string, val int) bool {
    fmt.Printf("%s: %d\n", key, val)
    return true // 返回 false 停止遍历
})
```

## 性能

### 基准测试结果

```
goos: darwin
goarch: arm64
cpu: Apple M1 Pro

# 并发读性能
Benchmark_Concurrent_Get_SafeMap-8      8167600    442.0 ns/op    24 B/op    1 allocs/op
Benchmark_Concurrent_Get_SyncMap-8      8375133    436.7 ns/op    24 B/op    1 allocs/op

# 并发写性能 (SafeMap 胜出!)
Benchmark_Concurrent_Set_SafeMap-8      7014656    510.7 ns/op    51 B/op    2 allocs/op
Benchmark_Concurrent_Set_SyncMap-8      6650262    543.9 ns/op   131 B/op    5 allocs/op
```

### 性能对比

| 指标 | SafeMap | sync.Map | 改进 |
|------|---------|----------|------|
| 并发写速度 | 510.7 ns/op | 543.9 ns/op | **快 6.1%** |
| 每次写入内存 | 51 B/op | 131 B/op | **减少 61%** |
| 每次写入分配次数 | 2 allocs/op | 5 allocs/op | **减少 60%** |
| 并发读速度 | 442.0 ns/op | 436.7 ns/op | 慢约 1% |

### 何时使用 SafeMap

- 读写混合场景 (50/50 或更多写入)
- 写密集场景 (>30% 写入)
- 内存敏感的应用
- 需要可预测性能的场景

## 设计原理

### 分片锁机制

SafeMap 将 map 分成多个 **bucket** (默认: 32),每个都有自己的锁:

```
┌─────────────────────────────────────┐
│           SafeMap                   │
├──────────┬──────────┬──────────────┤
│ Bucket 0 │ Bucket 1 │ ... Bucket N │
│  🔒      │  🔒      │     🔒       │
│ {k1:v1}  │ {k2:v2}  │   {kN:vN}    │
└──────────┴──────────┴──────────────┘
```

**优势:**

- 不同 bucket 上的操作可以**并发**进行
- 相比单锁方案减少锁竞争
- 随 CPU 核心数增加有更好的扩展性

## API 参考

### 核心方法

| 方法 | 描述 |
|------|------|
| `Get(key K) (V, bool)` | 通过键获取值 |
| `Set(key K, val V)` | 设置键值对 |
| `Delete(key K)` | 删除键 |
| `GetAndDelete(key K) (V, bool)` | 原子性地获取并删除 |
| `GetOrSet(key K, val V) (V, bool)` | 获取已存在的值或设置新值 |
| `Clear()` | 删除所有条目 |
| `Len() int` | 获取条目数量 |
| `IsEmpty() bool` | 检查 map 是否为空 |
| `Range(f func(K, V) bool)` | 遍历条目 |

## 其他并发 Map

本包还提供了两个额外的并发 map 实现:

### SyncMap

Go 标准库 `sync.Map` 的泛型封装:

```go
import "github.com/hy-shine/safemap-go/maps"

m := maps.NewSyncMap[string, int]()
m.Set("key", 42)
val, _ := m.Get("key")
```

### RwMap

使用单个 `sync.RWMutex` 保护的泛型 map:

```go
import "github.com/hy-shine/safemap-go/maps"

m := maps.NewRwMap[string, int]()
m.Set("key", 42)
val, _ := m.Get("key")
```

## 贡献

欢迎贡献!请随时提交 Pull Request。

## 许可证

查看 [LICENSE](./LICENSE)
