package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
)

func TestBatchAfterMaxBatchSize(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 5
	out := channels.Batch(in, batchSize, 0)

	maxBatches := cap(in) / batchSize
	require.Equal(t, maxBatches, cap(out))

	go func() {
		defer close(in)
		for i := 1; i <= 2*cap(in); i++ {
			in <- i
		}
	}()

	i := 0
	for batch := range out {
		expected := make([]int, 0, batchSize)
		for j := i * batchSize; j < i*batchSize+batchSize; j++ {
			expected = append(expected, j+1)
		}
		require.ElementsMatch(t, batch, expected)
		i++
	}

	require.Len(t, in, 0)
	require.Len(t, out, 0)
	_, ok := <-out
	require.False(t, ok)
}

func TestBatchAfterMaxDelay(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

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
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 2
	out := channels.Batch(in, batchSize, 0)

	close(in)
	time.Sleep(1 * time.Millisecond)

	require.Len(t, out, 0)
}
