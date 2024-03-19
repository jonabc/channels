package channels

import (
	"time"

	internalTime "github.com/jonabc/channels/internal/time"
)

// BatchConfig contains user configurable options for the Batch functions
type BatchConfig struct {
	panc     chan<- any
	capacity int
}

func defaultBatchOptions[T any](inc <-chan T, batchSize int) []Option[BatchConfig] {
	return []Option[BatchConfig]{
		ChannelCapacityOption[BatchConfig](cap(inc) / batchSize),
	}
}

// Batch N values from the input channel into an array of N values in the output channel.
// The output channel will have capacity $`cap(input channel) / batchSize`$.
// The output channel is closed once the input channel is closed and a partial batch is
// sent to the output channel, if a partial batch exists.
func Batch[T any](inc <-chan T, batchSize int, maxDelay time.Duration, opts ...Option[BatchConfig]) <-chan []T {
	cfg := parseOpts(append(defaultBatchOptions(inc, batchSize), opts...)...)

	outc := make(chan []T, cfg.capacity)
	panc := cfg.panc
	buffer := make([]T, 0, batchSize)

	timer := internalTime.NewTimer(maxDelay)
	timer.Stop()

	publishAndReset := func() {
		timer.Stop()
		if len(buffer) == 0 {
			return
		}

		keys := make([]T, len(buffer))
		copy(keys, buffer)
		outc <- keys
		buffer = buffer[:0]
	}

	go func() {
		defer handlePanicIfErrc(panc)
		defer close(outc)
		defer timer.Stop()

		for {
			select {
			case in, ok := <-inc:
				if !ok {
					publishAndReset()
					return
				}

				if len(buffer) == 0 {
					timer.Reset(maxDelay)
				}

				buffer = append(buffer, in)
				if len(buffer) == cap(buffer) {
					publishAndReset()
				}
			case <-timer.C:
				publishAndReset()
			}
		}
	}()

	return outc
}

// Like Batch, but blocks until the input channel is closed and all values are read.
// BatchValue reads all values from the input channel and returns an array of batches.
func BatchValues[T any](inc <-chan T, batchSize int, maxDelay time.Duration, opts ...Option[BatchConfig]) [][]T {
	outc := Batch(inc, batchSize, maxDelay, opts...)
	result := make([][]T, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}
