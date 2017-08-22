package retry

import (
	"context"
	"time"

	"github.com/morikuni/guard"
)

func New(n int, backoff guard.Backoff) guard.Guard {
	return guard.GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		bo := backoff.Reset()

		var err error
		for i := 0; i <= n; i++ {
			if i > 0 {
				t := time.NewTimer(bo.NextInterval())
				defer t.Stop()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-t.C:
				}
			}

			err = f(ctx)
			if err == nil {
				return nil
			}
		}
		return err
	})
}
