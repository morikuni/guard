// Package circuitbreaker provide a circuitbreaker pattern failure management.
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

// CircuitBreaker is a guard.Guard with an additional method.
type CircuitBreaker interface {
	guard.Guard

	// Subscribe returns a channel that receives events of state change of
	// the circuit breaker.
	Subscribe() <-chan StateChange
}

// New creates a new guard.Guard with capability of circuit breaker.
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

// ErrCircuitBreakerOpen is a error that is returned when the circuit breaker is "open".
var ErrCircuitBreakerOpen = errors.New("circuit breaker open")

type circuitBreaker struct {
	window     Window
	threashold float64
	state      int32
	backoff    guard.Backoff

	subscribers []chan<- StateChange
	mu          sync.RWMutex
}

func (cb *circuitBreaker) Run(ctx context.Context, f func(context.Context) error) error {
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

// StateChange is an event that represents a state change of the circuit breaker.
type StateChange int

const (
	// CloseToOpen is an event that state was changed from "close" to "open".
	CloseToOpen StateChange = iota
	// HalfOpenToOpen is an event that state was changed from "open" to "half-open".
	HalfOpenToOpen
	// HalfOpenToClose is an event that state was changed from "half-open" to "close".
	HalfOpenToClose
	// OpenToHalfOpen is an event that state was changed from "open" to "half-open".
	OpenToHalfOpen
)

// String implements fmt.Stringer.
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
