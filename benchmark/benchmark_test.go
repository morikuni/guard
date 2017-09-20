package benchmark

import (
	"context"
	"testing"
	"time"

	"github.com/morikuni/guard"
	"github.com/morikuni/guard/circuitbreaker"
	"github.com/morikuni/guard/panicguard"
	"github.com/morikuni/guard/retry"
	"github.com/morikuni/guard/semaphore"
)

func BenchmarkBase(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			func(ctx context.Context) error {
				return nil
			}(context.Background())
		}
	})
}

func BenchmarkRetry(b *testing.B) {
	g := retry.New(retry.Inf, guard.NewNoBackoff())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Run(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}
	})
}

func BenchmarkPanicguard(b *testing.B) {
	g := panicguard.New()

	b.Run("without panic", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				g.Run(context.Background(), func(ctx context.Context) error {
					return nil
				})
			}
		})
	})

	b.Run("with panic", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			for pb.Next() {
				g.Run(context.Background(), func(ctx context.Context) error {
					panic("hello")
				})
			}
		})
	})
}

func BenchmarkSemaphore(b *testing.B) {
	g := semaphore.New(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Run(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}
	})
}

func BenchmarkCircuitBreaker(b *testing.B) {
	window := circuitbreaker.NewCountBaseWindow(100)
	g := circuitbreaker.New(window, 0.2, guard.NewNoBackoff())

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			g.Run(context.Background(), func(ctx context.Context) error {
				return nil
			})
		}
	})
}

func BenchmarkCountBaseWindow(b *testing.B) {
	window := circuitbreaker.NewCountBaseWindow(100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			window.PutFailure()
			window.PutSuccess()
			window.FailureRate()
		}
	})
}

func BenchmarkTimeBaseWindow(b *testing.B) {
	window := circuitbreaker.NewTimeBaseWindow(time.Minute, time.Millisecond*100)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			window.PutFailure()
			window.PutSuccess()
			window.FailureRate()
		}
	})
}
