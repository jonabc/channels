package channels_test

import (
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
)

func TestBatchAfterMaxBatchSize(t *testing.T) {
	in := make(chan int, 100)
	defer close(in)

	batchSize := 5
	out := channels.Batch(in, batchSize, 0)

	maxBatches := cap(in) / batchSize
	require.Equal(t, maxBatches, cap(out))

	batches := 0
	for i := 1; i <= 2*cap(in); i++ {
		in <- i
		if i%batchSize == 0 {
			batches++
		}

		time.Sleep(1 * time.Millisecond)
		require.Len(t, out, int(math.Min(float64(batches), float64(cap(out)))), out)
	}

	// After sending all items to  the in channel, 105 items should have been read from it:
	// (maxBatches * batchSize) in the out channel + batchSize in the internal
	// batch buffer waiting to be sent.
	require.Len(t, in, cap(in)-batchSize)

	// The out channel should be at capacity
	require.Len(t, out, cap(out))

	// All batches should be sent into the in channel
	require.Equal(t, 2*cap(in)/batchSize, batches)

	for i := 0; i < batches; i++ {
		time.Sleep(1 * time.Millisecond)
		batch := <-out
		expected := make([]int, 0, batchSize)
		for j := i * batchSize; j < i*batchSize+batchSize; j++ {
			expected = append(expected, j+1)
		}
		require.ElementsMatch(t, batch, expected)
	}

	require.Len(t, in, 0)
	require.Len(t, out, 0)
}

func TestBatchAfterMaxDelay(t *testing.T) {
	in := make(chan int, 100)
	defer close(in)

	batchSize := 2
	maxDelay := 5 * time.Millisecond
	out := channels.Batch(in, batchSize, maxDelay)

	in <- 1
	time.Sleep(2 * maxDelay)

	require.Len(t, out, 1)
	require.Equal(t, <-out, []int{1})
}

func TestBatchDrainsItemsOnInputChannelClose(t *testing.T) {
	in := make(chan int, 100)

	batchSize := 2
	out := channels.Batch(in, batchSize, 0)

	in <- 1

	close(in)
	time.Sleep(1 * time.Millisecond)

	require.Len(t, out, 1)
	require.Equal(t, <-out, []int{1})
}

func TestDoesNotDrainEmptyBatchOnChannelClose(t *testing.T) {
	in := make(chan int, 100)

	batchSize := 2
	out := channels.Batch(in, batchSize, 0)

	close(in)
	time.Sleep(1 * time.Millisecond)

	require.Len(t, out, 0)
}
