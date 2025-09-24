// Package queue implements a thread-safe queue.
//
// Benchmarking
// Sections:
//  1. Payload size benchmarks
//  2. Concurrency / producer-consumer benchmarks
//  3. Mixed read/write ratio benchmarks
//  4. Empty-queue edge-case benchmarks
//  5. Context-cancellation benchmark
//
// Run with:
//
//	go test -bench=. -benchmem ./...
//
// The -benchmem flag prints allocation statistics, which helps spot hidden copies
// of large payloads.
package queue

import (
	"context"
	"runtime"
	"sync"
	"testing"
)

/* -------------------------------------------------------------------------- */
// 1. Payload size benchmarks
/* -------------------------------------------------------------------------- */

type smallStruct struct {
	id   int64
	data float64
}

// 1 KB payload - large enough to expose copy-overhead but still fits in L1 cache.
var largePayload = make([]byte, 1024)

// Enqueue / Dequeue with a tiny primitive (int) - baseline.
func BenchmarkEnqueueInt(b *testing.B) {
	q := New()
	defer q.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
	}
}

func BenchmarkDequeueInt(b *testing.B) {
	q := New()
	defer q.Close()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Dequeue()
	}
}

// Enqueue / Dequeue with a small struct (approx 24 bytes).
func BenchmarkEnqueueSmallStruct(b *testing.B) {
	q := New()
	defer q.Close()
	s := smallStruct{id: 42, data: 3.1415}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(s)
	}
}

func BenchmarkDequeueSmallStruct(b *testing.B) {
	q := New()
	defer q.Close()
	s := smallStruct{id: 42, data: 3.1415}
	for i := 0; i < b.N; i++ {
		q.Enqueue(s)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Dequeue()
	}
}

// Enqueue / Dequeue with a 1 KB byte slice - stresses copying of larger data.
func BenchmarkEnqueueLargeSlice(b *testing.B) {
	q := New()
	defer q.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q.Enqueue(largePayload)
	}
}

func BenchmarkDequeueLargeSlice(b *testing.B) {
	q := New()
	defer q.Close()
	for i := 0; i < b.N; i++ {
		q.Enqueue(largePayload)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Dequeue()
	}
}

/* -------------------------------------------------------------------------- */
// 2. Concurrency / producer-consumer benchmarks
/* -------------------------------------------------------------------------- */

// Helper that runs N producers and M consumers concurrently.
// Each goroutine performs exactly b.N operations.
func runProdCons(b *testing.B, producers, consumers int) {
	q := New()
	defer q.Close()

	var wg sync.WaitGroup
	wg.Add(producers + consumers)

	// Consumers - keep pulling until they have performed b.N dequeues each.
	for i := 0; i < consumers; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				q.Dequeue()
			}
		}()
	}

	// Producers - push b.N items each.
	for i := 0; i < producers; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < b.N; j++ {
				q.Enqueue(id<<32 + j) // unique-ish value, still an int
			}
		}(i)
	}

	b.ResetTimer()
	wg.Wait()
}

// 2 producers / 2 consumers
func BenchmarkProdCons_2P_2C(b *testing.B) { runProdCons(b, 2, 2) }

// 4 producers / 4 consumers
func BenchmarkProdCons_4P_4C(b *testing.B) { runProdCons(b, 4, 4) }

// 8 producers / 2 consumers - write-heavy scenario
func BenchmarkProdCons_8P_2C(b *testing.B) { runProdCons(b, 8, 2) }

/* -------------------------------------------------------------------------- */
// 3. Mixed read/write ratio benchmarks
/* -------------------------------------------------------------------------- */

// Helper that performs a deterministic mix of Enqueue/Dequeue calls.
// The writePct argument is the percentage of operations that are writes.
func benchMixedRatio(b *testing.B, writePct int) {
	q := New()
	defer q.Close()

	// Pre-fill a modest amount so Dequeue rarely hits the empty branch.
	const preload = 1000
	for i := 0; i < preload; i++ {
		q.Enqueue(i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if i%100 < writePct { // simple modulo-based probability
			q.Enqueue(i)
		} else {
			q.Dequeue()
		}
	}
}

// 90% reads, 10% writes (read-heavy)
func BenchmarkMixedReadHeavy(b *testing.B) { benchMixedRatio(b, 10) }

// 50% reads, 50% writes (balanced)
func BenchmarkMixedBalanced(b *testing.B) { benchMixedRatio(b, 50) }

// 90% writes, 10% reads (write-heavy)
func BenchmarkMixedWriteHeavy(b *testing.B) { benchMixedRatio(b, 90) }

/* -------------------------------------------------------------------------- */
// 4. Empty-queue edge-case benchmarks
/* -------------------------------------------------------------------------- */

// Repeatedly dequeue from an empty queue - ensures the fast-path stays cheap.
func BenchmarkDequeueEmpty(b *testing.B) {
	q := New()
	defer q.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Dequeue()
	}
}

// Call Front on an empty queue repeatedly.
func BenchmarkFrontEmpty(b *testing.B) {
	q := New()
	defer q.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Front()
	}
}

// Call Back on an empty queue repeatedly.
func BenchmarkBackEmpty(b *testing.B) {
	q := New()
	defer q.Close()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Back()
	}
}

/* -------------------------------------------------------------------------- */
// 5. Context-cancellation benchmark
/* -------------------------------------------------------------------------- */

// BenchmarkCancelDuringOps measures how quickly the queue shuts down when its
// context is cancelled while a burst of work is in flight. The key to avoiding
// a data race is to let the workers stop before the queue's internal
// goroutine closes its channel.
//
// Two separate contexts are used:
//   - queueCtx - passed to NewWithContext; this is the one that will be closed
//     after the workers have exited.
//   - workerCtx - watched by the workers; it is cancelled first, causing the
//     workers to return without attempting another Enqueue.
func BenchmarkCancelDuringOps(b *testing.B) {
	// 1. Context for the queue (will be cancelled after workers finish).
	queueCtx, cancelQueue := context.WithCancel(context.Background())
	q := NewWithContext(queueCtx)
	// Ensure the queue's internal goroutine stops at the end of the benchmark.
	defer q.Close()

	// 2. Separate context for the workers.
	workerCtx, cancelWorker := context.WithCancel(context.Background())

	// 3. Spin up a few workers that keep enqueuing while workerCtx is alive.
	var wg sync.WaitGroup
	const workers = 4
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		go func(id int) {
			defer wg.Done()
			for {
				select {
				case <-workerCtx.Done():
					// Worker context cancelled - exit before we might hit a closed channel.
					return
				default:
					// The actual value is irrelevant; we just need a write.
					q.Enqueue(id<<32 + b.N) // use b.N to keep the compiler from optimizing away the call
				}
			}
		}(i)
	}

	// 4. Let the workers run briefly, then cancel their context.
	// A single scheduler yield is enough for the benchmark; no arbitrary sleeps.
	runtime.Gosched()
	cancelWorker() // workers stop here, before the queue's channel is closed

	// 5. Wait for all workers to notice the cancellation and exit.
	wg.Wait()

	// 6. Now cancel the queue's own context - this will close the internal channel
	//    safely because no worker is trying to send on it any more.
	cancelQueue()
}

/* -------------------------------------------------------------------------- */
// End of benchmark suite
/* -------------------------------------------------------------------------- */
