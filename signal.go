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

// WithSignal returns two channels: a channel containing piped values
// from the input channel and a channel which contains values returnd
// by `signalFn` when `signalFn` also returns true.
// Both the output channel and the signal channel are closed when
// the input channels is closed and emptied.
func WithSignal[T any, S any](inc <-chan T, signalFn func(T) (S, bool)) (<-chan T, <-chan S) {
	outc := make(chan T, cap(inc))
	signalc := make(chan S, cap(inc))

	go func() {
		defer close(signalc)
		defer close(outc)

		for in := range inc {
			if s, ok := signalFn(in); ok {
				signalc <- s
			}
			outc <- in
		}
	}()

	return outc, signalc
}
