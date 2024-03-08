package channels

// WithDone returns two channels: a channel containing piped input
// from the input channel as well and a channel which will be closed
// when the input channel has been closed and all values written to
// the piped output channel.
func WithDone[T any](inc <-chan T) (<-chan T, <-chan struct{}) {
	outc := make(chan T, cap(inc))
	signal := make(chan struct{})

	go func() {
		defer close(signal)
		defer close(outc)

		for in := range inc {
			outc <- in
		}
	}()

	return outc, signal
}
