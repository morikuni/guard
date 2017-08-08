package guard

import (
	"context"
	"time"
)

func Retry(n int, backoffStrategy BackoffStrategy) Guard {
	return GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		backoff := backoffStrategy.Reset()

		var err error
		for i := 0; i <= n; i++ {
			if i > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.NewTimer(backoff.NextInterval()).C:
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
