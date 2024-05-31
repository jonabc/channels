package channels

import (
	"time"
)

// Delay reads values from the input channel, and writes each value to the
// output channel after `delay` duration.

// The channel returned by Delay is unbuffered by default.
// When the input channel is closed, any remaining values being delayed
// will be flushed to the output channel and the output channel will be closed.

// Delay also returns a function which returns the number of values
// that are currently being delayed.
func Delay[T any](inc <-chan T, delay time.Duration, opts ...Option[DelayConfig]) (<-chan T, func() int) {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)

	inBridge := make(chan *debounceInput[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceInput[T]{val: in, delay: delay}
		}
	}()

	outBridge, getDelayedCount := DelayCustom(inBridge,
		append(opts, ChannelCapacityOption[DelayConfig](0))...,
	)
	go func() {
		defer close(outc)
		for out := range outBridge {
			outc <- out.val
		}
	}()

	return outc, getDelayedCount
}
