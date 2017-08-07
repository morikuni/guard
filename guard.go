package guard

import (
	"context"
)

type Guard interface {
	Call(ctx context.Context, f func(context.Context) error) error
}

type GuardFunc func(ctx context.Context, f func(context.Context) error) error

func (g GuardFunc) Call(ctx context.Context, f func(context.Context) error) error {
	return g(ctx, f)
}

func Compose(guards ...Guard) Guard {
	switch len(guards) {
	case 0:
		return nil
	case 1:
		return guards[0]
	default:
		head := guards[0]
		tail := Compose(guards[1:]...)
		return GuardFunc(func(ctx context.Context, f func(context.Context) error) error {
			return head.Call(ctx, func(ctx2 context.Context) error { return tail.Call(ctx2, f) })
		})
	}
}
