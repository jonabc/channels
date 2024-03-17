package channels

import "context"

// Tap reads values from the input channel and calls the provided
// `[pre/post]Fn` functions with each value before and after writing
// the value to the output channel, respectivel.  The output channel
// has the same capacity as the input channel, and will be closed
// after the input channel is closed and drained.
func Tap[T any](ctx context.Context, inc <-chan T, preFn func(context.Context, T), postFn func(context.Context, T)) <-chan T {
	outc := make(chan T, cap(inc))

	go func() {
		defer close(outc)
		for val := range inc {
			if preFn != nil {
				preFn(ctx, val)
			}
			outc <- val
			if postFn != nil {
				postFn(ctx, val)
			}
		}
	}()

	return outc
}
