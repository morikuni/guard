package guard

import (
	"context"
	"fmt"
)

type PanicOccured struct {
	Reason interface{}
}

func (po PanicOccured) Error() string {
	return fmt.Sprintf("panic occured: %v", po.Reason)
}

func Panic() Guard {
	return GuardFunc(func(ctx context.Context, f func(context.Context) error) (err error) {
		defer func() {
			if e := recover(); e != nil {
				err = PanicOccured{e}
			}
		}()
		return f(ctx)
	})
}
