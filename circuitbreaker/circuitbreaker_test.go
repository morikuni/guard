package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/morikuni/guard"
	"github.com/stretchr/testify/assert"
)

func TestCircuitBreaker(t *testing.T) {
	t.Run("error should be returned", func(t *testing.T) {
		assert := assert.New(t)

		backoff := guard.NewNoBackoff()
		window := NewCountBaseWindow(3)
		g := New(window, 1, backoff)

		err := g.Run(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})

		assert.EqualError(err, "test error")
	})

	t.Run("circuitbreaker be open when error rate exceeds threashold", func(t *testing.T) {
		assert := assert.New(t)

		backoff := guard.NewConstantBackoff(time.Second)
		window := NewCountBaseWindow(1)
		g := New(window, 1, backoff)

		err := g.Run(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})
		assert.EqualError(err, "test error")

		called := false
		err = g.Run(context.Background(), func(_ context.Context) error {
			called = true
			return nil
		})
		assert.Equal(ErrCircuitBreakerOpen, err)
		assert.Equal(false, called)
	})

	t.Run("circuitbreaker be half-open after backoff", func(t *testing.T) {
		assert := assert.New(t)

		backoff := guard.NewNoBackoff()
		window := NewCountBaseWindow(1)
		g := New(window, 1, backoff)

		err := g.Run(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})
		assert.EqualError(err, "test error")

		time.Sleep(10 * time.Millisecond)

		called := false
		err = g.Run(context.Background(), func(_ context.Context) error {
			called = true
			return nil
		})
		assert.Nil(err)
		assert.Equal(true, called)
	})
}
