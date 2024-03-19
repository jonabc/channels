package channels

type SelectConfig struct {
	panc     chan<- any
	capacity int
}

func defaultSelectOptions[T any](inc <-chan T) []Option[SelectConfig] {
	return []Option[SelectConfig]{
		ChannelCapacityOption[SelectConfig](cap(inc)),
	}
}

// Selects values from the input channel that return true from the provided `selectFn`
// and pushes them to the output channel.  The output channel will have the same capacity
// as the input channel.  The output channel is closed once the input channel is closed
// and all selected values pushed to the output channel.
func Select[T any](inc <-chan T, selectFn func(T) bool, opts ...Option[SelectConfig]) <-chan T {
	cfg := parseOpts(append(defaultSelectOptions(inc), opts...)...)

	outc := make(chan T, cfg.capacity)
	panc := cfg.panc

	go func() {
		defer handlePanicIfErrc(panc)
		defer close(outc)

		for in := range inc {
			if selectFn(in) {
				outc <- in
			}
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
