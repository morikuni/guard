package semaphore

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSemaphore(t *testing.T) {
	t.Run("error should be returned", func(t *testing.T) {
		assert := assert.New(t)

		g := New(3)

		err := g.Call(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})

		assert.EqualError(err, "test error")
	})

	t.Run("err should be returnd immediately when the context is cancelled", func(t *testing.T) {
		assert := assert.New(t)

		g := New(0)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := g.Call(ctx, func(ctx context.Context) error {
			return errors.New("test error")
		})

		assert.Equal(context.Canceled, err)
	})

	t.Run("function should be executed concurrently expected numbers", func(t *testing.T) {
		assert := assert.New(t)

		n := 3
		g := New(n)

		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		var count int32 = 0
		for i := 0; i < 10; i++ {
			go g.Call(ctx, func(ctx context.Context) error {
				atomic.AddInt32(&count, 1)
				<-ctx.Done()
				return nil
			})
		}

		time.Sleep(10 * time.Millisecond)

		assert.Equal(int32(n), atomic.LoadInt32(&count))
	})
}
