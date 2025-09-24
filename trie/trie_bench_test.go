package trie

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

// Helper function to generate random strings
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// Helper function to generate sequential keys
func generateSequentialKey(i int) string {
	return fmt.Sprintf("key_%06d", i)
}

// Helper function to generate word-like keys
func generateWordKey(i int) string {
	prefixes := []string{"user", "item", "order", "product", "service"}
	return fmt.Sprintf("%s_%d", prefixes[i%len(prefixes)], i)
}

func BenchmarkInsert(b *testing.B) {
	tr := New()
	defer tr.Close()

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = generateSequentialKey(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Insert(keys[i], []interface{}{i})
	}
}

func BenchmarkInsertReplace(b *testing.B) {
	tr := New()
	defer tr.Close()
	tr.Replace = true

	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = generateSequentialKey(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Insert(keys[i], []interface{}{i})
	}
}

func BenchmarkInsertRandomKeys(b *testing.B) {
	tr := New()
	defer tr.Close()

	rand.Seed(time.Now().UnixNano())
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = generateRandomString(10)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Insert(keys[i], []interface{}{i})
	}
}

func BenchmarkGet(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with data
	keys := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		keys[i] = generateSequentialKey(i)
		tr.Insert(keys[i], []interface{}{i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Get(keys[i%10000])
	}
}

func BenchmarkGetMiss(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with data
	for i := 0; i < 10000; i++ {
		tr.Insert(generateSequentialKey(i), []interface{}{i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Get("nonexistent_" + strconv.Itoa(i))
	}
}

func BenchmarkExists(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with data
	keys := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		keys[i] = generateSequentialKey(i)
		tr.Insert(keys[i], []interface{}{i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Exists(keys[i%10000])
	}
}

func BenchmarkDelete(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with data
	keys := make([]string, b.N)
	for i := 0; i < b.N; i++ {
		keys[i] = generateSequentialKey(i)
		tr.Insert(keys[i], []interface{}{i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Delete(keys[i])
	}
}

func BenchmarkSearch(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with word-like keys for prefix searching
	for i := 0; i < 10000; i++ {
		tr.Insert(generateWordKey(i), []interface{}{i})
	}

	prefixes := []string{"user", "item", "order", "product", "service"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Search(prefixes[i%len(prefixes)])
	}
}

func BenchmarkSearchEmpty(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with data
	for i := 0; i < 10000; i++ {
		tr.Insert(generateWordKey(i), []interface{}{i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Search("")
	}
}

// Benchmark concurrent operations
func BenchmarkConcurrentInserts(b *testing.B) {
	tr := New()
	defer tr.Close()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tr.Insert(fmt.Sprintf("concurrent_%d_%d", b.N, i), []interface{}{i})
			i++
		}
	})
}

func BenchmarkConcurrentReads(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with data
	keys := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		keys[i] = generateSequentialKey(i)
		tr.Insert(keys[i], []interface{}{i})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			tr.Get(keys[i%10000])
			i++
		}
	})
}

func BenchmarkMixedOperations(b *testing.B) {
	tr := New()
	defer tr.Close()

	// Pre-populate with some data
	for i := 0; i < 1000; i++ {
		tr.Insert(generateSequentialKey(i), []interface{}{i})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		switch i % 4 {
		case 0:
			tr.Insert(generateSequentialKey(i), []interface{}{i})
		case 1:
			tr.Get(generateSequentialKey(i % 1000))
		case 2:
			tr.Exists(generateSequentialKey(i % 1000))
		case 3:
			if i > 1000 {
				tr.Delete(generateSequentialKey(i - 1000))
			}
		}
	}
}

// Benchmark memory allocation patterns
func BenchmarkInsertAppend(b *testing.B) {
	tr := New()
	defer tr.Close()

	key := "same_key"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tr.Insert(key, []interface{}{i})
	}
}

func BenchmarkInsertAppendVsReplace(b *testing.B) {
	b.Run("Append", func(b *testing.B) {
		tr := New()
		defer tr.Close()

		key := "same_key"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Insert(key, []interface{}{i})
		}
	})

	b.Run("Replace", func(b *testing.B) {
		tr := New()
		defer tr.Close()
		tr.Replace = true

		key := "same_key"

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Insert(key, []interface{}{i})
		}
	})
}

// Benchmark different key patterns
func BenchmarkKeyPatterns(b *testing.B) {
	b.Run("Sequential", func(b *testing.B) {
		tr := New()
		defer tr.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Insert(generateSequentialKey(i), []interface{}{i})
		}
	})

	b.Run("Random", func(b *testing.B) {
		tr := New()
		defer tr.Close()

		rand.Seed(time.Now().UnixNano())

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Insert(generateRandomString(10), []interface{}{i})
		}
	})

	b.Run("Prefixed", func(b *testing.B) {
		tr := New()
		defer tr.Close()

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			tr.Insert(generateWordKey(i), []interface{}{i})
		}
	})
}

// Benchmark trie depth impact
func BenchmarkTrieDepth(b *testing.B) {
	depths := []int{5, 10, 20, 50}

	for _, depth := range depths {
		b.Run(fmt.Sprintf("Depth%d", depth), func(b *testing.B) {
			tr := New()
			defer tr.Close()

			// Create keys of specific depth
			keys := make([]string, b.N)
			for i := 0; i < b.N; i++ {
				keys[i] = generateRandomString(depth)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				tr.Insert(keys[i], []interface{}{i})
			}
		})
	}
}

func BenchmarkSearchAux(b *testing.B) {
	// Setup a trie with some test data
	trie := New()
	words := []string{
		"apple", "application", "apply", "appreciate", "approach",
		"banana", "band", "bank", "bar", "base",
		"cat", "car", "card", "care", "careful", "carry",
		"dog", "door", "down", "draw", "drive",
		"elephant", "email", "end", "enter", "example",
	}

	for _, word := range words {
		trie.Insert(word, []interface{}{fmt.Sprintf("data-%s", word)})
	}

	b.Run("empty_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("")
		}
	})

	b.Run("single_char_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("c")
		}
	})

	b.Run("multi_char_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("app")
		}
	})

	b.Run("exact_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("apple")
		}
	})

	b.Run("no_match", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("xyz")
		}
	})
}

func BenchmarkSearchLargeDataset(b *testing.B) {
	// Create a larger dataset
	trie := New()

	// Insert 1000 keys with various prefixes
	for i := 0; i < 1000; i++ {
		key := fmt.Sprintf("key%04d", i)
		trie.Insert(key, []interface{}{fmt.Sprintf("value%d", i)})
	}

	b.Run("large_empty_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("")
		}
	})

	b.Run("large_common_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("key")
		}
	})

	b.Run("large_specific_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("key00")
		}
	})
}

func BenchmarkSearchDeepTrie(b *testing.B) {
	// Create a deep trie (long keys)
	trie := New()

	base := "verylongkeyprefix"
	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("%s%04d", base, i)
		trie.Insert(key, []interface{}{fmt.Sprintf("data%d", i)})
	}

	b.Run("deep_short_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("very")
		}
	})

	b.Run("deep_long_prefix", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_ = trie.Search("verylongkeyprefix")
		}
	})
}
