package channels

import (
	"time"

	"github.com/jonabc/channels/providers"
)

type RejectConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[SelectStats]
	capacity      int
}

// Selects values from the input channel that return false from the provided `rejectFn`
// and pushes them to the output channel.  The output channel is unbuffered by default,
// and is closed once the input channel is closed and all selected values pushed to the output channel.
func Reject[T any](inc <-chan T, rejectFn func(T) bool, opts ...Option[RejectConfig]) <-chan T {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		for in := range inc {
			start := time.Now()
			selected := !rejectFn(in)
			duration := time.Since(start)

			if selected {
				outc <- in
			}

			tryProvideStats(SelectStats{Duration: duration, Selected: selected}, statsProvider)
		}
	}()

	return outc
}

// Like Reject, but blocks until the input channel is closed and all values are read.
// RejectValues reads all values from the input channel and returns an array of values
// that return false from the provided `rejectFn` function.
func RejectValues[T any](inc <-chan T, rejectFn func(T) bool, opts ...Option[RejectConfig]) []T {
	outc := Reject(inc, rejectFn, opts...)
	result := make([]T, 0, len(inc))
	for out := range outc {
		result = append(result, out)
	}

	return result
}
