package channels

type ReduceConfig struct {
	capacity int
	panc     chan<- any
}

func defaultReduceOptions[T any](inc <-chan T) []Option[ReduceConfig] {
	return []Option[ReduceConfig]{
		ChannelCapacityOption[ReduceConfig](cap(inc)),
	}
}

// Reduce reads values from the input channel and applies the provided `reduceFn` to each value.
// The first argument to the reducer function is the accumulated reduced value, which is either
// 1. the default value for the type on the first call
// 2. the output from the previous call of the reducer function for all other iterations.

// The output of each call to the reducer function is pushed to the output channel.
func Reduce[TIn any, TOut any](inc <-chan TIn, reduceFn func(TOut, TIn) (TOut, bool), opts ...Option[ReduceConfig]) <-chan TOut {
	cfg := parseOpts(append(defaultReduceOptions(inc), opts...)...)

	outc := make(chan TOut, cfg.capacity)
	panc := cfg.panc

	go func() {
		defer handlePanicIfErrc(panc)
		defer close(outc)

		var result TOut
		for in := range inc {
			next, ok := reduceFn(result, in)
			if !ok {
				continue
			}
			result = next
			outc <- result
		}
	}()

	return outc
}

// Like Reduce, but blocks until the input channel is closed and all values are read.
// ReduceValues reads all values from the input channel and returns the value returned
// after all values from the input channel have been passed into `reduceFn`.
func ReduceValues[TIn any, TOut any](inc <-chan TIn, reduceFn func(TOut, TIn) (TOut, bool), opts ...Option[ReduceConfig]) TOut {
	outc := Reduce(inc, reduceFn, opts...)
	var result TOut
	for out := range outc {
		result = out
	}

	return result
}
