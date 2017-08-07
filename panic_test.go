package guard

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanic(t *testing.T) {
	t.Run("error should be returned from Call method", func(t *testing.T) {
		assert := assert.New(t)

		g := Panic()

		err := g.Call(context.Background(), func(_ context.Context) error {
			return errors.New("test error")
		})

		assert.EqualError(err, "test error")
	})

	t.Run("panic should be catched and wrapped", func(t *testing.T) {
		assert := assert.New(t)

		g := Panic()

		err := g.Call(context.Background(), func(ctx context.Context) error {
			panic("test error")
		})

		assert.Equal(PanicOccured{"test error"}, err)
	})
}
