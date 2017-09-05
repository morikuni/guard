package retry

import (
	"context"
	"errors"
	"testing"

	"github.com/morikuni/guard"
	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	noBackoff := guard.NewNoBackoff()

	t.Run("error should be returned", func(t *testing.T) {
		assert := assert.New(t)

		g := New(3, noBackoff)

		err := g.Run(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})

		assert.EqualError(err, "test error")
	})

	t.Run("function should be retried for the expected number of times", func(t *testing.T) {
		assert := assert.New(t)

		n := 3
		g := New(n, noBackoff)

		count := 0
		g.Run(context.Background(), func(ctx context.Context) error {
			count++
			return errors.New("test error")
		})

		assert.Equal(n+1, count) // 1 for first, n for retry
	})

	t.Run("err should be nil when function succeeds while retry", func(t *testing.T) {
		assert := assert.New(t)

		g := New(3, noBackoff)

		count := 0
		err := g.Run(context.Background(), func(ctx context.Context) error {
			count++
			if count == 2 {
				return nil
			}
			return errors.New("test error")
		})

		assert.NoError(err)
		assert.Equal(2, count)
	})

	t.Run("retry should be stopped when the context is cancelled", func(t *testing.T) {
		assert := assert.New(t)

		g := New(3, noBackoff)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		count := 0
		err := g.Run(ctx, func(ctx context.Context) error {
			count++
			if count == 2 {
				cancel()
			}
			return errors.New("test error")
		})

		assert.Equal(context.Canceled, err)
		assert.Equal(2, count)
	})
}
