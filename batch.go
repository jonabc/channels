package channels

import (
	"time"

	internalTime "github.com/jonabc/channels/internal/time"
)

// Batch N values from the input channel into an array of N values in the output channel.
// The output channel will have capacity $`cap(input channel) / batchSize`$.
// The output channel is closed once the input channel is closed and a partial batch is
// sent to the output channel, if a partial batch exists.
func Batch[T any](inc <-chan T, batchSize int, maxDelay time.Duration) <-chan []T {
	channelCapacity := cap(inc) / batchSize
	outc := make(chan []T, channelCapacity)
	buffer := make([]T, 0, batchSize)

	ticker := internalTime.NewTicker(maxDelay)

	publishAndReset := func() {
		keys := make([]T, len(buffer))
		copy(keys, buffer)
		outc <- keys
		buffer = buffer[:0]
		ticker.Reset(maxDelay)
	}

	go func() {
		defer close(outc)
		defer ticker.Stop()
		for {
			select {
			case in, ok := <-inc:
				if !ok {
					publishAndReset()
					return
				}

				buffer = append(buffer, in)
				if len(buffer) == cap(buffer) {
					publishAndReset()
				}
			case <-ticker.C:
				if len(buffer) > 0 {
					publishAndReset()
				}
			}
		}
	}()

	return outc
}

// Like Batch, but blocks until the input channel is closed and all values are read.
// BatchValue reads all values from the input channel and returns an array of batches.
func BatchValues[T any](inc <-chan T, batchSize int, maxDelay time.Duration) [][]T {
	outc := Batch(inc, batchSize, maxDelay)
	result := make([][]T, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}
