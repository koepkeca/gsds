package trie

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestBasicTrie(t *testing.T) {
	trie := New()
	defer trie.Close()
	tt := []string{"example", "1", "2", "3"}
	td := make([]interface{}, len(tt))
	for k, v := range tt {
		td[k] = v
	}
	trie.Insert("test", td)
	return
}

func TestBasicRead(t *testing.T) {
	trie := New()
	defer trie.Close()
	tt := []int64{1, 2, 3, 4, 5}
	td := make([]interface{}, len(tt))
	for i, v := range tt {
		td[i] = v
	}
	e := trie.Insert("testing", td)
	if e != nil {
		t.Fatal(e)
	}
	_, e = trie.Get("testing")
	if e != nil {
		t.Fatal(e)
	}
	return
}

func TestReadEmptyTrie(t *testing.T) {
	trie := New()
	defer trie.Close()
	v, e := trie.Get("example")
	if e != nil {
		t.Fatal(e)
	}
	if v != nil {
		t.Fatal("Got invalid read of empty trie")
	}
	return
}

func TestSearchAux(t *testing.T) {
	trie := New()
	defer trie.Close()
	e := trie.Insert("example", []interface{}{"example"})
	if e != nil {
		panic(e)
	}
	e = trie.Insert("authority", []interface{}{"authority"})
	if e != nil {
		panic(e)
	}
	e = trie.Insert("exam", []interface{}{"exam"})
	if e != nil {
		panic(e)
	}
	tmp := trie.Search("exa")
	if len(tmp) != 2 {
		t.Fatalf("Depth search failed, expected 2 results got %d", len(tmp))
	}
	return
}

func TestAscii(t *testing.T) {
	pullList := []string{"abc", "zxy", "def", "xxx", "abacab", "lmnop", "ponml"}
	ordList := []string{"abacab", "abc", "def", "lmnop", "ponml", "xxx", "zxy"}
	trie := New()
	defer trie.Close()
	for _, next := range pullList {
		e := trie.Insert(next, []interface{}{next})
		if e != nil {
			t.Fatal(e)
		}
	}
	tmp := trie.Search("")
	for idx, next := range ordList {
		if next != tmp[idx] {
			t.Fatalf("Mismatch expected %s got %s", next, tmp[idx])
		}
	}
	return
}

func TestUTF(t *testing.T) {
	pullList := []string{"こんにちは", "こんばんは", "今日"}
	trie := New()
	defer trie.Close()
	e := trie.Insert("こんにちは", []interface{}{"こんにちは"})
	if e != nil {
		t.Fatal(e)
	}
	e = trie.Insert("今日", []interface{}{"今日"})
	if e != nil {
		t.Fatal(e)
	}
	e = trie.Insert("こんばんは", []interface{}{"こんばんは"})
	if e != nil {
		t.Fatal(e)
	}
	tmp := trie.Search("")
	for i, rlt := range tmp {
		if pullList[i] != rlt {
			t.Fatalf("Mismatch expected %s got %s", pullList[i], rlt)
		}
	}
	return
}

func TestEmptyInvalidResult(t *testing.T) {
	trie := New()
	tmp := trie.Search("")
	if len(tmp) != 0 {
		t.Fatal("Empty trie got search result??")
	}
	trie.Close()
	trie = New()
	defer trie.Close()
	trie.Insert("aaa", []interface{}{"aaa"})
	tmp = trie.Search("bbbbb")
	if len(tmp) != 0 {
		t.Fatalf("Trie search with no result obtained a result??")
	}
	return
}

func TestReplace(t *testing.T) {
	tr := New()
	defer tr.Close()

	// Test replacing non-existent key (should create new)
	tr.Replace = true
	err := tr.Insert("hello", []interface{}{"world"})
	if err != nil {
		t.Fatalf("Replace should not error when creating new key: %v", err)
	}

	// Verify the key was created
	data, err := tr.Get("hello")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(data) != 1 || data[0] != "world" {
		t.Fatalf("Expected [world], got %v", data)
	}

	// Test replacing existing key
	err = tr.Insert("hello", []interface{}{"replaced", "data"})
	if err != nil {
		t.Fatalf("Replace should not error when replacing existing key: %v", err)
	}

	// Verify the key was replaced (not appended)
	data, err = tr.Get("hello")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(data) != 2 || data[0] != "replaced" || data[1] != "data" {
		t.Fatalf("Expected [replaced data], got %v", data)
	}

	// Test replace with empty key
	err = tr.Insert("", []interface{}{"test"})
	if err == nil {
		t.Fatal("Replace should error with empty key")
	}

	// Test replace with empty data
	err = tr.Insert("test", []interface{}{})
	if err == nil {
		t.Fatal("Replace should error with empty data")
	}
}

