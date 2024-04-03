package channels

import (
	"time"
)

type debounceInput[T any] struct {
	val   T
	delay time.Duration
}

func (i *debounceInput[T]) Key() string {
	return ""
}

func (i *debounceInput[T]) Delay() time.Duration {
	return i.delay
}

func (i *debounceInput[T]) Reduce(*debounceInput[T]) (*debounceInput[T], bool) {
	return i, true
}

// Debounce reads values from the input channel, and for every distinct debouncing
// period of `delay` duration, writes a single value to the output channel either
// before, after, or before and after a `delay` debounce period.
// The `DebounceType` value used in the function controls when the value is pushed
// to the output channel.  Only the first value read from the input channel at the start
// of the debounce period will be written to the output channel.

// The channel returned by Debounce is unbuffered by default.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.

// Debounce also returns a function which returns the number of debounced values
// that are currently being delayed
func Debounce[T any](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)

	inBridge := make(chan *debounceInput[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceInput[T]{val: in, delay: delay}
		}
	}()

	outBridge, getDebouncedCount := DebounceCustom(inBridge,
		append(opts, ChannelCapacityOption[DebounceConfig](0))...,
	)
	go func() {
		defer close(outc)
		for out := range outBridge {
			outc <- out.val
		}
	}()

	return outc, getDebouncedCount
}
