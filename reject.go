package channels

// Selects values from the input channel that return false from the provided `rejectFn`
// and pushes them to the output channel.  The output channel is unbuffered by default,
// and is closed once the input channel is closed and all selected values pushed to the output channel.
func Reject[T any](inc <-chan T, rejectFn func(T) bool, opts ...Option[SelectConfig]) <-chan T {
	return Select(inc, func(t T) bool { return !rejectFn(t) }, opts...)
}

// Like Reject, but blocks until the input channel is closed and all values are read.
// RejectValues reads all values from the input channel and returns an array of values
// that return false from the provided `rejectFn` function.
func RejectValues[T any](inc <-chan T, rejectFn func(T) bool, opts ...Option[SelectConfig]) []T {
	return SelectValues(inc, func(t T) bool { return !rejectFn(t) }, opts...)
}
