// Package semaphore provides a semaphore that limits a number of concurrent processes.
package semaphore

import (
	"context"

	"github.com/morikuni/guard"
)

// New creates a new guard.Guard with semaphore capability.
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
