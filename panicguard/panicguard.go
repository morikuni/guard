package panicguard

import (
	"context"
	"fmt"

	"github.com/morikuni/guard"
)

type PanicOccured struct {
	Reason interface{}
}

func (po PanicOccured) Error() string {
	return fmt.Sprintf("panic occured: %v", po.Reason)
}

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
