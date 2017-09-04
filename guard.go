// Package guard wraps context based process and manage it.
package guard

import (
	"context"
)

// Guard is a process manager that calls and manages a function.
type Guard interface {
	// Call calls function and manage it.
	Call(ctx context.Context, f func(context.Context) error) error
}

// GuardFunc is an adapter to use a function as Guard.
type GuardFunc func(ctx context.Context, f func(context.Context) error) error

// Call implements Guard.
func (g GuardFunc) Call(ctx context.Context, f func(context.Context) error) error {
	return g(ctx, f)
}

// Compose composes multiple Guards ans create a single Guard.
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
