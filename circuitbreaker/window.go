package circuitbreaker

import (
	"fmt"
	"sync"
	"time"
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

// NewCountBaseWindow creates a Window that accumulates events up to given size.
// The window is represented by a ring buffer, so old events are overwrited by new events.
func NewCountBaseWindow(size int) Window {
	return &countBaseWindow{
		size:           size,
		failureHistory: make([]bool, size),
	}
}

type countBaseWindow struct {
	idx            int
	failures       int
	size           int
	failureHistory []bool
	mu             sync.RWMutex
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

func NewTimeBaseWindow(duration, truncate time.Duration) Window {
	return &timeBaseWindow{
		overallDuration: overall.Nanoseconds(),
		slidingDuration: sliding.Nanoseconds(),
		mu:              sync.Mutex{},
	}
}

type timeBaseWindow struct {
	oldest          *counter
	latest          *counter
	overallDuration int64
	slidingDuration int64
	total           uint64
	failures        uint64
	mu              sync.Mutex
}

func (w *timeBaseWindow) FailureRate() float64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.cleanUpCounter()
	return float64(w.failures) / float64(w.total)
}

func (w *timeBaseWindow) PutSuccess() {
	w.put(false)
}

func (w *timeBaseWindow) PutFailure() {
	w.put(true)
}

func (w *timeBaseWindow) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.oldest = nil
	w.latest = nil
	w.total = 0
	w.failures = 0
}

func (w *timeBaseWindow) String() string {
	w.mu.Lock()
	defer w.mu.Unlock()
	return fmt.Sprintf("%v", w.oldest)
}

func (w *timeBaseWindow) put(failure bool) {
	w.mu.Lock()
	defer w.mu.Unlock()

	latest := w.cleanUpCounter()
	if failure {
		w.failures++
		latest.putFailure()
	} else {
		latest.putSuccess()
	}
	w.total++
}

func (w *timeBaseWindow) cleanUpCounter() *counter {
	now := time.Now().UnixNano()
	expirationTime := now - w.overallDuration
	for w.oldest != nil && w.oldest.isExpired(expirationTime) {
		w.shift()
	}
	if w.latest == nil || w.latest.isExpired(now) {
		w.push(now)
	}

	return w.latest
}

func (w *timeBaseWindow) shift() {
	c := w.oldest
	w.oldest = c.next
	if w.oldest == nil {
		w.latest = nil
	}
	w.total -= c.total
	w.failures -= c.failures
}

func (w *timeBaseWindow) push(now int64) {
	c := newCounter(now + w.slidingDuration)
	if w.latest != nil {
		w.latest.next = c
	}
	if w.oldest == nil {
		w.oldest = c
	}
	w.latest = c
}

func newCounter(expiresAt int64) *counter {
	return &counter{0, 0, nil, expiresAt}
}

type counter struct {
	total     uint64
	failures  uint64
	next      *counter
	expiresAt int64
}

func (c *counter) isExpired(at int64) bool {
	return c.expiresAt < at
}

func (c *counter) putSuccess() {
	c.total++
}

func (c *counter) putFailure() {
	c.total++
	c.failures++
}

func (c *counter) String() string {
	return fmt.Sprintf("%v %d %d\n- %s", time.Unix(0, c.expiresAt), c.failures, c.total, c.next)
}
