// Package rate limit provides a manager to control the execution rate of the process.
package ratelimit

import (
	"context"

	"github.com/morikuni/guard"
)

// Limiter manage the state of rate limit.
type Limiter interface {
	// Wait sleeps until the process becomes executable according to
	// the state of rate limit.
	Wait(ctx context.Context) error
}

// New creates a new guard.Guard with capability of rate limit.
func New(limitter Limiter) guard.Guard {
	return guard.GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		if err := limitter.Wait(ctx); err != nil {
			return err
		}

		return f(ctx)
	})
}
