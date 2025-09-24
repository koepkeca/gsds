// Package stack implements a thread safe stack in go.
package stack

import (
	"context"
)

// S is the structure that contains the channel and context
// used to communicate with the stack.
type S struct {
	op     chan (func(*stack))
	cancel context.CancelFunc
}

// Len will get return the number of items in the stack.
func (s *S) Len() (i int64) {
	lChan := make(chan int64)
	//If a panic is pointing you here, you might be
	//trying to operate on an already closed data structure.
	s.op <- func(curr *stack) {
		lChan <- int64(len(*curr))
	}
	return <-lChan
}

// Pop will perform a pop on the stack, removing the first item
// and returning it's value.
func (s *S) Pop() (v interface{}) {
	vChan := make(chan interface{})
	//If a panic is pointing you here, you might be
	//trying to operate on an already closed data structure.
	s.op <- func(curr *stack) {
		old := *curr
		n := len(old)
		if n == 0 {
			vChan <- nil
			return
		}
		item := old[n-1]
		*curr = old[0 : n-1]
		vChan <- item
		return
	}
	return <-vChan
}

// Push will push the value v onto the stack.
func (s *S) Push(v interface{}) {
	//If a panic is pointing you here, you might be
	//trying to operate on an already closed data structure.
	s.op <- func(curr *stack) {
		*curr = append(*curr, v)
		return
	}
	return
}

// Close closes the primary channel thus stopping
// the running go-routine.
func (s *S) Close() {
	//the cancellation here will close the channel in loop()
	s.cancel()
	return
}

// New creates a new Safe Stack, this also starts the go-routine
// so once this is called, you need to clean up after yourself
// by using the Close method.
func New() (s *S) {
	s = &S{op: make(chan func(*stack))}
	go s.loop(nil)
	return
}

// NewWithContext creates a new stack which will cancel upon context cancellation.
func NewWithContext(ctx context.Context) (s *S) {
	s = &S{op: make(chan func(*stack))}
	go s.loop(ctx)
	return
}

// We emulate a stack using an interface slice to reduce memory overhead
type stack []interface{}

// loop creates the guarded data structure and listens for
// methods on the op channel. loop terminates when the op
// channel is closed.
func (s *S) loop(c context.Context) {
	st := &stack{}
	ctx, cancel := context.WithCancel(context.Background())
	if c != nil {
		ctx, cancel = context.WithCancel(c)
	}
	s.cancel = cancel
	for {
		select {
		case op := <-s.op:
			op(st)
		case <-ctx.Done():
			cancel()
			close(s.op)
			return
		}
	}
}
