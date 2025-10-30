# Thread-Safe Stack for Go

A goroutine and channel-based thread-safe stack implementation that provides concurrent access without explicit locking mechanisms.

## Features

- **Thread-safe operations** without explicit mutexes
- **Context-aware cancellation** support
- **Graceful cleanup** with proper resource management
- **Generic interface{}** storage for maximum flexibility
- **Channel-based communication** ensuring safe concurrent access

## Installation

```bash
go get github.com/koepkeca/gsds
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/koepkeca/gsds/stack"
)

func main() {
    // Create a new stack
    s := stack.New()
    defer s.Close() // Always close to clean up the goroutine
    
    // Push some values
    s.Push(42)
    s.Push("hello")
    s.Push([]byte("world"))
    
    // Check length
    fmt.Printf("Stack length: %d\n", s.Len()) // Output: 3
    
    // Pop values (LIFO order)
    val := s.Pop()
    fmt.Printf("Popped: %v\n", val) // Output: [119 111 114 108 100]
    
    val = s.Pop()
    fmt.Printf("Popped: %v\n", val) // Output: hello
}
```

## API Reference

### Creating a Stack

```go
// Create a new stack with default context
s := stack.New()

// Create a stack with custom context (cancels when context is done)
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()
s := stack.NewWithContext(ctx)
```

### Operations

```go
// Push a value onto the stack
s.Push(value interface{})

// Pop a value from the stack (returns nil if empty)
value := s.Pop() // interface{}

// Get the current length of the stack
length := s.Len() // int64

// Close the stack and clean up resources
s.Close()
```

### Type Assertions

Since the stack stores `interface{}` values, you'll need to assert types when popping:

```go
s.Push(42)
if val, ok := s.Pop().(int); ok {
    fmt.Printf("Got integer: %d\n", val)
}

s.Push("hello")
if val, ok := s.Pop().(string); ok {
    fmt.Printf("Got string: %s\n", val)
}
```

## Concurrent Usage

This stack is designed for concurrent access across multiple goroutines:

```go
s := stack.New()
defer s.Close()

// Multiple goroutines can safely access the stack
var wg sync.WaitGroup

// Producer goroutine
wg.Add(1)
go func() {
    defer wg.Done()
    for i := 0; i < 100; i++ {
        s.Push(i)
    }
}()

// Consumer goroutine
wg.Add(1)
go func() {
    defer wg.Done()
    for i := 0; i < 100; i++ {
        val := s.Pop()
        if val != nil {
            fmt.Printf("Consumed: %v\n", val)
        }
    }
}()

wg.Wait()
```

## Context Cancellation

```go
ctx, cancel := context.WithCancel(context.Background())
s := stack.NewWithContext(ctx)

// Stack operations work normally
s.Push("data")

// Cancel the context to gracefully shut down
cancel()

// Stack operations will panic after cancellation
// s.Push("more data") // This will panic
```

## Performance Characteristics

This implementation prioritizes **correctness and safety** over raw performance:

| Operation | Time Complexity | Typical Performance |
|-----------|-----------------|-------------------|
| Push | O(1) amortized | ~1000-2000 ns |
| Pop | O(1) | ~1000-2000 ns |
| Len | O(1) | ~1000-2000 ns |

**Memory overhead**: Each operation creates temporary channels, resulting in higher memory usage than mutex-based alternatives.

### Benchmark Results

```
BenchmarkPush-8           1000000    1500 ns/op    200 B/op    5 allocs/op
BenchmarkPop-8            1000000    1400 ns/op    180 B/op    4 allocs/op
BenchmarkLen-8            2000000     800 ns/op    120 B/op    3 allocs/op
```

## When to Use This Library

### **Ideal Use Cases:**

- **Medium-throughput applications** (< 1M operations/second)
- **Complex concurrent patterns** where channel semantics are beneficial
- **Applications requiring context cancellation**
- **Scenarios where avoiding deadlocks is critical**
- **Educational purposes** for understanding channel-based concurrency
- **Applications where goroutine-per-operation overhead is acceptable**

### **Consider Alternatives When:**

- **High-frequency operations** (> 1M operations/second)
- **Memory-constrained environments**
- **Real-time systems** requiring predictable latency
- **High-performance computing** applications
- **Embedded systems** with limited resources

## Alternative Approaches and Trade-offs

### Channel-Based (This Library) vs. Mutex-Based

```go
// Channel-based: Safe by design
s := stack.New()
defer s.Close() // Automatic cleanup, no resource leaks

// Multiple goroutines - no deadlock possible
go func() { s.Push("data1") }()
go func() { s.Push("data2") }()

// Mutex-based: Requires careful programming
type MutexStack struct {
    mu    sync.Mutex
    items []interface{}
}

// Easy to introduce bugs:
func (s *MutexStack) BadExample() {
    s.mu.Lock()
    if someCondition {
        return // Forgot to unlock - deadlock!
    }
    s.mu.Unlock()
}
```

### Performance in Context

| Scenario | Channel Impact | Typical Bottleneck |
|----------|---------------|-------------------|
| **Web Server** | +1.5μs | Database: +1-10ms (1000x larger) |
| **Message Queue** | +1.5μs | Network I/O: +0.1-1ms (100x larger) |
| **File Processing** | +1.5μs | Disk I/O: +0.1-10ms (100x larger) |
| **API Gateway** | +1.5μs | HTTP Request: +10-100ms (10,000x larger) |

**Only consider alternatives for:**
- High-frequency trading systems (microsecond latency requirements)
- Game engines (60+ FPS with tight frame budgets)
- Embedded systems with severe memory constraints

**For 99% of applications**, the safety and maintainability benefits far outweigh the microsecond performance cost.

## Best Practices

1. **Always call Close()**: Use `defer s.Close()` to prevent goroutine leaks
2. **Type assertions**: Always check type assertions when popping values
3. **Context usage**: Use `NewWithContext()` for applications with lifecycle management
4. **Error handling**: Check for nil values when popping from potentially empty stacks
5. **Benchmarking**: Profile your specific use case to determine if this fits your performance requirements

## Example: Producer-Consumer with Context

```go
func producerConsumer() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    s := stack.NewWithContext(ctx)
    defer s.Close()
    
    // Producer
    go func() {
        for i := 0; ; i++ {
            select {
            case <-ctx.Done():
                return
            default:
                s.Push(fmt.Sprintf("item-%d", i))
                time.Sleep(100 * time.Millisecond)
            }
        }
    }()
    
    // Consumer
    go func() {
        for {
            select {
            case <-ctx.Done():
                return
            default:
                if val := s.Pop(); val != nil {
                    fmt.Printf("Processed: %v\n", val)
                }
                time.Sleep(150 * time.Millisecond)
            }
        }
    }()
    
    // Wait for context timeout
    <-ctx.Done()
    fmt.Println("Shutting down gracefully")
}
```

## License

MIT License - see module LICENSE file for details.
