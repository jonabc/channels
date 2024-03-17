package channels

import "context"

// Map reads values from the input channel and applies the provided `mapFn`
// to each value before pushing it to the output channel.  The output channel
// will have the same capacity as the input channel.  The output channel is
// closed once the input channel is closed and all mapped values pushed to
// the output channel.  The type of the output channel does not need to match
// the type of the input channel.
func Map[TIn any, TOut any](ctx context.Context, inc <-chan TIn, mapFn func(context.Context, TIn) (TOut, bool)) <-chan TOut {
	outc := make(chan TOut, cap(inc))

	go func() {
		defer close(outc)
		for in := range inc {
			val, ok := mapFn(ctx, in)
			if !ok {
				continue
			}
			outc <- val
		}
	}()

	return outc
}

// Like Map, but blocks until the input channel is closed and all values are read.
// MapsValues reads all values from the input channel and returns an array of
// values returned from passing each input value into `mapFn`.
func MapValues[TIn any, TOut any](ctx context.Context, inc <-chan TIn, mapFn func(context.Context, TIn) (TOut, bool)) []TOut {
	outc := Map(ctx, inc, mapFn)
	result := make([]TOut, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}
