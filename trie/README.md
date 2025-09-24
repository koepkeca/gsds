# Trie

A high-performance, thread-safe trie (prefix tree) implementation in Go that follows Go's concurrency via communication philosophy

## Introduction

This trie implementation provides fast prefix-based operations with complete thread safety through Go's channel-based concurrency model. Each trie instance runs its own goroutine and communicates via channels, eliminating the need for mutexes and the associated complexity of shared memory synchronization.

### Features

- **Thread-Safe**: Concurrent operations are handled safely through channels
- **High Performance**: Sub-microsecond operations for basic Insert/Get/Delete
- **Unicode Support**: Full UTF-8 string support for keys
- **Flexible Insert Modes**: Support for both append and replace semantics
- **Prefix Search**: Efficient prefix-based key searching with lexical ordering
- **Resource Management**: Proper cleanup with context cancellation
- **Memory Efficient**: Optimized search operations with minimal allocations

## Installation

```bash
go get github.com/koepkeca/gsds
```

## Usage

### Basic Operations

```go
package main

import (
    "fmt"
    "github.com/koepkeca/gsds/trie"
)

func main() {
    // Create a new trie
    tr := trie.New()
    defer tr.Close() // Always close to clean up the goroutine

    // Insert data
    err := tr.Insert("hello", []interface{}{"world", "there"})
    if err != nil {
        panic(err)
    }

    // Get data
    data, err := tr.Get("hello")
    if err != nil {
        panic(err)
    }
    fmt.Println(data) // [world there]

    // Check if key exists
    exists, err := tr.Exists("hello")
    if err != nil {
        panic(err)
    }
    fmt.Println(exists) // true

    // Delete a key
    deleted, err := tr.Delete("hello")
    if err != nil {
        panic(err)
    }
    fmt.Println(deleted) // true
}
```

### Replace Mode

By default, multiple inserts to the same key will append data. You can enable replace mode to overwrite existing data instead:

```go
tr := trie.New()
defer tr.Close()

// Default behavior - append mode
tr.Insert("key", []interface{}{"first"})
tr.Insert("key", []interface{}{"second"})
data, _ := tr.Get("key")
fmt.Println(data) // [first second]

// Enable replace mode
tr.Replace = true
tr.Insert("key", []interface{}{"replaced"})
data, _ = tr.Get("key")
fmt.Println(data) // [replaced]

// Disable replace mode
tr.Replace = false
tr.Insert("key", []interface{}{"appended"})
data, _ = tr.Get("key")
fmt.Println(data) // [replaced appended]
```

### Key Search

```go
tr := trie.New()
defer tr.Close()

// Insert some data
tr.Insert("cat", []interface{}{"feline"})
tr.Insert("car", []interface{}{"vehicle"})
tr.Insert("card", []interface{}{"payment"})
tr.Insert("care", []interface{}{"concern"})

// Search for all keys starting with "car"
keys := tr.Search("car")
fmt.Println(keys) // [car card care]

// Search entire trie (empty prefix)
allKeys := tr.Search("")
fmt.Println(allKeys) // [car card care cat] (lexically sorted)
```

**Important Note**: The `Search` method returns a snapshot of keys at search time. Keys may be deleted by other goroutines between the `Search()` call and subsequent `Get()` operations on those keys. Callers should handle potential errors from `Get()` calls gracefully.

### Context Management

For long-running applications, you can create a trie with context for proper lifecycle management:

```go
ctx, cancel := context.WithCancel(context.Background())
defer cancel()

tr := trie.NewWithContext(ctx)
// Trie will automatically shut down when context is cancelled
// No need to call tr.Close() explicitly
```

### Unicode Support

The trie fully supports Unicode characters:

```go
tr := trie.New()
defer tr.Close()

tr.Insert("café", []interface{}{"coffee"})
tr.Insert("naïve", []interface{}{"innocent"})
tr.Insert("🌟star", []interface{}{"emoji"})

data, _ := tr.Get("café")
fmt.Println(data) // [coffee]
```

## Performance

Benchmark results on AMD Ryzen 9 5950X:

| Operation | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| Insert | 1,772 | 238 | 6 |
| Get | 1,813 | 128 | 2 |
| Delete | 3,134 | 224 | 3 |
| Exists | 1,873 | 128 | 2 |
| Search (prefix) | 827,132 | 187,886 | 4,811 |
| Search (empty prefix) | 4,880,822 | 1,106,670 | 24,098 |

## API Reference

### Types

```go
type T struct {
    Replace bool // If true, Insert replaces existing data instead of appending
    // ... private fields
}
```

### Functions

#### `New() *T`
Creates a new trie instance.

#### `NewWithContext(ctx context.Context) *T`
Creates a new trie instance that will shut down when the context is cancelled.

### Methods

#### `Insert(key string, data []interface{}) error`
Inserts data at the specified key. If `Replace` is false (default), data is appended to existing data. If `Replace` is true, existing data is replaced.

#### `Get(key string) ([]interface{}, error)`
Retrieves data associated with the key. Returns nil if key doesn't exist.

#### `Exists(key string) (bool, error)`
Checks if a key exists in the trie (has associated data).

#### `Delete(key string) (bool, error)`
Removes a key from the trie and cleans up unnecessary nodes. Returns true if the key was deleted.

#### `Search(prefix string) []string`
Returns all keys that start with the given prefix in lexical order. An empty prefix returns all keys in the trie. **Note**: This returns a snapshot of keys at search time - keys may be deleted by other goroutines between this call and subsequent operations.

#### `Close()`
Shuts down the trie's goroutine and closes channels. Must be called to prevent goroutine leaks.

## Notes

### Thread Safety
All operations are thread-safe and can be called concurrently from multiple goroutines. The implementation uses Go's channel-based concurrency model rather than mutexes.

### Race Conditions
The `Search` method returns keys that exist at search time, but those keys may be deleted by other goroutines before you can perform `Get` operations on them. This is documented behavior - callers should handle `Get` errors gracefully and expect that some keys returned by `Search` might no longer exist.

### Memory Management
- The trie automatically cleans up unnecessary nodes when keys are deleted
- Search operations are optimized to minimize memory allocations
- Each trie runs its own goroutine that must be cleaned up with `Close()`

### Performance Considerations
- Basic operations (Insert/Get/Delete/Exists) have sub-microsecond performance
- Search operations are more expensive as they traverse potentially large subtrees
- Random key patterns perform slower than sequential patterns due to memory locality
- Replace mode is faster than append mode for repeated updates to the same key

### Error Handling
Operations return errors for invalid inputs:
- Empty keys are not allowed
- Empty data arrays are not allowed for Insert/Replace

### Unicode Support
Keys are treated as UTF-8 strings and fully support Unicode characters including emojis and multi-byte characters.

## License

See the gsds module base directory LICENSE file for information.
