package channels

import (
	"github.com/jonabc/channels/providers"
)

type RejectConfig struct {
	panicProvider providers.Provider[any]
	capacity      int
}

func defaultRejectOptions[T any](inc <-chan T) []Option[RejectConfig] {
	return []Option[RejectConfig]{
		ChannelCapacityOption[RejectConfig](cap(inc)),
	}
}

// Selects values from the input channel that return false from the provided `rejectFn`
// and pushes them to the output channel.  The output channel will have the same capacity
// as the input channel.  The output channel is closed once the input channel is closed
// and all selected values pushed to the output channel.
func Reject[T any](inc <-chan T, rejectFn func(T) bool, opts ...Option[RejectConfig]) <-chan T {
	cfg := parseOpts(append(defaultRejectOptions(inc), opts...)...)

	outc := make(chan T, cfg.capacity)
	panicProvider := cfg.panicProvider

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		for in := range inc {
			if !rejectFn(in) {
				outc <- in
			}
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