func TestDelete(t *testing.T) {
	tr := New()
	defer tr.Close()

	// Insert some test data
	tr.Insert("hello", []interface{}{"world"})
	tr.Insert("help", []interface{}{"me"})
	tr.Insert("helmet", []interface{}{"safety"})
	tr.Insert("he", []interface{}{"pronoun"})

	// Test deleting existing key
	deleted, err := tr.Delete("hello")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !deleted {
		t.Fatal("Delete should return true for existing key")
	}

	// Verify key is deleted
	data, err := tr.Get("hello")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if data != nil {
		t.Fatalf("Expected nil after delete, got %v", data)
	}

	// Test deleting non-existent key
	deleted, err = tr.Delete("nonexistent")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if deleted {
		t.Fatal("Delete should return false for non-existent key")
	}

	// Test that other keys still exist
	data, err = tr.Get("help")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(data) != 1 || data[0] != "me" {
		t.Fatalf("Expected [me], got %v", data)
	}

	// Test deleting with empty key
	deleted, err = tr.Delete("")
	if err == nil {
		t.Fatal("Delete should error with empty key")
	}
	if deleted {
		t.Fatal("Delete should return false with empty key")
	}

	// Test cleanup of unnecessary nodes
	tr.Insert("test", []interface{}{"data"})
	tr.Insert("testing", []interface{}{"more"})

	// Delete "testing" - should clean up nodes that are no longer needed
	deleted, err = tr.Delete("testing")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}
	if !deleted {
		t.Fatal("Delete should return true for existing key")
	}

	// "test" should still exist
	data, err = tr.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(data) != 1 || data[0] != "data" {
		t.Fatalf("Expected [data], got %v", data)
	}
}

func TestExists(t *testing.T) {
	tr := New()
	defer tr.Close()

	// Test non-existent key
	exists, err := tr.Exists("nonexistent")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("Exists should return false for non-existent key")
	}

	// Insert a key
	tr.Insert("hello", []interface{}{"world"})

	// Test existing key
	exists, err = tr.Exists("hello")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if !exists {
		t.Fatal("Exists should return true for existing key")
	}

	// Test prefix that exists but has no data
	exists, err = tr.Exists("hel")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("Exists should return false for prefix without data")
	}

	// Test with empty key
	exists, err = tr.Exists("")
	if err == nil {
		t.Fatal("Exists should error with empty key")
	}
	if exists {
		t.Fatal("Exists should return false with empty key")
	}

	// Delete the key and test again
	tr.Delete("hello")
	exists, err = tr.Exists("hello")
	if err != nil {
		t.Fatalf("Exists failed: %v", err)
	}
	if exists {
		t.Fatal("Exists should return false for deleted key")
	}
}

func TestReplaceVsInsert(t *testing.T) {
	tr := New()
	defer tr.Close()

	// Insert initial data
	tr.Insert("key", []interface{}{"value1"})

	// Insert more data (should append)
	tr.Insert("key", []interface{}{"value2"})

	data, _ := tr.Get("key")
	if len(data) != 2 || data[0] != "value1" || data[1] != "value2" {
		t.Fatalf("Insert should append. Expected [value1 value2], got %v", data)
	}
	tr.Replace = true
	// Replace data (should replace, not append)
	tr.Insert("key", []interface{}{"replaced"})

	data, _ = tr.Get("key")
	if len(data) != 1 || data[0] != "replaced" {
		t.Fatalf("Replace should replace. Expected [replaced], got %v", data)
	}
}

func TestDeleteCleanup(t *testing.T) {
	tr := New()
	defer tr.Close()

	// Create a chain: a -> ab -> abc
	tr.Insert("a", []interface{}{"data_a"})
	tr.Insert("ab", []interface{}{"data_ab"})
	tr.Insert("abc", []interface{}{"data_abc"})

	// Delete "abc" - should only remove the leaf
	deleted, _ := tr.Delete("abc")
	if !deleted {
		t.Fatal("Should have deleted abc")
	}

	// "ab" should still exist
	exists, _ := tr.Exists("ab")
	if !exists {
		t.Fatal("ab should still exist after deleting abc")
	}

	// Delete "ab" - should remove the node but keep the path to "a"
	deleted, _ = tr.Delete("ab")
	if !deleted {
		t.Fatal("Should have deleted ab")
	}

	// "a" should still exist
	exists, _ = tr.Exists("a")
	if !exists {
		t.Fatal("a should still exist after deleting ab")
	}

	// Delete "a" - should clean up all nodes
	deleted, _ = tr.Delete("a")
	if !deleted {
		t.Fatal("Should have deleted a")
	}

	// Search should return empty
	results := tr.Search("a")
	if results != nil {
		t.Fatalf("Search should return nil after all deletions, got %v", results)
	}
}

