package retry

import (
	"context"
	"time"

	"github.com/morikuni/guard"
)

var Inf int = -1

func New(n int, backoff guard.Backoff) guard.Guard {
	return guard.GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		bo := backoff.Reset()
		stop := stopper(n)

		for {
			err := f(ctx)
			if err == nil {
				return nil
			}

			if stop() {
				return err
			}

			t := time.NewTimer(bo.NextInterval())
			defer t.Stop()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
			}
		}
	})
}

func stopper(n int) func() bool {
	if n < -1 {
		return inf
	}
	return count(n)
}

func inf() bool {
	return false
}

func count(n int) func() bool {
	i := 0
	return func() bool {
		i++
		if i > n {
			return true
		}
		return false
	}
}
