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

func NewConstantBackoff(d time.Duration) BackoffStrategy {
	return &ConstantBackoff{d}
}

type ConstantBackoff struct {
	Interval time.Duration
}

func (c *ConstantBackoff) NextInterval() time.Duration {
	return c.Interval
}

func (c *ConstantBackoff) Reset() BackoffStrategy {
	return c
}

func NewNoBackoff() BackoffStrategy {
	return NoBackoff{}
}

type NoBackoff struct{}

func (n NoBackoff) NextInterval() time.Duration {
	return 0
}

func (n NoBackoff) Reset() BackoffStrategy {
	return n
}

func NewExponentialBackoff(options ...ExponentialBackoffOption) BackoffStrategy {
	e := &ExponentialBackoff{
		initialInterval:     float64(time.Second),
		maxInterval:         float64(time.Minute),
		multiplier:          2,
		randomizationFactor: 0.5,
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

type ExponentialBackoff struct {
	initialInterval     float64
	maxInterval         float64
	multiplier          float64
	randomizationFactor float64
	randomizer          Randomizer
	retryCount          int64
}

func (e *ExponentialBackoff) NextInterval() time.Duration {
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

func (e *ExponentialBackoff) Reset() BackoffStrategy {
	clone := *e
	clone.retryCount = 0
	return &clone
}

type ExponentialBackoffOption func(*ExponentialBackoff)

func WithInitialInterval(d time.Duration) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *ExponentialBackoff) {
		e.initialInterval = float64(d)
	})
}

func WithMaxInterval(d time.Duration) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *ExponentialBackoff) {
		e.maxInterval = float64(d)
	})
}

func WithMultiplier(f float64) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *ExponentialBackoff) {
		e.multiplier = f
	})
}

func WithRandomizationFactor(f float64) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *ExponentialBackoff) {
		e.randomizationFactor = f
	})
}

func WithRandomizer(r Randomizer) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *ExponentialBackoff) {
		e.randomizer = r
	})
}
