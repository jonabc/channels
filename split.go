package channels

import (
	"sync"
	"time"

	"github.com/jonabc/channels/providers"
)

type SplitConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[Stats]
	capacities    []int
}

func defaultSplitOptions(count int) []Option[SplitConfig] {
	capacities := make([]int, count)
	for i := 0; i < count; i++ {
		capacities[i] = 1
	}
	return []Option[SplitConfig]{
		MultiChannelCapacitiesOption[SplitConfig](capacities),
	}
}

// Split reads values from the input channel and routes the values into `N`
// output channels using the provided `splitFn`.  The channel slice provided
// to `splitFn` will have the same length and order as the channel slice
// returned from the function, e.g. in the above example `Split` guarantees
// that chans[0] will hold even values and chans[1] will hold odd values.

// Each output channel will have the same capacity as the input channel and
// will be closed after the input channel is closed and emptied.
func Split[T any](inc <-chan T, count int, splitFn func(T, []chan<- T), opts ...Option[SplitConfig]) []<-chan T {
	cfg := parseOpts(append(defaultSplitOptions(count), opts...)...)

	writeOutc := make([]chan<- T, count)
	readOutc := make([]<-chan T, count)
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	for i := 0; i < count; i++ {
		c := make(chan T, cfg.capacities[i])
		writeOutc[i] = c
		readOutc[i] = c
	}

	go func() {
		defer tryHandlePanic(panicProvider)
		defer func() {
			for _, c := range writeOutc {
				close(c)
			}
		}()

		for in := range inc {
			start := time.Now()
			splitFn(in, writeOutc)
			duration := time.Since(start)
			tryProvideStats(Stats{Duration: duration}, statsProvider)
		}
	}()

	return readOutc
}

// Like Split, but blocks until the input channel is closed and all values are read.
// SplitValues reads all values from the input channel and returns `[][]T`, a
// two-dimensional slice containing the results from each split channel.
//   - The first dimension, `i` in `[i][j]T` matches the size and order of channels
//     provided to `splitFn`
//   - The second dimension, `j` in `[i][j]T` matches the size and order of values
//     written to `chans[i]` in `splitFn`
func SplitValues[T any](inc <-chan T, count int, splitFn func(T, []chan<- T), opts ...Option[SplitConfig]) [][]T {
	outc := Split(inc, count, splitFn, opts...)
	results := make([][]T, count)

	var wg sync.WaitGroup
	wg.Add(len(outc))

	for i := 0; i < count; i++ {
		go func(i int) {
			defer wg.Done()

			results[i] = make([]T, 0)
			for result := range outc[i] {
				results[i] = append(results[i], result)
			}
		}(i)
	}

	wg.Wait()
	return results
}

// Helper function that wraps Split, returning two output channels.
// See Split for additional details.
func Split2[T any](inc <-chan T, splitFn func(T, []chan<- T), opts ...Option[SplitConfig]) (<-chan T, <-chan T) {
	c := Split(inc, 2, splitFn, opts...)
	return c[0], c[1]
}

// Helper function that wraps Split, returning three output channels.
// See Split for additional details.
func Split3[T any](inc <-chan T, splitFn func(T, []chan<- T), opts ...Option[SplitConfig]) (<-chan T, <-chan T, <-chan T) {
	c := Split(inc, 3, splitFn, opts...)
	return c[0], c[1], c[2]
}
