package circuitbreaker

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/morikuni/guard"
)

func New(window Window, threashold float64, backoff guard.Backoff) guard.Guard {
	window.Reset()
	cb := &circuitBreaker{
		window,
		threashold,
		close,
		backoff.Reset(),
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
		cb.close(state)
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

func (cb *circuitBreaker) change(from, to int32) bool {
	return atomic.CompareAndSwapInt32(&cb.state, from, to)
}

func (cb *circuitBreaker) open(state int32) {
	if cb.change(state, open) {
		time.AfterFunc(cb.backoff.NextInterval(), func() {
			cb.change(open, halfopen)
		})
	}
}

func (cb *circuitBreaker) close(state int32) {
	if cb.change(state, close) {
		cb.backoff = cb.backoff.Reset()
		cb.window.Reset()
	}
}
