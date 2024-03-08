package channels

// FlatMap reads values from the input channel and applies the provided `mapFn`
// to each value.  Each element in the slice returned by `mapFn` is then sent to the
// output channel.
// The output channel will have the same capacity as the input channel, and is
// closed once the input channel is closed and all mapped values are pushed to
// the output channel.
func FlatMap[TIn any, TOut any, TOutSlice []TOut](inc <-chan TIn, mapFn func(TIn) TOutSlice) <-chan TOut {
	outc := make(chan TOut, cap(inc))

	go func() {
		defer close(outc)
		for in := range inc {
			outSlice := mapFn(in)
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
func FlatMapValues[TIn any, TOut any, TOutSlice []TOut](inc <-chan TIn, mapFn func(TIn) TOutSlice) []TOut {
	outc := FlatMap(inc, mapFn)
	result := make([]TOut, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}
