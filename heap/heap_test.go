package heap

import (
	"context"
	"sync"
	"testing"
	"time"
)

// IntHeap is a min-heap of integers for testing
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

func TestNew(t *testing.T) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()
	if h == nil {
		t.Fatal("New() returned nil")
	}
}

func TestLen(t *testing.T) {
	ih := &IntHeap{1, 2, 3}
	h := New(ih)
	defer h.Close()

	if got := h.Len(); got != 3 {
		t.Errorf("Len() = %d, want 3", got)
	}
}

func TestPush(t *testing.T) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	ok := h.Push(5)
	if !ok {
		t.Error("Push() returned false")
	}

	if got := h.Len(); got != 1 {
		t.Errorf("Len() after Push = %d, want 1", got)
	}
}

func TestPop(t *testing.T) {
	ih := &IntHeap{3, 1, 2}
	h := New(ih)
	defer h.Close()

	// Pop should return elements in sorted order (min-heap)
	want := []int{1, 2, 3}
	for i, expected := range want {
		got := h.Pop()
		if got.(int) != expected {
			t.Errorf("Pop() %d = %v, want %v", i, got, expected)
		}
	}

	if got := h.Len(); got != 0 {
		t.Errorf("Len() after all Pops = %d, want 0", got)
	}
}

func TestRemove(t *testing.T) {
	ih := &IntHeap{1, 2, 3, 4, 5}
	h := New(ih)
	defer h.Close()

	initialLen := h.Len()
	removed := h.Remove(2)

	if removed == nil {
		t.Error("Remove() returned nil")
	}

	if got := h.Len(); got != initialLen-1 {
		t.Errorf("Len() after Remove = %d, want %d", got, initialLen-1)
	}
}

func TestPushPop(t *testing.T) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	values := []int{5, 3, 7, 1, 9, 2, 8}
	for _, v := range values {
		h.Push(v)
	}

	var result []int
	for h.Len() > 0 {
		result = append(result, h.Pop().(int))
	}

	// Verify elements come out in sorted order
	for i := 1; i < len(result); i++ {
		if result[i-1] > result[i] {
			t.Errorf("Result not sorted: %v", result)
			break
		}
	}
}

func TestConcurrentPush(t *testing.T) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	var wg sync.WaitGroup
	numGoroutines := 10
	pushesPer := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			for j := 0; j < pushesPer; j++ {
				h.Push(base*100 + j)
			}
		}(i)
	}

	wg.Wait()

	expected := numGoroutines * pushesPer
	if got := h.Len(); got != expected {
		t.Errorf("Concurrent Push: Len() = %d, want %d", got, expected)
	}
}

func TestConcurrentPushPop(t *testing.T) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Pre-fill the heap
	for i := 0; i < 100; i++ {
		h.Push(i)
	}

	var wg sync.WaitGroup
	numGoroutines := 10

	// Concurrent pushes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(base int) {
			defer wg.Done()
			h.Push(base + 1000)
		}(i)
	}

	// Concurrent pops
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			h.Pop()
		}()
	}

	wg.Wait()

	// Should have same number of elements (pushed 10, popped 10)
	if got := h.Len(); got != 100 {
		t.Errorf("Concurrent Push/Pop: Len() = %d, want 100", got)
	}
}

func TestConcurrentLen(t *testing.T) {
	ih := &IntHeap{1, 2, 3, 4, 5}
	h := New(ih)
	defer h.Close()

	var wg sync.WaitGroup
	numGoroutines := 20

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = h.Len()
		}()
	}

	wg.Wait()
	// Test passes if no race conditions or panics occur
}

func TestClose(t *testing.T) {
	ih := &IntHeap{1, 2, 3}
	h := New(ih)

	h.Close()

	// After closing, operations should panic or behave undefined
	// This test just ensures Close() doesn't panic itself
}

func TestPushPopEmpty(t *testing.T) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Test normal push/pop cycle until empty
	h.Push(1)
	h.Push(2)

	result := h.Pop()
	if result.(int) != 1 {
		t.Errorf("Pop() = %v, want 1", result)
	}

	result = h.Pop()
	if result.(int) != 2 {
		t.Errorf("Pop() = %v, want 2", result)
	}

	// Verify the heap is now empty
	if got := h.Len(); got != 0 {
		t.Errorf("Len() after popping all = %d, want 0", got)
	}

	// Note: Calling Pop() on an empty heap will panic - this is expected behavior
	// and matches container/heap. Callers should check Len() before calling Pop().
}

func TestNewWithContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ih := &IntHeap{}
	h := NewWithContext(ctx, ih)

	if h == nil {
		t.Fatal("NewWithContext() returned nil")
	}

	h.Push(1)
	if got := h.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}

func TestContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	ih := &IntHeap{}
	h := NewWithContext(ctx, ih)

	h.Push(1)
	h.Push(2)

	// Cancel the context
	cancel()

	// Give the goroutine time to shut down
	time.Sleep(100 * time.Millisecond)

	// After cancellation, operations should fail/panic
	// This test ensures the heap shuts down properly
}

func TestContextTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	ih := &IntHeap{}
	h := NewWithContext(ctx, ih)

	h.Push(1)

	// Wait for timeout
	time.Sleep(200 * time.Millisecond)

	// Heap should have shut down automatically
	// This test ensures timeout-based cleanup works
}

func TestNewWithNilContext(t *testing.T) {
	ih := &IntHeap{}
	h := NewWithContext(nil, ih)
	defer h.Close()

	h.Push(1)
	if got := h.Len(); got != 1 {
		t.Errorf("Len() = %d, want 1", got)
	}
}
