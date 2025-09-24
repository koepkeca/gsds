package trie

import (
	"fmt"
	"testing"
)

func TestSearch(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*T) // Function to set up the trie
		prefix   string
		expected []string
	}{
		{
			name: "empty trie with empty prefix",
			setup: func(trie *T) {
				// No setup needed - empty trie
			},
			prefix:   "",
			expected: []string{},
		},
		{
			name: "empty trie with non-empty prefix",
			setup: func(trie *T) {
				// No setup needed - empty trie
			},
			prefix:   "hello",
			expected: nil,
		},
		{
			name: "single key - exact match",
			setup: func(trie *T) {
				trie.Insert("hello", []interface{}{"world"})
			},
			prefix:   "hello",
			expected: []string{"hello"},
		},
		{
			name: "single key - prefix match",
			setup: func(trie *T) {
				trie.Insert("hello", []interface{}{"world"})
			},
			prefix:   "hel",
			expected: []string{"hello"},
		},
		{
			name: "single key - no match",
			setup: func(trie *T) {
				trie.Insert("hello", []interface{}{"world"})
			},
			prefix:   "world",
			expected: nil,
		},
		{
			name: "multiple keys - empty prefix returns all",
			setup: func(trie *T) {
				trie.Insert("apple", []interface{}{"fruit1"})
				trie.Insert("banana", []interface{}{"fruit2"})
				trie.Insert("cherry", []interface{}{"fruit3"})
			},
			prefix:   "",
			expected: []string{"apple", "banana", "cherry"},
		},
		{
			name: "multiple keys with common prefix",
			setup: func(trie *T) {
				trie.Insert("cat", []interface{}{"animal1"})
				trie.Insert("car", []interface{}{"vehicle1"})
				trie.Insert("card", []interface{}{"object1"})
				trie.Insert("care", []interface{}{"emotion1"})
				trie.Insert("careful", []interface{}{"adjective1"})
			},
			prefix:   "car",
			expected: []string{"car", "card", "care", "careful"},
		},
		{
			name: "keys with shared prefix - partial match",
			setup: func(trie *T) {
				trie.Insert("test", []interface{}{"data1"})
				trie.Insert("testing", []interface{}{"data2"})
				trie.Insert("tester", []interface{}{"data3"})
				trie.Insert("temp", []interface{}{"data4"})
			},
			prefix:   "test",
			expected: []string{"test", "tester", "testing"},
		},
		{
			name: "prefix exists but has no data",
			setup: func(trie *T) {
				trie.Insert("testing", []interface{}{"data1"})
				trie.Insert("tester", []interface{}{"data2"})
				// "test" prefix exists in trie structure but has no data
			},
			prefix:   "test",
			expected: []string{"tester", "testing"},
		},
		{
			name: "lexical ordering verification",
			setup: func(trie *T) {
				trie.Insert("zebra", []interface{}{"animal"})
				trie.Insert("apple", []interface{}{"fruit"})
				trie.Insert("banana", []interface{}{"fruit"})
				trie.Insert("ant", []interface{}{"insect"})
				trie.Insert("zoo", []interface{}{"place"})
			},
			prefix:   "",
			expected: []string{"ant", "apple", "banana", "zebra", "zoo"},
		},
		{
			name: "unicode characters",
			setup: func(trie *T) {
				trie.Insert("café", []interface{}{"place1"})
				trie.Insert("naïve", []interface{}{"adjective1"})
				trie.Insert("résumé", []interface{}{"document1"})
			},
			prefix:   "",
			expected: []string{"café", "naïve", "résumé"},
		},
		{
			name: "unicode with prefix",
			setup: func(trie *T) {
				trie.Insert("café", []interface{}{"place1"})
				trie.Insert("cafeteria", []interface{}{"place2"})
				trie.Insert("car", []interface{}{"vehicle"})
			},
			prefix:   "caf",
			expected: []string{"cafeteria", "café"},
		},
		{
			name: "nested prefixes",
			setup: func(trie *T) {
				trie.Insert("a", []interface{}{"data1"})
				trie.Insert("ab", []interface{}{"data2"})
				trie.Insert("abc", []interface{}{"data3"})
				trie.Insert("abcd", []interface{}{"data4"})
			},
			prefix:   "ab",
			expected: []string{"ab", "abc", "abcd"},
		},
		{
			name: "single character keys",
			setup: func(trie *T) {
				trie.Insert("a", []interface{}{"data1"})
				trie.Insert("b", []interface{}{"data2"})
				trie.Insert("c", []interface{}{"data3"})
			},
			prefix:   "",
			expected: []string{"a", "b", "c"},
		},
		{
			name: "single character prefix",
			setup: func(trie *T) {
				trie.Insert("apple", []interface{}{"fruit"})
				trie.Insert("ant", []interface{}{"insect"})
				trie.Insert("ball", []interface{}{"toy"})
				trie.Insert("bat", []interface{}{"animal"})
			},
			prefix:   "a",
			expected: []string{"ant", "apple"},
		},
		{
			name: "long prefix no match",
			setup: func(trie *T) {
				trie.Insert("hello", []interface{}{"greeting"})
			},
			prefix:   "helloooooo",
			expected: nil,
		},
		{
			name: "numbers and special characters",
			setup: func(trie *T) {
				trie.Insert("test123", []interface{}{"data1"})
				trie.Insert("test-case", []interface{}{"data2"})
				trie.Insert("test_file", []interface{}{"data3"})
				trie.Insert("test.txt", []interface{}{"data4"})
			},
			prefix:   "test",
			expected: []string{"test-case", "test.txt", "test123", "test_file"},
		},
		{
			name: "case sensitivity",
			setup: func(trie *T) {
				trie.Insert("Apple", []interface{}{"fruit1"})
				trie.Insert("apple", []interface{}{"fruit2"})
				trie.Insert("APPLE", []interface{}{"fruit3"})
			},
			prefix:   "",
			expected: []string{"APPLE", "Apple", "apple"},
		},
		{
			name: "empty string key",
			setup: func(trie *T) {
				trie.Insert("", []interface{}{"empty"})
				trie.Insert("hello", []interface{}{"greeting"})
			},
			prefix:   "",
			expected: []string{"hello"}, // Empty string key might not be collected properly
		},
		{
			name: "complex branching structure",
			setup: func(trie *T) {
				trie.Insert("the", []interface{}{"article"})
				trie.Insert("their", []interface{}{"possessive"})
				trie.Insert("there", []interface{}{"location"})
				trie.Insert("these", []interface{}{"demonstrative"})
				trie.Insert("them", []interface{}{"pronoun"})
				trie.Insert("then", []interface{}{"temporal"})
			},
			prefix:   "the",
			expected: []string{"the", "their", "them", "then", "there", "these"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new trie instance
			trie := New()

			// Set up the trie with test data
			tt.setup(trie)

			// Execute the search
			result := trie.Search(tt.prefix)

			// Compare results
			if !equalStringSlices(result, tt.expected) {
				t.Errorf("Search(%q) = %v, expected %v", tt.prefix, result, tt.expected)
			}
		})
	}
}

