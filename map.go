package channels

// Map reads values from the input channel and applies the provided `mapFn`
// to each value before pushing it to the output channel.  The output channel
// will have the same capacity as the input channel.  The output channel is
// closed once the input channel is closed and all mapped values pushed to
// the output channel.  The type of the output channel does not need to match
// the type of the input channel.
func Map[TIn any, TOut any](inc <-chan TIn, mapFn func(TIn) TOut) <-chan TOut {
	outc := make(chan TOut, cap(inc))

	go func() {
		defer close(outc)
		for in := range inc {
			outc <- mapFn(in)
		}
	}()

	return outc
}

// Like Map, but blocks until the input channel is closed and all values are read.
// MapsValues reads all values from the input channel and returns an array of
// values returned from passing each input value into `mapFn`.
func MapValues[TIn any, TOut any](inc <-chan TIn, mapFn func(TIn) TOut) []TOut {
	outc := Map(inc, mapFn)
	result := make([]TOut, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}