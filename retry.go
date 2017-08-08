package guard

import (
	"context"
	"time"
)

func Retry(n int, backoffStrategy BackoffStrategy) Guard {
	return GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
		backoff := backoffStrategy.Reset()

		var err error
		for i := 0; i < n; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			err = f(ctx)
			if err == nil {
				return nil
			}
			time.Sleep(backoff.NextInterval())
		}
		return err
	})
}
