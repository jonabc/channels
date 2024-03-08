package channels

// Selects values from the input channel that return true from the provided `selectFn`
// and pushes them to the output channel.  The output channel will have the same capacity
// as the input channel.  The output channel is closed once the input channel is closed
// and all selected values pushed to the output channel.
func Select[T any](inc <-chan T, selectFn func(T) bool) <-chan T {
	outc := make(chan T, cap(inc))

	go func() {
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
func SelectValues[T any](inc <-chan T, selectFn func(T) bool) []T {
	outc := Select(inc, selectFn)
	result := make([]T, 0, len(inc))
	for out := range outc {
		result = append(result, out)
	}

	return result
}

// Selects values from the input channel that return false from the provided `rejectFn`
// and pushes them to the output channel.  The output channel will have the same capacity
// as the input channel.  The output channel is closed once the input channel is closed
// and all selected values pushed to the output channel.
func Reject[T any](inc <-chan T, rejectFn func(T) bool) <-chan T {
	return Select(inc, func(t T) bool { return !rejectFn(t) })
}

// Like Reject, but blocks until the input channel is closed and all values are read.
// RejectValues reads all values from the input channel and returns an array of values
// that return false from the provided `rejectFn` function.
func RejectValues[T any](inc <-chan T, rejectFn func(T) bool) []T {
	return SelectValues(inc, func(t T) bool { return !rejectFn(t) })
}
