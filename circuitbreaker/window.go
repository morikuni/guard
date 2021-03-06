package circuitbreaker

import (
	"sync"
)

// Window accumulate successes and failures for circuit breaker.
type Window interface {
	// FailureRate returns failure rate that is calculated by
	// `failures / (successes + failures)`.
	FailureRate() float64

	// PutSuccess notify a success to the window.
	PutSuccess()

	// PutSuccess notify a failure to the window.
	PutFailure()

	// Reset resets the accumulated events.
	Reset()
}

type countBaseWindow struct {
	idx            int
	failures       int
	size           int
	failureHistory []bool
	mu             sync.RWMutex
}

// NewCountBaseWindow creates a Window that accumulates events up to given size.
// The window is represented by a ring buffer, so old events are overwrited by new events.
func NewCountBaseWindow(size int) Window {
	return &countBaseWindow{
		size:           size,
		failureHistory: make([]bool, size),
	}
}

func (w *countBaseWindow) FailureRate() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return float64(w.failures) / float64(w.size)
}

func (w *countBaseWindow) PutSuccess() {
	w.put(false)
}

func (w *countBaseWindow) PutFailure() {
	w.put(true)
}

func (w *countBaseWindow) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := range w.failureHistory {
		w.failureHistory[i] = false
	}
	w.failures = 0
	w.idx = 0
}

func (w *countBaseWindow) put(failure bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.failureHistory[w.idx] {
		w.failures--
	}
	if failure {
		w.failures++
	}

	w.failureHistory[w.idx] = failure
	w.idx = (w.idx + 1) % w.size
}
