package heap

import (
	"testing"
)

func BenchmarkPush(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
	}
}

func BenchmarkPop(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Pre-fill the heap
	for i := 0; i < b.N; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Pop()
	}
}

func BenchmarkPushPop(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		if i%2 == 0 {
			h.Pop()
		}
	}
}

func BenchmarkLen(b *testing.B) {
	ih := &IntHeap{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	h := New(ih)
	defer h.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = h.Len()
	}
}

func BenchmarkRemove(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Pre-fill with enough elements
	for i := 0; i < b.N*2; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if h.Len() > 0 {
			h.Remove(0)
		}
	}
}

func BenchmarkConcurrentPush(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			h.Push(i)
			i++
		}
	})
}

func BenchmarkConcurrentPop(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Pre-fill the heap
	for i := 0; i < b.N*2; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if h.Len() > 0 {
				h.Pop()
			}
		}
	})
}

func BenchmarkConcurrentMixed(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Pre-fill the heap
	for i := 0; i < 1000; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			if i%3 == 0 {
				h.Push(i)
			} else if i%3 == 1 && h.Len() > 0 {
				h.Pop()
			} else {
				_ = h.Len()
			}
			i++
		}
	})
}

func BenchmarkSequentialOperations(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		_ = h.Len()
		if i%10 == 0 && h.Len() > 0 {
			h.Pop()
		}
	}
}

// Benchmark with different heap sizes
func BenchmarkPushSmallHeap(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Keep heap size around 10
	for i := 0; i < 10; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		if h.Len() > 10 {
			h.Pop()
		}
	}
}

func BenchmarkPushLargeHeap(b *testing.B) {
	ih := &IntHeap{}
	h := New(ih)
	defer h.Close()

	// Keep heap size around 10000
	for i := 0; i < 10000; i++ {
		h.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h.Push(i)
		if h.Len() > 10000 {
			h.Pop()
		}
	}
}
