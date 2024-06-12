package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
)

func TestBatchAfterMaxBatchSize(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 5
	out := channels.Batch(in, batchSize, 0)

	require.Equal(t, 0, cap(out))

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

func TestBatchValuesAfterMaxBatchSize(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 5
	for i := 1; i <= cap(in); i++ {
		in <- i
	}
	close(in)

	out := channels.BatchValues(in, batchSize, 0)
	require.Equal(t, cap(in)/batchSize, len(out))

	for i, batch := range out {
		expected := make([]int, 0, batchSize)
		for j := i * batchSize; j < i*batchSize+batchSize; j++ {
			expected = append(expected, j+1)
		}
		require.ElementsMatch(t, batch, expected)
	}
}

func TestBatchAfterMaxDelay(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	batchSize := 2
	maxDelay := 5 * time.Millisecond
	out := channels.Batch(in, batchSize, maxDelay)

	in <- 1

	require.Equal(t, <-out, []int{1})
}

func TestBatchValuesAfterMaxDelay(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 2
	maxDelay := 5 * time.Millisecond
	go func() {
		in <- 1
		time.Sleep(maxDelay + 2*time.Millisecond)
		in <- 2
		close(in)
	}()

	out := channels.BatchValues(in, batchSize, maxDelay)

	require.Equal(t, [][]int{{1}, {2}}, out)
}

func TestBatchDrainsItemsOnInputChannelClose(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 2
	out := channels.Batch(in, batchSize, 0)

	in <- 1

	close(in)

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

func TestBatchChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)
	outCapacity := 1

	out := channels.Batch(in, 5, 0,
		channels.ChannelCapacityOption[channels.BatchConfig](outCapacity),
	)

	require.Equal(t, outCapacity, cap(out))
}

func TestBatchProviderOptionWithStatsReporting(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.BatchStats](0)
	defer provider.Close()

	out := channels.Batch(in, 2, 2*time.Millisecond,
		channels.BatchStatsProviderOption(provider),
	)

	in <- 1
	in <- 2
	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, time.Duration(0))
	require.Equal(t, uint(2), stats[0].BatchSize)
	require.Equal(t, 1, stats[0].QueueLength)

	<-out

	stats, ok = <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, 2*time.Millisecond)
	require.Equal(t, uint(1), stats[0].BatchSize)
	require.Equal(t, 0, stats[0].QueueLength)
}
