package queue

import (
	"testing"
)

func TestQueueCreation(t *testing.T) {
	q := New()
	defer q.Close()
	if q.Len() != 0 {
		t.Errorf("Failed, invalid queue length.")
	}
	return
}

func TestQueueLength(t *testing.T) {
	q := New()
	defer q.Close()
	q.Enqueue(14)
	q.Enqueue(42)
	q.Enqueue("testing")
	q.Enqueue([]byte("Viper"))
	len := q.Len()
	if len != 4 {
		t.Errorf("Failed, invalid stack length, got %d expected 4", len)
	}
	return
}

func TestQueueOrder(t *testing.T) {
	q := New()
	defer q.Close()
	q.Enqueue(16)
	q.Enqueue(32)
	q.Enqueue(64)
	nv, ok := q.Dequeue().(int)
	if !ok {
		t.Errorf("Failed, Dequeue got wrong type")
		return
	}
	if nv != 16 {
		t.Errorf("Failed, got incorrect value order")
		return
	}
	return
}

func TestSizeAfterDequeue(t *testing.T) {
	q := New()
	defer q.Close()
	q.Enqueue(16)
	q.Enqueue("test")
	q.Enqueue("私は笑い男だ")
	_ = q.Dequeue()
	_ = q.Dequeue()
	_ = q.Dequeue()
	if q.Len() != 0 {
		t.Errorf("Failed, poped through entire stack, yet size is non-zero")
	}
	return
}

func TestEmptyDequeue(t *testing.T) {
	q := New()
	defer q.Close()
	v := q.Dequeue()
	if v != nil {
		t.Errorf("Empty Pop got non-nil value")
	}
}

func TestEmptyDequeueWithValues(t *testing.T) {
	q := New()
	defer q.Close()
	q.Enqueue("Thingy")
	_ = q.Dequeue()
	v := q.Dequeue()
	if v != nil {
		t.Errorf("Empty stack with values got non-nil value")
	}
}

func BenchmarkEqualRWWithInt(b *testing.B) {
	q := New()
	defer q.Close()
	write := false
	for i := 0; i < b.N; i++ {
		if q.Dequeue() == nil || write {
			q.Enqueue(i)
		} else {
			q.Dequeue()
			write = true
		}
	}
}

func BenchmarkROnlyWithInt(b *testing.B) {
	q := New()
	defer q.Close()
	nbr := b.N
	for i := 0; i < nbr; i++ {
		q.Enqueue(i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = q.Dequeue()
	}
}

func BenchmarkWOnlyWithInt(b *testing.B) {
	q := New()
	defer q.Close()
	for i := 0; i < b.N; i++ {
		q.Enqueue(i)
	}
}
