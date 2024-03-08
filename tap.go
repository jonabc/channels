package channels

// Tap reads values from the input channel and calls the provided `tapFn`
// with each value before writing the value to the output channel.  The
// output channel has the same capacity as the input channel, and will
// be closed after the input channel is closed and drained.
func Tap[T any](inc <-chan T, tapFn func(T)) <-chan T {
	outc := make(chan T, cap(inc))

	go func() {
		defer close(outc)
		for val := range inc {
			tapFn(val)
			outc <- val
		}
	}()

	return outc
}
