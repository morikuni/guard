// Package guard wraps context based process and manage it.
package guard

import (
	"context"
)

// Guard is a process manager that runs and manages the process.
type Guard interface {
	// Run runs the function with its own capability.
	Run(ctx context.Context, f func(context.Context) error) error
}

// GuardFunc is an adapter to use a function as Guard.
type GuardFunc func(ctx context.Context, f func(context.Context) error) error

// Run implements Guard.
func (g GuardFunc) Run(ctx context.Context, f func(context.Context) error) error {
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
			return head.Run(ctx, func(ctx2 context.Context) error { return tail.Run(ctx2, f) })
		})
	}
}
