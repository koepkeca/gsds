// Package queue implements a thread safe queue in go.
package queue

import (
	"context"
)

// Q is the structure that contains the channel used to
// communicate with the queue.
type Q struct {
	op     chan (func(*queue))
	cancel context.CancelFunc
}

// Len will get return the number of items in the queue.
func (q *Q) Len() int64 {
	lChan := make(chan int64)
	q.op <- func(curr *queue) {
		lChan <- int64(len(*curr))
	}
	return <-lChan
}

// Dequeue removes the next item in the queue and returns it's value
func (q *Q) Dequeue() interface{} {
	vChan := make(chan interface{})
	q.op <- func(curr *queue) {
		if len(*curr) == 0 {
			vChan <- nil
			return
		}
		old := *curr
		n := len(old)
		val := old[0]
		*curr = old[1:n]
		vChan <- val
		return
	}
	return <-vChan
}

// Enqueue places a new item at the end of the queue
func (q *Q) Enqueue(v interface{}) {
	q.op <- func(curr *queue) {
		*curr = append(*curr, v)
		return
	}
}

// Front returns the value from the front of the queue.
// It does not remove it from the queue.
func (q *Q) Front() interface{} {
	vChan := make(chan interface{})
	q.op <- func(curr *queue) {
		if len(*curr) == 0 {
			vChan <- nil
			return
		}
		tmp := *curr
		vChan <- tmp[0]
		return
	}
	return <-vChan
}

// Back returns the value from the back of the queue.
// It does not remove it from the queue.
func (q *Q) Back() interface{} {
	vChan := make(chan interface{})
	q.op <- func(curr *queue) {
		if len(*curr) == 0 {
			vChan <- nil
			return
		}
		tmp := *curr
		vChan <- tmp[len(tmp)]
		return
	}
	return <-vChan
}

// Close closes the primary channel thus stopping
// the running go-routine.
func (q *Q) Close() {
	q.cancel()
	return
}

// New creates a new Safe Stack, this also starts the go-routine
// so once this is called, you need to clean up after yourself
// by using the Close method.
func New() (q *Q) {
	q = NewWithContext(nil)
	return
}

// NewWithContext creates a new queue which wil cancel upon context cancellation
func NewWithContext(ctx context.Context) (q *Q) {
	q = &Q{op: make(chan func(*queue))}
	ready := make(chan struct{})
	go q.loop(ctx, ready)
	<-ready
	return
}

// queue contains the queue represented as a slice of interfaces.
type queue []interface{}

// loop creates the guarded data structure and listens for
// methods on the op channel. loop terminates when the op
// channel is closed.
func (q *Q) loop(c context.Context, r chan struct{}) {
	st := &queue{}
	ctx, cancel := context.WithCancel(context.Background())
	if c != nil {
		ctx, cancel = context.WithCancel(c)
	}
	q.cancel = cancel
	close(r)
	for {
		select {
		case op := <-q.op:
			op(st)
		case <-ctx.Done():
			cancel()
			close(q.op)
			return
		}
	}
}
