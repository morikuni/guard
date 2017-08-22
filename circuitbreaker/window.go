package circuitbreaker

import (
	"sync"
)

type Window interface {
	FailureRate() float64
	PutSuccess()
	PutFailure()
	Reset()
}

type countBasedWindow struct {
	idx            int
	failures       int
	size           int
	failureHistory []bool
	mu             sync.RWMutex
}

func CountWindow(size int) Window {
	return &countBasedWindow{
		size:           size,
		failureHistory: make([]bool, size),
	}
}

func (w *countBasedWindow) FailureRate() float64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return float64(w.failures) / float64(w.size)
}

func (w *countBasedWindow) PutSuccess() {
	w.put(false)
}

func (w *countBasedWindow) PutFailure() {
	w.put(true)
}

func (w *countBasedWindow) Reset() {
	w.mu.Lock()
	defer w.mu.Unlock()
	for i := range w.failureHistory {
		w.failureHistory[i] = false
	}
	w.failures = 0
	w.idx = 0
}

func (w *countBasedWindow) put(failure bool) {
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
