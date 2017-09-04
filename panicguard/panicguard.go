// Package panicguard provide a panic-safe process management.
package panicguard

import (
	"context"
	"fmt"

	"github.com/morikuni/guard"
)

// PanicOccured is a error that is returned when panicguard recovers a panic.
type PanicOccured struct {
	// Reason is a original panic reason.
	Reason interface{}
}

// Error implements error.
func (po PanicOccured) Error() string {
	return fmt.Sprintf("panic occured: %v", po.Reason)
}

// New creates a new guard.Guard with capability of recovering panic.
func New() guard.Guard {
	return guard.GuardFunc(func(ctx context.Context, f func(context.Context) error) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = PanicOccured{e}
			}
		}()
		return f(ctx)
	})
}
