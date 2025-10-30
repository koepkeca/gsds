package heap

import (
	"container/heap"
	"context"
)

// H is a thread-safe wrapper around heap.Interface that serializes all heap
// operations through a single goroutine. This ensures safe concurrent access
// to the underlying heap without requiring external synchronization.
type H struct {
	op     chan (func(heap.Interface))
	cancel context.CancelFunc
}

// Len returns the number of elements in the heap.
// This operation is thread-safe and blocks until the result is available.
func (h *H) Len() int {
	lch := make(chan int)
	h.op <- func(curr heap.Interface) {
		lch <- curr.Len()
		return
	}
	return <-lch
}

// Push adds an element to the heap and maintains the heap invariant.
// This operation is thread-safe and blocks until the push is complete.
// Returns true if the operation completed successfully.
func (h *H) Push(x interface{}) bool {
	bch := make(chan bool)
	h.op <- func(curr heap.Interface) {
		heap.Push(curr, x)
		bch <- true
		return
	}
	return <-bch
}

// Pop removes and returns the minimum element (according to Less) from the heap.
// This operation is thread-safe and blocks until the result is available.
// The complexity is O(log n) where n = h.Len().
func (h *H) Pop() interface{} {
	ich := make(chan interface{})
	h.op <- func(curr heap.Interface) {
		ich <- heap.Pop(curr)
		return
	}
	return <-ich
}

// Remove removes and returns the element at index idx from the heap.
// This operation is thread-safe and blocks until the result is available.
// The complexity is O(log n) where n = h.Len().
func (h *H) Remove(idx int) interface{} {
	ich := make(chan interface{})
	h.op <- func(curr heap.Interface) {
		ich <- heap.Remove(curr, idx)
		return
	}
	return <-ich
}

// Close shuts down the heap's operation channel, preventing further operations.
// After calling Close, any subsequent operations will panic.
// This should be called when the heap is no longer needed to prevent goroutine leaks.
func (h *H) Close() {
	h.cancel()
	return
}

// New creates a new thread-safe heap wrapper around the provided heap.Interface.
// The heap is initialized immediately and operations are processed in the order
// they are received. The caller must eventually call Close() to prevent goroutine leaks.
func New(i heap.Interface) *H {
	return NewWithContext(nil, i)
}

// NewWithContext creates a new heap which will cancel upon context cancellation.
func NewWithContext(ctx context.Context, i heap.Interface) *H {
	h := &H{op: make(chan func(heap.Interface))}
	ready := make(chan struct{})
	go h.loop(ctx, i, ready)
	<-ready
	return h
}

// loop is the main event loop that processes heap operations sequentially.
// It runs in its own goroutine and ensures thread-safe access to the heap.
// The loop terminates when the op channel is closed or context is cancelled.
func (h *H) loop(c context.Context, i heap.Interface, ready chan struct{}) {
	heap.Init(i)
	ctx, cancel := context.WithCancel(context.Background())
	if c != nil {
		ctx, cancel = context.WithCancel(c)
	}
	h.cancel = cancel
	close(ready)
	for {
		select {
		case op := <-h.op:
			op(i)
		case <-ctx.Done():
			close(h.op)
			cancel()
			return
		}
	}
}
