# Queue Package

A thread-safe First-In-First-Out (FIFO) queue implementation for Go.

## Overview

The queue package provides a concurrent-safe queue data structure that uses goroutines and channels to serialize all operations. This ensures safe access from multiple goroutines without requiring external synchronization.

## Features

- **Thread-safe**: All operations are safe for concurrent use
- **Context support**: Supports context-based cancellation
- **Simple API**: Clean, intuitive interface
- **Efficient**: O(1) enqueue and dequeue operations
- **Proper cleanup**: Built-in cleanup to prevent goroutine leaks

## Installation

```bash
go get github.com/koepkeca/gsds
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/koepkeca/gsds/queue"
)

func main() {
    q := queue.New()
    defer q.Close() // Always clean up!
    
    // Add items to the queue
    q.Enqueue("first")
    q.Enqueue("second")
    q.Enqueue("third")
    
    // Check the size
    fmt.Println(q.Len()) // 3
    
    // Peek at front and back
    fmt.Println(q.Front()) // "first"
    fmt.Println(q.Back())  // "third"
    
    // Remove items in FIFO order
    fmt.Println(q.Dequeue()) // "first"
    fmt.Println(q.Dequeue()) // "second"
    fmt.Println(q.Dequeue()) // "third"
}
```

## API Reference

### Creating a Queue

#### `New() *Q`
Creates a new thread-safe queue.

```go
q := queue.New()
defer q.Close()
```

#### `NewWithContext(ctx context.Context) *Q`
Creates a new queue that will automatically close when the context is cancelled.

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

q := queue.NewWithContext(ctx)
// Queue will automatically close after 5 seconds
```

### Operations

#### `Enqueue(v interface{})`
Adds an item to the back of the queue. This operation is O(1).

```go
q.Enqueue(42)
q.Enqueue("hello")
q.Enqueue(struct{ Name string }{"Alice"})
```

#### `Dequeue() interface{}`
Removes and returns the item at the front of the queue. Returns `nil` if the queue is empty. This operation is O(1).

```go
item := q.Dequeue()
if item != nil {
    fmt.Println("Got:", item)
}
```

#### `Front() interface{}`
Returns the item at the front of the queue without removing it. Returns `nil` if the queue is empty.

```go
next := q.Front()
if next != nil {
    fmt.Println("Next item:", next)
}
```

#### `Back() interface{}`
Returns the item at the back of the queue without removing it. Returns `nil` if the queue is empty.

```go
last := q.Back()
if last != nil {
    fmt.Println("Last item:", last)
}
```

#### `Len() int64`
Returns the current number of items in the queue.

```go
size := q.Len()
fmt.Printf("Queue has %d items\n", size)
```

#### `Close()`
Shuts down the queue's internal goroutine. Must be called to prevent goroutine leaks.

```go
q := queue.New()
defer q.Close() // Best practice: use defer
```

## Usage Examples

### Basic Queue Operations

```go
q := queue.New()
defer q.Close()

// Build up the queue
for i := 1; i <= 5; i++ {
    q.Enqueue(i)
}

// Process all items
for q.Len() > 0 {
    item := q.Dequeue()
    fmt.Println("Processing:", item)
}
```

### Producer-Consumer Pattern

```go
q := queue.New()
defer q.Close()

// Producer goroutine
go func() {
    for i := 0; i < 100; i++ {
        q.Enqueue(i)
        time.Sleep(10 * time.Millisecond)
    }
}()

// Consumer goroutine
go func() {
    for {
        if q.Len() > 0 {
            item := q.Dequeue()
            fmt.Println("Consumed:", item)
        }
        time.Sleep(50 * time.Millisecond)
    }
}()

time.Sleep(2 * time.Second)
```

### Context-Based Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())

q := queue.NewWithContext(ctx)

// Start some work
go func() {
    for i := 0; i < 1000; i++ {
        q.Enqueue(i)
        time.Sleep(10 * time.Millisecond)
    }
}()

// Cancel after 1 second
time.Sleep(1 * time.Second)
cancel() // Queue automatically closes
```

### Checking Queue State

```go
q := queue.New()
defer q.Close()

q.Enqueue("first")
q.Enqueue("second")

// Check what's coming next without removing it
if next := q.Front(); next != nil {
    fmt.Println("Next up:", next)
}

// Check what's at the end
if last := q.Back(); last != nil {
    fmt.Println("Last in line:", last)
}

// Get the size
fmt.Printf("Queue has %d items\n", q.Len())
```

## Performance Characteristics

- **Enqueue**: O(1) - Constant time to add an item
- **Dequeue**: O(1) - Constant time to remove an item
- **Front/Back**: O(1) - Constant time to peek
- **Len**: O(1) - Constant time to get size
- **Space**: O(n) - Linear space where n is the number of items

## Thread Safety

All operations are thread-safe and can be called from multiple goroutines simultaneously:

```go
q := queue.New()
defer q.Close()

// Safe to use from multiple goroutines
for i := 0; i < 10; i++ {
    go func(id int) {
        q.Enqueue(fmt.Sprintf("message from goroutine %d", id))
    }(i)
}

// Also safe to read from multiple goroutines
for i := 0; i < 5; i++ {
    go func() {
        for q.Len() > 0 {
            item := q.Dequeue()
            fmt.Println("Received:", item)
        }
    }()
}
```

## Important Notes

1. **Always call Close()**: Each queue spawns a goroutine that must be cleaned up. Use `defer q.Close()` immediately after creating a queue.

2. **Nil returns**: `Dequeue()`, `Front()`, and `Back()` return `nil` when the queue is empty. Always check for nil if your queue might be empty.

3. **Type assertions**: Since the queue uses `interface{}`, you'll need to assert types when retrieving items:
   ```go
   item := q.Dequeue()
   if str, ok := item.(string); ok {
       fmt.Println("Got string:", str)
   }
   ```

4. **Blocking operations**: All operations block until they complete. This ensures consistency but means you should avoid calling queue operations while holding other locks.

## Comparison with Channels

Go's built-in channels also provide FIFO semantics, but this queue offers:
- Dynamic sizing (no fixed buffer size)
- Peek operations (Front/Back)
- Length queries
- More explicit control over lifecycle

Use channels when you need Go's select statement or want blocking semantics. Use this queue when you need dynamic sizing and inspection capabilities.

## License

Part of the gsds (Go Safe Data Structures) module.
