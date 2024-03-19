package channels

// FlatMapConfig contains user configurable options for the FlatMap functions
type FlatMapConfig struct {
	panc     chan<- any
	capacity int
}

func defaultFlatMapOptions[T any](inc <-chan T) []Option[FlatMapConfig] {
	return []Option[FlatMapConfig]{
		ChannelCapacityOption[FlatMapConfig](cap(inc)),
	}
}

// FlatMap reads values from the input channel and applies the provided `mapFn`
// to each value.  Each element in the slice returned by `mapFn` is then sent to the
// output channel.
// The output channel will have the same capacity as the input channel, and is
// closed once the input channel is closed and all mapped values are pushed to
// the output channel.
func FlatMap[TIn any, TOut any, TOutSlice []TOut](inc <-chan TIn, mapFn func(TIn) (TOutSlice, bool), opts ...Option[FlatMapConfig]) <-chan TOut {
	cfg := parseOpts(append(defaultFlatMapOptions(inc), opts...)...)

	outc := make(chan TOut, cfg.capacity)
	panc := cfg.panc

	go func() {
		defer handlePanicIfErrc(panc)
		defer close(outc)

		for in := range inc {
			outSlice, ok := mapFn(in)
			if !ok {
				continue
			}

			for _, out := range outSlice {
				outc <- out
			}
		}
	}()

	return outc
}

// Like FlatMap, but blocks until the input channel is closed and all values are read.
// FlatMapValues reads all values from the input channel and returns a flattened array of
// values returned from passing each input value into `mapFn`.
func FlatMapValues[TIn any, TOut any, TOutSlice []TOut](inc <-chan TIn, mapFn func(TIn) (TOutSlice, bool), opts ...Option[FlatMapConfig]) []TOut {
	outc := FlatMap(inc, mapFn, opts...)
	result := make([]TOut, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}
