package channels

type TapConfig struct {
	panc     chan<- any
	capacity int
}

func defaultTapOptions[T any](inc <-chan T) []Option[TapConfig] {
	return []Option[TapConfig]{
		ChannelCapacityOption[TapConfig](cap(inc)),
	}
}

// Tap reads values from the input channel and calls the provided
// `[pre/post]Fn` functions with each value before and after writing
// the value to the output channel, respectivel.  The output channel
// has the same capacity as the input channel, and will be closed
// after the input channel is closed and drained.
func Tap[T any](inc <-chan T, preFn func(T), postFn func(T), opts ...Option[TapConfig]) <-chan T {
	cfg := parseOpts(append(defaultTapOptions(inc), opts...)...)

	outc := make(chan T, cfg.capacity)
	panc := cfg.panc

	go func() {
		defer handlePanicIfErrc(panc)
		defer close(outc)

		for val := range inc {
			if preFn != nil {
				preFn(val)
			}
			outc <- val
			if postFn != nil {
				postFn(val)
			}
		}
	}()

	return outc
}
