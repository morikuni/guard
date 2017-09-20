package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/morikuni/guard/circuitbreaker"
)

func main() {
	window := circuitbreaker.NewTimeBaseWindow(time.Second*1, time.Millisecond*100)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	ticker := time.NewTicker(time.Millisecond * 1000)
	defer ticker.Stop()
	enable := true
	mu := sync.RWMutex{}

	go func() {
		for _ = range ticker.C {
			fmt.Println(time.Now())
			fmt.Println(window)
			fmt.Println(window.FailureRate())
			fmt.Println()
			mu.Lock()
			// enable = !enable
			mu.Unlock()
		}
	}()

	wg := sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
		LOOP:
			for {
				select {
				case <-ctx.Done():
					break LOOP
				default:
					mu.RLock()
					e := enable
					mu.RUnlock()
					if e {
						if rand.Float64() < 0.1 {
							window.PutFailure()
						} else {
							window.PutSuccess()
						}
					}
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}
