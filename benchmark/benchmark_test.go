package benchmark

import (
	"context"
	"testing"

	"github.com/morikuni/guard"
	"github.com/morikuni/guard/circuitbreaker"
	"github.com/morikuni/guard/panicguard"
	"github.com/morikuni/guard/retry"
	"github.com/morikuni/guard/semaphore"
)

func BenchmarkBase(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		func(ctx context.Context) error {
			return nil
		}(context.Background())
	}
}

func BenchmarkRetry(b *testing.B) {
	g := retry.New(retry.Inf, guard.NewNoBackoff())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Call(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkPanicguard(b *testing.B) {
	g := panicguard.New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Call(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkSemaphore(b *testing.B) {
	g := semaphore.New(3)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Call(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}
}

func BenchmarkCircuitBreaker(b *testing.B) {
	window := circuitbreaker.NewCountBaseWindow(100)
	g := circuitbreaker.New(window, 0.2, guard.NewNoBackoff())

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		g.Call(context.Background(), func(ctx context.Context) error {
			return nil
		})
	}
}
