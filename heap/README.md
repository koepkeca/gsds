# Heap - Thread-Safe Heap Wrapper

A Go package that provides a thread-safe wrapper around Go's standard `container/heap` interface.

## Overview

This package wraps any `heap.Interface` implementation and makes it safe for concurrent access by serializing all operations through a single goroutine. This eliminates the need for external synchronization when multiple goroutines need to access the same heap.

## Features

- **Thread-safe**: All operations are automatically synchronized
- **Simple API**: Familiar heap operations (Push, Pop, Remove, Len)
- **Generic**: Works with any `heap.Interface` implementation
- **No external locks needed**: Internal serialization handles concurrency

## Installation

```bash
go get github.com/koepkeca/gsds
```

## Usage

### Basic Example

```go
package main

import (
    "container/heap"
    "fmt"

	"github.com/koepkeca/gsds/heap"
)

// IntHeap implements heap.Interface
type IntHeap []int

func (h IntHeap) Len() int           { return len(h) }
func (h IntHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h IntHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *IntHeap) Push(x interface{}) {
    *h = append(*h, x.(int))
}
func (h *IntHeap) Pop() interface{} {
    old := *h
    n := len(old)
    x := old[n-1]
    *h = old[0 : n-1]
    return x
}

func main() {
    // Create the underlying heap
    ih := &IntHeap{2, 1, 5}
    
    // Wrap it in a thread-safe heap
    h := heap.New(ih)
    defer h.Close() // Always close to prevent goroutine leaks
    
    // Push elements
    h.Push(3)
    h.Push(4)
    
    // Get the length
    fmt.Println("Heap size:", h.Len())
    
    // Pop elements (will be in sorted order for min-heap)
    fmt.Println(h.Pop()) // 1
    fmt.Println(h.Pop()) // 2
    fmt.Println(h.Pop()) // 3
}
```

### Using Context for Automatic Cleanup

```go
func main() {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    
    ih := &IntHeap{}
    h := heap.NewWithContext(ctx, ih)
    // No need to call Close() - heap automatically shuts down when context expires
    
    h.Push(1)
    h.Push(2)
    
    // Heap will automatically clean up after 5 seconds or when context is cancelled
}
```

### Concurrent Usage

```go
func main() {
    ih := &IntHeap{}
    h := heap.New(ih)
    defer h.Close()
    
    // Multiple goroutines can safely access the heap
    var wg sync.WaitGroup
    
    // Producer goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(val int) {
            defer wg.Done()
            h.Push(val)
        }(i)
    }
    
    // Consumer goroutines
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            if h.Len() > 0 {
                fmt.Println(h.Pop())
            }
        }()
    }
    
    wg.Wait()
}
```

## API Reference

### `New(i heap.Interface) *H`

Creates a new thread-safe heap wrapper. The provided `heap.Interface` is initialized immediately.

**Important**: Always call `Close()` when done to prevent goroutine leaks.

### `NewWithContext(ctx context.Context, i heap.Interface) *H`

Creates a new thread-safe heap wrapper that will automatically shut down when the context is cancelled.
If ctx is nil, a background context is used. This is useful for automatic cleanup based on timeouts or
cancellation signals.

### `Push(x interface{}) bool`

Adds an element to the heap. Returns `true` when the operation completes successfully.

### `Pop() interface{}`

Removes and returns the minimum element from the heap (for a min-heap).

### `Remove(idx int) interface{}`

Removes and returns the element at the specified index.

### `Len() int`

Returns the number of elements currently in the heap.

### `Close()`

Shuts down the heap's internal goroutine. Must be called to prevent goroutine leaks when not using
context-based lifecycle management.

## Implementation Details

- All operations are serialized through a channel to a single goroutine
- Each operation blocks until it completes
- The underlying `heap.Interface` is never accessed directly outside the serialization goroutine
- Operations are processed in FIFO order

## Performance Considerations

- All operations have the overhead of channel communication
- For high-throughput scenarios, consider batching operations
- Best suited for scenarios where thread-safety is more important than raw performance

## License

MIT License - See LICENSE in module root
