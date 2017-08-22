package semaphore

import (
	"context"

	"github.com/morikuni/guard"
)

func New(n int) guard.Guard {
	ch := make(chan struct{}, n)
	return guard.GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		select {
		case ch <- struct{}{}:
			defer func() { <-ch }()
		case <-ctx.Done():
			return ctx.Err()
		}
		return f(ctx)
	})
}
