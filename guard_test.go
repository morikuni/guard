package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testGuard struct {
	CallStack *[]int
	ID        int
}

func (g testGuard) Call(ctx context.Context, f func(context.Context) error) error {
	if g.CallStack != nil {
		*g.CallStack = append(*g.CallStack, g.ID)
	}
	return f(ctx)
}

func TestCompose(t *testing.T) {
	t.Run("first guard should be called first", func(t *testing.T) {
		assert := assert.New(t)

		stack := []int{}
		g := Compose(testGuard{&stack, 1}, testGuard{&stack, 2}, testGuard{&stack, 3})

		err := g.Call(context.Background(), func(_ context.Context) error {
			return nil
		})

		assert.NoError(err)
		assert.Equal([]int{1, 2, 3}, stack)
	})

	t.Run("error should be returned", func(t *testing.T) {
		assert := assert.New(t)

		g := Compose(testGuard{}, testGuard{}, testGuard{})

		err := g.Call(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})

		assert.EqualError(err, "test error")
	})

	t.Run("context should be passed to function", func(t *testing.T) {
		assert := assert.New(t)

		g := Compose(testGuard{}, testGuard{}, testGuard{})

		ctx := context.WithValue(context.Background(), "aaa", "bbb")
		err := g.Call(ctx, func(ctx context.Context) error {
			assert.Equal("bbb", ctx.Value("aaa"))
			return nil
		})

		assert.NoError(err)
	})
}
