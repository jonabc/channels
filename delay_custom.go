package channels

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/jonabc/channels/providers"
)

// DelayConfig contains user configurable options for the Delay functions
type DelayConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[Stats]
	capacity      int
}

type Delayable interface {
	Delay() time.Duration
}

// DelayCustom is like Delay but with per-item configurability over delays.
// Where DelayCustom requires types that implement the `Delayable`
// interface.
func DelayCustom[T Delayable](inc <-chan T, opts ...Option[DelayConfig]) (<-chan T, func() int) {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)
	done := make(chan struct{})
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	var count atomic.Int32
	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		var wg sync.WaitGroup
		for next := range inc {
			wg.Add(1)
			count.Add(1)
			go func(item T) {
				defer tryHandlePanic(panicProvider)
				defer count.Add(-1)
				defer wg.Done()

				delay := item.Delay()
				if delay > 0 {
					timer := time.NewTimer(delay)
					select {
					case <-done:
						timer.Stop()
					case <-timer.C:
					}
				}

				outc <- item
				tryProvideStats(Stats{Duration: delay, QueueLength: len(inc)}, statsProvider)
			}(next)
		}

		close(done)
		wg.Wait()
	}()

	return outc, func() int { return int(count.Load()) }
}
