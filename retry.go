package guard

import (
	"context"
)

func Retry(n int) Guard {
	return GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
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
		}
		return err
	})
}
