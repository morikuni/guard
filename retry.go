package guard

import (
	"context"
	"time"
)

func Retry(n int, backoffStrategy Backoff) Guard {
	return GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		backoff := backoffStrategy.Reset()

		var err error
		for i := 0; i <= n; i++ {
			if i > 0 {
				t := time.NewTimer(backoff.NextInterval())
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
