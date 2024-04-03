package channels

import "time"

type debounceValuesInput[T comparable] struct {
	val   T
	delay time.Duration
}

func (i *debounceValuesInput[T]) Key() T {
	return i.val
}

func (i *debounceValuesInput[T]) Delay() time.Duration {
	return i.delay
}

func (i *debounceValuesInput[T]) Reduce(*debounceValuesInput[T]) (*debounceValuesInput[T], bool) {
	return i, true
}

// DebounceValues is like Debounce but with per-value debouncing, where
// each unique value read from the input channel will start a debouncing
// period for that value.  Any duplicate values read from the input channel
// during a debouncing period are ignored.
func DebounceValues[T comparable](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)

	inBridge := make(chan *debounceValuesInput[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceValuesInput[T]{val: in, delay: delay}
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
