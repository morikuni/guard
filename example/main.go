//
package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/morikuni/guard"
	"github.com/morikuni/guard/circuitbreaker"
	"github.com/morikuni/guard/panicguard"
	"github.com/morikuni/guard/ratelimit"
	"github.com/morikuni/guard/retry"
	"github.com/morikuni/guard/semaphore"
	"golang.org/x/time/rate"
)

func main() {
	wg := sync.WaitGroup{}
	backoff := guard.NewExponentialBackoff(
		guard.WithInitialInterval(50*time.Millisecond),
		guard.WithMaxInterval(2000*time.Millisecond),
		guard.WithMultiplier(2),
		guard.WithRandomizationFactor(0),
	)
	window := circuitbreaker.NewCountBaseWindow(3)
	cb := circuitbreaker.New(window, 0.5, backoff)
	g := guard.Compose(
		ratelimit.New(rate.NewLimiter(rate.Every(2*time.Second), 1)),
		retry.New(25, guard.NewConstantBackoff(100*time.Millisecond)),
		semaphore.New(100),
		cb,
		panicguard.New(),
	)

	// go func() {
	// 	for event := range cb.Subscribe() {
	// 		fmt.Println(event)
	// 	}
	// }()

	for i := 0; i < 10; i++ {
		x := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := 0
			ctx := context.Background()
			ctx, cancel := context.WithTimeout(ctx, time.Second*10)
			defer cancel()
			start := time.Now()
			err := g.Run(ctx, func(_ context.Context) error {
				count++
				now := time.Now()
				fmt.Println(x, count, now.Sub(start).Nanoseconds()/1000000)
				start = now
				panic("aho")
			})

			if err != nil {
				fmt.Println(x, count, err)
			}
		}()
	}
	wg.Wait()
}
