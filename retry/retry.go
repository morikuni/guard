// Package retry provides a failure management with retry for the process.
package retry

import (
	"context"
	"time"

	"github.com/morikuni/guard"
)

// Inf represents inifinite number of retries.
var Inf int = -1

// New creates a new guard.Guard with retry capability.
func New(n int, backoff guard.Backoff) guard.Guard {
	return guard.GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		err := f(ctx)
		if err == nil {
			return nil
		}

		bo := backoff.Reset()
		for i := 0; i < n || n < 0; i++ {
			err := sleep(ctx, bo.NextInterval())
			if err != nil {
				return err
			}

			err = f(ctx)
			if err == nil {
				return nil
			}
		}
		return err
	})
}

func sleep(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
