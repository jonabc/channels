package channels

import (
	"time"

	"github.com/jonabc/channels/providers"
)

type TapConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[TapStats]
	capacity      int
}

// Tap reads values from the input channel and calls the provided
// `[pre/post]Fn` functions with each value before and after writing
// the value to the output channel, respectivel.  The output channel
// is unbuffered by default, and will be closed after the input channel
// is closed and drained.
func Tap[T any](inc <-chan T, preFn, postFn func(T), opts ...Option[TapConfig]) <-chan T {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		for val := range inc {
			start := time.Now()
			if preFn != nil {
				preFn(val)
			}
			preDuration := time.Since(start)

			outc <- val

			start = time.Now()
			if postFn != nil {
				postFn(val)
			}
			postDuration := time.Since(start)

			tryProvideStats(TapStats{PreDuration: preDuration, PostDuration: postDuration}, statsProvider)
		}
	}()

	return outc
}
