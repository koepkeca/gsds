package stack

import (
	"fmt"
	"math/rand"
	"runtime"
	"testing"
	"time"
)

// Benchmark different data types
func BenchmarkPushPopString(b *testing.B) {
	s := New()
	defer s.Close()
	testString := "This is a test string for benchmarking purposes"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(testString)
		s.Pop()
	}
}

func BenchmarkPushPopLargeStruct(b *testing.B) {
	type LargeStruct struct {
		ID       int64
		Name     string
		Data     [1024]byte
		Metadata map[string]interface{}
	}

	s := New()
	defer s.Close()
	testStruct := LargeStruct{
		ID:   12345,
		Name: "BenchmarkStruct",
		Metadata: map[string]interface{}{
			"key1": "value1",
			"key2": 42,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(testStruct)
		s.Pop()
	}
}

func BenchmarkPushPopSlice(b *testing.B) {
	s := New()
	defer s.Close()
	testSlice := make([]int, 100)
	for i := range testSlice {
		testSlice[i] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(testSlice)
		s.Pop()
	}
}

// Memory allocation benchmarks
func BenchmarkMemoryAllocation(b *testing.B) {
	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	s := New()
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push(i)
	}

	runtime.ReadMemStats(&m2)
	b.ReportMetric(float64(m2.TotalAlloc-m1.TotalAlloc)/float64(b.N), "bytes/op")
}

// Random access pattern benchmarks
func BenchmarkRandomPushPop(b *testing.B) {
	s := New()
	defer s.Close()
	rand.Seed(time.Now().UnixNano())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Random decision: push or pop (with 60% push probability to maintain some stack depth)
		if rand.Float32() < 0.6 || s.Len() == 0 {
			s.Push(i)
		} else {
			s.Pop()
		}
	}
}

// Burst operation benchmarks
func BenchmarkBurstPush_1000(b *testing.B) {
	benchmarkBurstPush(b, 1000)
}

func BenchmarkBurstPush_10000(b *testing.B) {
	benchmarkBurstPush(b, 10000)
}

func benchmarkBurstPush(b *testing.B, burstSize int) {
	s := New()
	defer s.Close()
	// Calculate how many complete bursts we can do
	iterations := b.N / burstSize
	if iterations == 0 {
		iterations = 1
	}
	b.ResetTimer()
	for i := 0; i < iterations; i++ {
		b.StopTimer()
		for j := 0; j < burstSize; j++ {
			s.Push(j)
		}
		b.StartTimer()
		for j := 0; j < burstSize; j++ {
			s.Pop()
		}
	}
}

func BenchmarkBurstPop_1000(b *testing.B) {
	benchmarkBurstPop(b, 1000)
}

func BenchmarkBurstPop_10000(b *testing.B) {
	benchmarkBurstPop(b, 10000)
}

func benchmarkBurstPop(b *testing.B, burstSize int) {
	s := New()
	defer s.Close()

	// Calculate how many complete bursts we can do
	iterations := b.N / burstSize
	if iterations == 0 {
		iterations = 1
	}

	b.ResetTimer()
	for i := 0; i < iterations; i++ {
		b.StopTimer()
		// Pre-populate
		for j := 0; j < burstSize; j++ {
			s.Push(j)
		}
		b.StartTimer()

		// Burst pop
		for j := 0; j < burstSize; j++ {
			s.Pop()
		}
	}
}

// Growth pattern benchmarks - measure operations on stacks of different sizes
func BenchmarkStackGrowth_1K(b *testing.B) {
	benchmarkStackGrowth(b, 1000)
}

func BenchmarkStackGrowth_10K(b *testing.B) {
	benchmarkStackGrowth(b, 10000)
}

func BenchmarkStackGrowth_100K(b *testing.B) {
	benchmarkStackGrowth(b, 100000)
}

func BenchmarkStackGrowth_1M(b *testing.B) {
	benchmarkStackGrowth(b, 1000000)
}

func benchmarkStackGrowth(b *testing.B, size int) {
	s := New()
	defer s.Close()

	/*
		iterations := b.N / size
		if iterations == 0 {
			iterations = 1
		}
	*/

	// Pre-populate stack once
	for j := 0; j < size; j++ {
		s.Push(j)
	}

	b.ResetTimer()
	// Just measure push/pop operations on a stack that's already at the target size
	for i := 0; i < b.N; i++ {
		s.Push(i)
		s.Pop()
	}
}

// Concurrent access simulation (single-threaded but simulating workload)
func BenchmarkHighFrequency(b *testing.B) {
	s := New()
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate high-frequency push/pop pattern
		s.Push(i)
		s.Push(i + 1)
		s.Pop()
		s.Push(i + 2)
		s.Pop()
		s.Pop()
	}
}

// Deep stack operations
func BenchmarkDeepStackAccess(b *testing.B) {
	s := New()
	defer s.Close()

	// Create a deep stack (10,000 elements)
	const depth = 10000
	for i := 0; i < depth; i++ {
		s.Push(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Push("top")
		s.Pop()
	}
}

// Length operation benchmark
func BenchmarkLenOperation(b *testing.B) {
	s := New()
	defer s.Close()

	// Populate with various sizes and benchmark Len() calls
	sizes := []int{0, 10, 100, 1000, 10000}

	for _, size := range sizes {
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			// Populate to target size
			for len := s.Len(); len < int64(size); len = s.Len() {
				s.Push(len)
			}
			for len := s.Len(); len > int64(size); len = s.Len() {
				s.Pop()
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = s.Len()
			}
		})
	}
}

// Mixed workload benchmark
func BenchmarkMixedWorkload(b *testing.B) {
	s := New()
	defer s.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate realistic mixed workload
		cycle := i % 10
		switch cycle {
		case 0, 1, 2, 3, 4: // 50% push operations
			s.Push(i)
		case 5, 6, 7: // 30% pop operations
			if s.Len() > 0 {
				s.Pop()
			}
		case 8: // 10% length checks
			_ = s.Len()
		case 9: // 10% peek-like operation (push then immediate pop)
			s.Push("temp")
			s.Pop()
		}
	}
}

// Interface{} vs specific type benchmarks
func BenchmarkInterfaceOverhead(b *testing.B) {
	s := New()
	defer s.Close()

	b.Run("WithInterfaceAssertion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.Push(i)
			if val, ok := s.Pop().(int); ok {
				_ = val
			}
		}
	})

	b.Run("WithoutAssertion", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s.Push(i)
			_ = s.Pop()
		}
	})
}

// Stack reuse benchmark
func BenchmarkStackReuse(b *testing.B) {
	b.Run("NewStackEachTime", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			s := New()
			s.Push(i)
			s.Pop()
			s.Close()
		}
	})

	b.Run("ReuseStack", func(b *testing.B) {
		s := New()
		defer s.Close()
		for i := 0; i < b.N; i++ {
			s.Push(i)
			s.Pop()
		}
	})
}
