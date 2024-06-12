package channels

import (
	"time"

	internalTime "github.com/jonabc/channels/internal/time"
	"golang.org/x/exp/maps"
)

type keyedWrapper[T comparable] struct {
	val T
}

func (w *keyedWrapper[T]) Key() T {
	return w.val
}

// Batch unique values from the input channel into an array of values written to the output channel.
// The output channel is unbuffered by default, and will be closed when the input channel
// is closed and drained.  If a partial batch exists when the input channel is closed,
// the partial batch will be sent to the output channel.
func Unique[T comparable](inc <-chan T, batchSize int, maxDelay time.Duration, opts ...Option[BatchConfig]) <-chan []T {
	cfg := parseOpts(opts...)

	outc := make(chan []T, cfg.capacity)

	inBridge := make(chan *keyedWrapper[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &keyedWrapper[T]{val: in}
		}
	}()

	outBridge := UniqueKeyed(inBridge, batchSize, maxDelay,
		append(opts, ChannelCapacityOption[BatchConfig](0))...,
	)
	go func() {
		defer close(outc)
		for out := range outBridge {
			vals := make([]T, len(out))
			for i, wrapper := range out {
				vals[i] = wrapper.val
			}

			outc <- vals
		}
	}()

	return outc
}

// Batch unique values from the input channel into an array of values written to the output channel.  Uniqueness is determed
// by the value returned by each value's Key() function.
// The output channel is unbuffered by default, and will be closed when the input channel
// is closed and drained.  If a partial batch exists when the input channel is closed,
// the partial batch will be sent to the output channel.
func UniqueKeyed[K comparable, V Keyable[K]](inc <-chan V, batchSize int, maxDelay time.Duration, opts ...Option[BatchConfig]) <-chan []V {
	cfg := parseOpts(opts...)

	outc := make(chan []V, cfg.capacity)
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	buffer := make(map[K]V, batchSize)

	timer := internalTime.NewTimer(maxDelay)
	timer.Stop()

	var batchStart time.Time

	publishAndReset := func() {
		timer.Stop()
		if len(buffer) == 0 {
			return
		}

		duration := time.Since(batchStart)
		batchSize := len(buffer)

		keys := maps.Values(buffer)
		outc <- keys
		clear(buffer)
		tryProvideStats(BatchStats{Duration: duration, BatchSize: uint(batchSize), QueueLength: len(inc)}, statsProvider)
	}

	go func() {
		defer tryHandlePanic(panicProvider)
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
					batchStart = time.Now()
				}

				buffer[in.Key()] = in
				if len(buffer) == batchSize {
					publishAndReset()
				}
			case <-timer.C:
				publishAndReset()
			}
		}
	}()

	return outc
}

// Like Unique, but blocks until the input channel is closed and all values are read.
// UniqueValues reads all values from the input channel and returns an array of batches.
func UniqueValues[T comparable](inc <-chan T, batchSize int, maxDelay time.Duration, opts ...Option[BatchConfig]) [][]T {
	outc := Unique(inc, batchSize, maxDelay, opts...)
	result := make([][]T, 0, len(inc))

	for out := range outc {
		result = append(result, out)
	}

	return result
}
