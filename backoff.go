package guard

import (
	"math"
	"math/rand"
	"sync/atomic"
	"time"
)

// Backoff is a strategy of backoff interval.
type Backoff interface {
	// NextInterval returns the next interval.
	NextInterval() time.Duration

	// Reset creates the clone of the current strategy with an initialized state.
	Reset() Backoff
}

// ConstantBackoff creates Backoff with a constant interval.
// NextInterval() always returns given parameter d.
func ConstantBackoff(d time.Duration) Backoff {
	return &constantBackoff{d}
}

type constantBackoff struct {
	Interval time.Duration
}

func (c *constantBackoff) NextInterval() time.Duration {
	return c.Interval
}

func (c *constantBackoff) Reset() Backoff {
	return c
}

// NoBackoff creates Backoff without an interval.
// NextInterval() always returns 0.
func NoBackoff() Backoff {
	return noBackoff{}
}

type noBackoff struct{}

func (n noBackoff) NextInterval() time.Duration {
	return 0
}

func (n noBackoff) Reset() Backoff {
	return n
}

// ExponentialBackoff creates Backoff with an exponential backoff.
//
// Let N be a retry count of the process, the value of NextInterval(N) is calculated by following formula.
//
//  NextInterval(N) = BaseInterval(N) * [1-RandomizationFactor, 1+RandomizationFactor)
//  BaseInterval(N) = min(BaseInterval(N-1) * Multiplier, MaxInterval)
//  BaseInterval(1) = min(InitialInterval, MaxInterval)
//
// The default parameters.
//
//  InitialInterval:     200 (ms)
//  MaxInterval:         1 (min)
//  Multiplier:          2
//  RandomizationFactor: 0.2
//  Randomizer:          rand.New(rand.NewSource(time.Now().Unix()))
//
// Example intervals.
//
//  +----+----------------------+----------------------+
//  | N  | BaseInterval(N) (ms) | NextInterval(N) (ms) |
//  +----+----------------------+----------------------+
//  |  1 |                  200 | [160, 240)           |
//  |  2 |                  400 | [320, 480)           |
//  |  3 |                  800 | [640, 960)           |
//  |  4 |                 1600 | [1280, 1920)         |
//  |  5 |                 3200 | [2560, 3840)         |
//  |  6 |                 6400 | [5120, 7680)         |
//  |  7 |                12800 | [10240, 15360)       |
//  |  8 |                25600 | [20480, 30720)       |
//  |  9 |                51200 | [40960, 61440)       |
//  | 10 |                60000 | [48000, 72000)       |
//  | 11 |                60000 | [48000, 72000)       |
//  +----+----------------------+----------------------+
//
// Note: MaxInterval effects only the base interval.
// The actual interval may exceed MaxInterval depending on RandomizationFactor.
func ExponentialBackoff(options ...ExponentialBackoffOption) Backoff {
	e := &exponentialBackoff{
		initialInterval:     float64(200 * time.Millisecond),
		maxInterval:         float64(time.Minute),
		multiplier:          2,
		randomizationFactor: 0.2,
	}

	for _, o := range options {
		o(e)
	}

	if e.randomizer == nil {
		e.randomizer = rand.New(rand.NewSource(time.Now().Unix()))
	}
	e.baseInterval = math.Float64bits(e.initialInterval)

	return e
}

type exponentialBackoff struct {
	initialInterval     float64
	maxInterval         float64
	multiplier          float64
	randomizationFactor float64
	randomizer          Randomizer

	baseInterval uint64 // baseInterval actually represents float64. use uint64 for CompareAndSwap.
}

func (e *exponentialBackoff) NextInterval() time.Duration {
	baseInterval := e.BaseInterval()

	rnd := (1 - e.randomizationFactor) + (2 * e.randomizationFactor * e.randomizer.Float64())
	nextBackoff := time.Duration(baseInterval * rnd)

	return nextBackoff
}

func (e *exponentialBackoff) BaseInterval() float64 {
	for {
		old := atomic.LoadUint64(&e.baseInterval)
		baseInterval := math.Float64frombits(old)
		new := baseInterval * e.multiplier

		if new > e.maxInterval {
			new = e.maxInterval
		}
		if atomic.CompareAndSwapUint64(&e.baseInterval, old, math.Float64bits(new)) {
			return baseInterval
		}
	}
}

func (e *exponentialBackoff) Reset() Backoff {
	clone := *e
	clone.baseInterval = math.Float64bits(clone.initialInterval)
	return &clone
}

// ExponentialBackoffOption is the optional parameter for ExponentialBackoff.
type ExponentialBackoffOption func(*exponentialBackoff)

// WithInitialInterval set the initial interval of ExponentialBackoff.
func WithInitialInterval(d time.Duration) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.initialInterval = float64(d)
	})
}

// WithMaxInterval set the maximum interval of ExponentialBackoff.
func WithMaxInterval(d time.Duration) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.maxInterval = float64(d)
	})
}

// WithMultiplier set the multiplier of ExponentialBackoff.
func WithMultiplier(f float64) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.multiplier = f
	})
}

// WithRandomizationFactor set the randomization factor of ExponentialBackoff.
func WithRandomizationFactor(f float64) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.randomizationFactor = f
	})
}

// WithRandomizer set the randomizer of ExponentialBackoff.
func WithRandomizer(r Randomizer) ExponentialBackoffOption {
	return ExponentialBackoffOption(func(e *exponentialBackoff) {
		e.randomizer = r
	})
}