func TestConcurrentOperations(t *testing.T) {
	tr1 := New()
	tr2 := New()
	tr3 := New()

	done := make(chan bool, 3)

	// Concurrent inserts in append mode
	go func() {
		for i := 0; i < 100; i++ {
			tr1.Insert("concurrent1", []interface{}{i})
		}
		done <- true
	}()

	// Concurrent inserts in replace mode
	go func() {
		tr2.Replace = true
		for i := 0; i < 100; i++ {
			tr2.Insert("concurrent2", []interface{}{i})
		}
		done <- true
	}()

	// Concurrent deletes and re-inserts
	go func() {
		for i := 0; i < 100; i++ {
			tr3.Insert("concurrent3", []interface{}{i})
			tr3.Delete("concurrent3")
		}
		done <- true
	}()

	// Wait for all goroutines
	for i := 0; i < 3; i++ {
		<-done
	}

	// Check that operations completed without crashing
	data1, _ := tr1.Get("concurrent1")
	if len(data1) != 100 {
		t.Fatalf("Expected 100 items for concurrent1, got %d", len(data1))
	}

	data2, _ := tr2.Get("concurrent2")
	if len(data2) != 1 {
		t.Fatalf("Expected 1 item for concurrent2 (replaced), got %d", len(data2))
	}

	exists, _ := tr3.Exists("concurrent3")
	if exists {
		t.Fatal("concurrent3 should not exist after delete")
	}

	// Close all tries after testing
	tr1.Close()
	tr2.Close()
	tr3.Close()
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	tr := NewWithContext(ctx)

	// Insert some data
	tr.Insert("test", []interface{}{"data"})

	// Verify data exists
	data, err := tr.Get("test")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(data) != 1 || data[0] != "data" {
		t.Fatalf("Expected [data], got %v", data)
	}

	// Cancel context
	cancel()

	// Give some time for the goroutine to clean up
	time.Sleep(100 * time.Millisecond)

	// Operations should still work briefly due to buffering, but eventually fail
	// This test mainly ensures no panic occurs during shutdown
}

func TestUnicodeSupport(t *testing.T) {
	tr := New()
	defer tr.Close()

	// Test with Unicode characters
	unicodeKey := "测试🌟"
	tr.Insert(unicodeKey, []interface{}{"unicode_data"})

	data, err := tr.Get(unicodeKey)
	if err != nil {
		t.Fatalf("Get failed for unicode key: %v", err)
	}
	if len(data) != 1 || data[0] != "unicode_data" {
		t.Fatalf("Expected [unicode_data], got %v", data)
	}

	// Test delete with Unicode
	deleted, err := tr.Delete(unicodeKey)
	if err != nil {
		t.Fatalf("Delete failed for unicode key: %v", err)
	}
	if !deleted {
		t.Fatal("Should have deleted unicode key")
	}

	exists, _ := tr.Exists(unicodeKey)
	if exists {
		t.Fatal("Unicode key should not exist after deletion")
	}
}

