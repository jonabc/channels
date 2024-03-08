package channels

import "sync"

// Split reads values from the input channel and routes the values into `N`
// output channels using the provided `splitFn`.  The channel slice provided
// to `splitFn` will have the same length and order as the channel slice
// returned from the function, e.g. in the above example `Split` guarantees
// that chans[0] will hold even values and chans[1] will hold odd values.

// Each output channel will have the same capacity as the input channel and
// will be closed after the input channel is closed and emptied.
func Split[T any](inc <-chan T, count int, splitFn func(T, []chan<- T)) []<-chan T {
	writeOutc := make([]chan<- T, count)
	readOutc := make([]<-chan T, count)
	for i := 0; i < count; i++ {
		c := make(chan T, cap(inc))
		writeOutc[i] = c
		readOutc[i] = c
	}

	go func() {
		defer func() {
			for _, c := range writeOutc {
				close(c)
			}
		}()

		for in := range inc {
			splitFn(in, writeOutc)
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
func SplitValues[T any](inc <-chan T, count int, splitFn func(T, []chan<- T)) [][]T {
	outc := Split(inc, count, splitFn)
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