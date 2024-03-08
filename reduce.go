package channels

// Reduce reads values from the input channel and applies the provided `reduceFn` to each value.
// The first argument to the reducer function is the accumulated reduced value, which is either
// 1. the default value for the type on the first call
// 2. the output from the previous call of the reducer function for all other iterations.

// The output of each call to the reducer function is pushed to the output channel.
func Reduce[TIn any, TOut any](inc <-chan TIn, reduceFn func(TOut, TIn) TOut) <-chan TOut {
	outc := make(chan TOut, cap(inc))

	go func() {
		defer close(outc)
		var result TOut
		for in := range inc {
			result = reduceFn(result, in)
			outc <- result
		}
	}()

	return outc
}

// Like Reduce, but blocks until the input channel is closed and all values are read.
// ReduceValues reads all values from the input channel and returns the value returned
// after all values from the input channel have been passed into `reduceFn`.
func ReduceValues[TIn any, TOut any](inc <-chan TIn, reduceFn func(TOut, TIn) TOut) TOut {
	outc := Reduce(inc, reduceFn)
	var result TOut
	for out := range outc {
		result = out
	}

	return result
}
