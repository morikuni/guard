package guard

import (
	"context"
)

func Semaphore(n int) Guard {
	ch := make(chan struct{}, n)
	return GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		select {
		case ch <- struct{}{}:
			defer func() { <-ch }()
		case <-ctx.Done():
			return ctx.Err()
		}
		return f(ctx)
	})
}