// TestSearchRaceCondition tests the documented TOCTOU behavior
// where keys returned by Search might be deleted before Get() is called
func TestSearchRaceCondition(t *testing.T) {
	trie := New()

	// Set up initial data
	keys := []string{"test1", "test2", "test3", "test4", "test5"}
	for i, key := range keys {
		trie.Insert(key, []interface{}{i})
	}

	var wg sync.WaitGroup
	results := make(chan testResult, 10)

	// Goroutine 1: Search and then try to Get the results
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Search for all keys with "test" prefix
		foundKeys := trie.Search("test")

		// Small delay to increase chance of race condition
		time.Sleep(10 * time.Millisecond)

		// Try to Get each key that was found
		for _, key := range foundKeys {
			value, err := trie.Get(key)
			results <- testResult{
				key:    key,
				found:  err == nil && value != nil,
				action: "get_after_search",
			}
		}
	}()

	// Goroutine 2: Delete some keys while search/get is happening
	wg.Add(1)
	go func() {
		defer wg.Done()

		// Small delay to let search happen first
		time.Sleep(5 * time.Millisecond)

		// Delete a couple of keys
		keysToDelete := []string{"test2", "test4"}
		for _, key := range keysToDelete {
			trie.Delete(key)
			results <- testResult{
				key:    key,
				found:  false,
				action: "delete",
			}
		}
	}()

	// Wait for both goroutines to complete
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results
	deletedKeys := make(map[string]bool)
	getResults := make(map[string]bool)

	for result := range results {
		if result.action == "delete" {
			deletedKeys[result.key] = true
		} else if result.action == "get_after_search" {
			getResults[result.key] = result.found
		}
	}

	// Verify the race condition behavior
	raceConditionOccurred := false
	for deletedKey := range deletedKeys {
		if found, exists := getResults[deletedKey]; exists && !found {
			raceConditionOccurred = true
			t.Logf("TOCTOU race condition confirmed: key '%s' was found in search but missing in get", deletedKey)
		}
	}

	// This test documents the behavior - we expect race conditions can occur
	if !raceConditionOccurred {
		t.Logf("No race condition occurred in this run (timing dependent)")
	}

	// Verify that non-deleted keys can still be retrieved
	nonDeletedFound := 0
	for key, found := range getResults {
		if !deletedKeys[key] && found {
			nonDeletedFound++
		}
	}

	if nonDeletedFound == 0 {
		t.Error("Expected at least some non-deleted keys to be retrievable")
	}
}

// TestSearchConsistentBehavior tests that the race condition is properly documented
// and that callers should handle missing keys gracefully
func TestSearchConsistentBehavior(t *testing.T) {
	trie := New()

	// Insert test data
	trie.Insert("prefix1", []interface{}{"data1"})
	trie.Insert("prefix2", []interface{}{"data2"})
	trie.Insert("prefix3", []interface{}{"data3"})

	// Get the keys
	keys := trie.Search("prefix")
	expectedCount := 3

	if len(keys) != expectedCount {
		t.Errorf("Expected %d keys, got %d", expectedCount, len(keys))
	}

	// Delete one key
	trie.Delete("prefix2")

	// Try to get all previously found keys - this demonstrates the TOCTOU issue
	successfulGets := 0
	for _, key := range keys {
		if value, err := trie.Get(key); err == nil && value != nil {
			successfulGets++
		}
	}

	// We should have one less successful get than keys found
	expectedSuccessful := expectedCount - 1
	if successfulGets != expectedSuccessful {
		t.Errorf("Expected %d successful gets, got %d", expectedSuccessful, successfulGets)
	}

	t.Logf("Demonstrated TOCTOU: Search found %d keys, but only %d were retrievable after deletion",
		len(keys), successfulGets)
}

// TestConcurrentSearchAndModify runs multiple concurrent operations to stress test
func TestConcurrentSearchAndModify(t *testing.T) {
	trie := New()

	// Initial data
	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%03d", i)
		trie.Insert(key, []interface{}{i})
	}

	var wg sync.WaitGroup
	numSearchers := 5
	numModifiers := 3

	// Start multiple searchers
	for i := 0; i < numSearchers; i++ {
		wg.Add(1)
		go func(searcherID int) {
			defer wg.Done()

			for j := 0; j < 10; j++ {
				keys := trie.Search("key")

				// Try to get some of the found keys
				for k, key := range keys {
					if k >= 10 { // Only try first 10 to not overwhelm
						break
					}
					_, _ = trie.Get(key) // Result might be nil/error due to race condition
				}
			}
		}(i)
	}

	// Start multiple modifiers
	for i := 0; i < numModifiers; i++ {
		wg.Add(1)
		go func(modifierID int) {
			defer wg.Done()

			for j := 0; j < 20; j++ {
				// Delete and re-insert keys
				keyToModify := fmt.Sprintf("key%03d", (modifierID*20+j)%100)
				trie.Delete(keyToModify)

				// Small delay
				time.Sleep(1 * time.Millisecond)

				// Re-insert
				trie.Insert(keyToModify, []interface{}{modifierID*1000 + j})
			}
		}(i)
	}

	wg.Wait()

	// Final verification - trie should still be functional
	finalKeys := trie.Search("key")
	if len(finalKeys) == 0 {
		t.Error("Expected some keys to remain after concurrent operations")
	}

	t.Logf("Concurrent test completed: %d keys remain", len(finalKeys))
}

type testResult struct {
	key    string
	found  bool
	action string
}
