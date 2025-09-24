# Go Safe Data Structures

A collection of thread-safe data structures for Go, implemented using goroutines and channels for safe concurrent access.

## Overview

This module provides thread-safe implementations of common data structures that can be safely used across multiple goroutines without external synchronization. Each data structure runs its own goroutine and uses channels for communication, ensuring thread safety through Go's concurrency primitives.

## Features

- **Thread-safe**: All operations are safe for concurrent use
- **Context support**: All data structures support context-based cancellation
- **Memory efficient**: Optimized implementations using Go slices and maps
- **Clean API**: Simple, intuitive interfaces for each data structure
- **Proper cleanup**: Built-in cleanup methods to prevent goroutine leaks

## Data Structures

### Stack (`stack` package)
A Last-In-First-Out (LIFO) data structure.

**Key Methods:**
- `Push(v interface{})` - Add an item to the top of the stack
- `Pop() interface{}` - Remove and return the top item
- `Len() int64` - Get the current stack size
- `Close()` - Clean up resources

### Queue (`queue` package)
A First-In-First-Out (FIFO) data structure.

**Key Methods:**
- `Enqueue(v interface{})` - Add an item to the back of the queue
- `Dequeue() interface{}` - Remove and return the front item
- `Front() interface{}` - Peek at the front item without removing it
- `Back() interface{}` - Peek at the back item without removing it
- `Len() int64` - Get the current queue size
- `Close()` - Clean up resources

### Trie (`trie` package)
A prefix tree data structure for efficient string-based operations.

**Key Methods:**
- `Insert(k string, v []interface{}) error` - Insert data at a key
- `Get(k string) ([]interface{}, error)` - Retrieve data for a key
- `Exists(k string) (bool, error)` - Check if a key exists
- `Delete(k string) (bool, error)` - Remove a key and clean up nodes
- `Search(prefix string) []string` - Find all keys with given prefix
- `Close()` - Clean up resources

**Special Features:**
- `Replace` field: Controls whether inserts replace or append to existing data
- Lexically ordered search results
- Automatic cleanup of empty nodes on deletion

## Installation

```bash
go get github.com/koepkeca/gsds
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/koepkeca/gsds/stack"
    "github.com/koepkeca/gsds/queue"
    "github.com/koepkeca/gsds/trie"
)

func main() {
    // Create a stack
    s := stack.New()
    defer s.Close() // Always clean up!
    
    s.Push("hello")
    s.Push("world")
    fmt.Println(s.Pop()) // "world"
    
    // Create a queue
    q := queue.New()
    defer q.Close()
    
    q.Enqueue(1)
    q.Enqueue(2)
    fmt.Println(q.Dequeue()) // 1
    
    // Create a trie
    t := trie.New()
    defer t.Close()
    
    t.Insert("hello", []interface{}{"greeting"})
    data, _ := t.Get("hello")
    fmt.Println(data) // ["greeting"]
}
```

## Context Support

All data structures support context-based cancellation:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

s := stack.NewWithContext(ctx)
// Stack will automatically close when context is cancelled
```

## Memory Management

**Important**: Always call `Close()` on your data structures to prevent goroutine leaks. Each data structure spawns a goroutine that must be properly cleaned up.

```go
s := stack.New()
defer s.Close() // This is essential!
```

## Error Handling

Operations that can fail return appropriate errors:

```go
t := trie.New()
defer t.Close()

// This will return an error
err := t.Insert("", []interface{}{"data"}) // Empty key not allowed
if err != nil {
    log.Printf("Insert failed: %v", err)
}
```

## Thread Safety

All data structures are designed to be used safely from multiple goroutines:

```go
s := stack.New()
defer s.Close()

// Safe to use from multiple goroutines
go func() {
    s.Push("from goroutine 1")
}()

go func() {
    s.Push("from goroutine 2")
}()
```

## Performance Characteristics

- **Stack**: O(1) push/pop operations
- **Queue**: O(1) enqueue/dequeue operations  
- **Trie**: O(k) operations where k is key length
  - Insert/Get/Exists: O(k)
  - Delete: O(k) with automatic cleanup
  - Search: O(p + m) where p is prefix length and m is number of matches

## Contributing

Each data structure has its own package with detailed documentation. See the individual README files in each package directory for more specific usage examples and implementation details.

## Notes

- All data structures use `interface{}` for maximum flexibility
- Channel-based communication ensures thread safety
- Context support allows for graceful shutdown
- Proper resource cleanup prevents goroutine leaks
