package guard

import (
	"math/rand"
	"sync/atomic"
	"time"
)

type BackoffStrategy interface {
	NextInterval() time.Duration
	Reset() BackoffStrategy
}

func ConstantBackoff(d time.Duration) BackoffStrategy {
	return &constantBackoff{d}
}

type constantBackoff struct {
	Interval time.Duration
}

func (c *constantBackoff) NextInterval() time.Duration {
	return c.Interval
}

func (c *constantBackoff) Reset() BackoffStrategy {
	return c
}

func NoBackoff() BackoffStrategy {
	return noBackoff{}
}

type noBackoff struct{}

func (n noBackoff) NextInterval() time.Duration {
	return 0
}

func (n noBackoff) Reset() BackoffStrategy {
	return n
}

func ExponentialBackoff(options ...ExponentialBackoffOption) BackoffStrategy {
	e := &exponentialBackoff{
		initialInterval:     float64(200 * time.Millisecond),
		maxInterval:         float64(time.Minute),
		multiplier:          2,
		randomizationFactor: 0.2,
		retryCount:          0,
	}

	for _, o := range options {
		o(e)
	}

	if e.randomizer == nil {
		e.randomizer = rand.New(rand.NewSource(time.Now().Unix()))
	}

	return e
}

type exponentialBackoff struct {
	initialInterval     float64
	maxInterval         float64
	multiplier          float64
	randomizationFactor float64
	randomizer          Randomizer
	retryCount          int64
}

func (e *exponentialBackoff) NextInterval() time.Duration {
	n := e.retryCount

	interval := e.initialInterval
	for i := int64(0); i < n; i++ {
		interval *= e.multiplier
	}

	if interval > e.maxInterval {
		interval = e.maxInterval
	} else {
		atomic.CompareAndSwapInt64(&e.retryCount, n, n+1)
	}

	rnd := (1 - e.randomizationFactor) + (2 * e.randomizationFactor * e.randomizer.Float64())
	nextBackoff := time.Duration(interval * rnd)

	return nextBackoff
}

func (e *exponentialBackoff) Reset() BackoffStrategy {
	clone := *e
	clone.retryCount = 0
	return &clone
}

type ExponentialBackoffOption func(*exponentialBackoff)

func WithInitialInterval(d time.Duration) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.initialInterval = float64(d)
	})
}

func WithMaxInterval(d time.Duration) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.maxInterval = float64(d)
	})
}

func WithMultiplier(f float64) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.multiplier = f
	})
}

func WithRandomizationFactor(f float64) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.randomizationFactor = f
	})
}

func WithRandomizer(r Randomizer) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.randomizer = r
	})
}
