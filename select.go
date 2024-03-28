package channels

import (
	"time"

	"github.com/jonabc/channels/providers"
)

type SelectConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[SelectStats]
	capacity      int
}

// Selects values from the input channel that return true from the provided `selectFn`
// and pushes them to the output channel.  The output channel is unbuffered by default,
// and is closed once the input channel is closed and all selected values pushed to the output channel.
func Select[T any](inc <-chan T, selectFn func(T) bool, opts ...Option[SelectConfig]) <-chan T {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		for in := range inc {
			start := time.Now()
			selected := selectFn(in)
			duration := time.Since(start)

			if selected {
				outc <- in
			}

			tryProvideStats(SelectStats{Duration: duration, Selected: selected}, statsProvider)
		}
	}()

	return outc
}

// Like Select, but blocks until the input channel is closed and all values are read.
// SelectValues reads all values from the input channel and returns an array values
// that return true from the provided `selectFn` function.
func SelectValues[T any](inc <-chan T, selectFn func(T) bool, opts ...Option[SelectConfig]) []T {
	outc := Select(inc, selectFn, opts...)
	result := make([]T, 0, len(inc))
	for out := range outc {
		result = append(result, out)
	}

	return result
}