// Additional edge case tests
func TestSearchEdgeCases(t *testing.T) {
	t.Run("concurrent access safety", func(t *testing.T) {
		trie := New()
		trie.Insert("test", []interface{}{"data"})

		// Test that multiple concurrent searches work correctly
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				result := trie.Search("test")
				if len(result) != 1 || result[0] != "test" {
					t.Errorf("Concurrent search failed: got %v", result)
				}
				done <- true
			}()
		}

		for i := 0; i < 10; i++ {
			<-done
		}
	})

	t.Run("very long key", func(t *testing.T) {
		trie := New()
		longKey := make([]rune, 1000)
		for i := range longKey {
			longKey[i] = 'a'
		}
		longKeyStr := string(longKey)

		trie.Insert(longKeyStr, []interface{}{"long data"})
		result := trie.Search(longKeyStr[:500]) // Search with first 500 chars

		expected := []string{longKeyStr}
		if !equalStringSlices(result, expected) {
			t.Errorf("Long key search failed: got %v, expected %v", result, expected)
		}
	})

	t.Run("many keys with same prefix", func(t *testing.T) {
		trie := New()
		prefix := "prefix"

		// Insert 100 keys with same prefix
		for i := 0; i < 100; i++ {
			key := fmt.Sprintf("%s%d", prefix, i)
			trie.Insert(key, []interface{}{fmt.Sprintf("data%d", i)})
		}

		result := trie.Search(prefix)
		if len(result) != 100 {
			t.Errorf("Expected 100 results, got %d", len(result))
		}

		// Verify lexical ordering
		for i := 1; i < len(result); i++ {
			if result[i] < result[i-1] {
				t.Errorf("Results not in lexical order: %s should come before %s", result[i-1], result[i])
			}
		}
	})
}

// Helper function to compare string slices
func equalStringSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
