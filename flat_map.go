package channels

import (
	"time"

	"github.com/jonabc/channels/providers"
)

// FlatMapConfig contains user configurable options for the FlatMap functions
type FlatMapConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[Stats]
	capacity      int
}

// FlatMap reads values from the input channel and applies the provided `mapFn`
// to each value.  Each element in the slice returned by `mapFn` is then sent to the
// output channel.
// The output channel will have the same capacity as the input channel, and is
// closed once the input channel is closed and all mapped values are pushed to
// the output channel.
func FlatMap[TIn any, TOut any, TOutSlice []TOut](inc <-chan TIn, mapFn func(TIn) (TOutSlice, bool), opts ...Option[FlatMapConfig]) <-chan TOut {
	cfg := parseOpts(opts...)

	outc := make(chan TOut, cfg.capacity)
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		for in := range inc {
			start := time.Now()
			outSlice, ok := mapFn(in)
			duration := time.Since(start)

			if ok {
				for _, out := range outSlice {
					outc <- out
				}
			}

			tryProvideStats(Stats{Duration: duration}, statsProvider)
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
