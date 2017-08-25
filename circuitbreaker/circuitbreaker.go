package circuitbreaker

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/morikuni/guard"
)

type CircuitBreaker interface {
	guard.Guard

	Subscribe() <-chan StateChange
}

func New(window Window, threashold float64, backoff guard.Backoff) CircuitBreaker {
	window.Reset()
	cb := &circuitBreaker{
		window,
		threashold,
		close,
		backoff.Reset(),

		[]chan<- StateChange{},
		sync.RWMutex{},
	}
	return cb
}

const (
	close int32 = iota
	halfopen
	open
)

var ErrCircuitBreakerOpen = errors.New("circuit breaker open")

type circuitBreaker struct {
	window     Window
	threashold float64
	state      int32
	backoff    guard.Backoff

	subscribers []chan<- StateChange
	mu          sync.RWMutex
}

func (cb *circuitBreaker) Call(ctx context.Context, f func(context.Context) error) error {
	state, available := cb.currentState()
	if !available {
		return ErrCircuitBreakerOpen
	}

	err := f(ctx)
	switch err {
	case nil:
		cb.succeed(state)
	case context.Canceled:
		// this is normal, so do nothing.
	default:
		// context.DeadlimeExceeded and other errors are regarded as a failure.
		cb.fail(state)
	}
	return err
}

func (cb *circuitBreaker) succeed(state int32) {
	cb.window.PutSuccess()
	switch state {
	case close:
	case halfopen:
		cb.close()
	default:
		panic("never come here")
	}
}

func (cb *circuitBreaker) fail(state int32) {
	cb.window.PutFailure()
	switch state {
	case close:
		if cb.window.FailureRate() >= cb.threashold {
			cb.open(state)
		}
	case halfopen:
		cb.open(state)
	default:
		panic("never come here")
	}
}

func (cb *circuitBreaker) currentState() (state int32, available bool) {
	switch atomic.LoadInt32(&cb.state) {
	case close:
		return close, true
	case halfopen:
		return halfopen, true
	case open:
		return open, false
	}
	panic("never come here")
}

func (cb *circuitBreaker) change(from, to int32, sc StateChange) bool {
	ok := atomic.CompareAndSwapInt32(&cb.state, from, to)
	if ok {
		cb.notify(sc)
	}
	return ok
}

func (cb *circuitBreaker) open(state int32) {
	sc := CloseToOpen
	if state == halfopen {
		sc = HalfOpenToOpen
	}
	if cb.change(state, open, sc) {
		time.AfterFunc(cb.backoff.NextInterval(), func() {
			cb.change(open, halfopen, OpenToHalfOpen)
		})
	}
}

func (cb *circuitBreaker) close() {
	if cb.change(halfopen, close, HalfOpenToClose) {
		cb.backoff = cb.backoff.Reset()
		cb.window.Reset()
	}
}

func (cb *circuitBreaker) Subscribe() <-chan StateChange {
	c := make(chan StateChange, 100)

	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.subscribers = append(cb.subscribers, c)
	return c
}

func (cb *circuitBreaker) notify(sc StateChange) {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	for _, subscriber := range cb.subscribers {
		subscriber <- sc
	}
}

type StateChange int

const (
	CloseToOpen StateChange = iota
	HalfOpenToOpen
	HalfOpenToClose
	OpenToHalfOpen
)

func (sc StateChange) String() string {
	switch sc {
	case CloseToOpen:
		return "close to open"
	case HalfOpenToOpen:
		return "half-open to open"
	case HalfOpenToClose:
		return "half-open to close"
	case OpenToHalfOpen:
		return "open to half-open"
	default:
		panic(fmt.Sprint("unknown state change", int(sc)))
	}
}
