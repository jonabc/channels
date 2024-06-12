package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
)

func TestUniqueAfterMaxBatchSize(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)

	batchSize := 5
	out := channels.Unique(in, batchSize, 0)

	require.Equal(t, 0, cap(out))

	go func() {
		defer close(in)
		for i := 1; i <= 2*cap(in); i++ {
			in <- i

			// Avoid sending a duplicate of the last item in a batch,
			// the first write of the item will trigger sending a batch
			// and the duplicate will end up starting a new batch.  This
			// makes validation harder for minimal added value, uniqueness
			// is still validated for all other items in each batch
			if i%batchSize != 0 {
				in <- i
			}
		}
	}()

	i := 0
	for batch := range out {
		expected := make([]int, 0, batchSize)
		for j := i * batchSize; j < (i+1)*batchSize; j++ {
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

func TestUniqueValuesAfterMaxBatchSize(t *testing.T) {
	t.Parallel()

	batchSize := 5
	in := make(chan int, 10)

	go func() {
		defer close(in)
		for i := 1; i <= cap(in); i++ {
			in <- i

			// Avoid sending a duplicate of the last item in a batch,
			// the first write of the item will trigger sending a batch
			// and the duplicate will end up starting a new batch.  This
			// makes validation harder for minimal added value, uniqueness
			// is still validated for all other items in each batch
			if i%batchSize != 0 {
				in <- i
			}
		}
	}()

	out := channels.UniqueValues(in, batchSize, 0)
	require.Equal(t, cap(in)/batchSize, len(out))

	for i, batch := range out {
		expected := make([]int, 0, batchSize)
		for j := i * batchSize; j < i*batchSize+batchSize; j++ {
			expected = append(expected, j+1)
		}
		require.ElementsMatch(t, batch, expected)
	}
}

func TestUniqueAfterMaxDelay(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	batchSize := 2
	maxDelay := 5 * time.Millisecond
	out := channels.Unique(in, batchSize, maxDelay)

	in <- 1
	in <- 1

	require.Equal(t, <-out, []int{1})
}

func TestUniqueValuesAfterMaxDelay(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 2
	maxDelay := 5 * time.Millisecond
	go func() {
		in <- 1
		in <- 1
		time.Sleep(maxDelay + 2*time.Millisecond)
		in <- 2
		in <- 2
		close(in)
	}()

	out := channels.UniqueValues(in, batchSize, maxDelay)

	require.Equal(t, [][]int{{1}, {2}}, out)
}

func TestUniqueDrainsItemsOnInputChannelClose(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 2
	out := channels.Unique(in, batchSize, 0)

	in <- 1
	in <- 1

	close(in)

	require.Equal(t, <-out, []int{1})
}

func TestUniqueDoesNotDrainEmptyBatchOnChannelClose(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	batchSize := 2
	out := channels.Unique(in, batchSize, 0)

	close(in)
	time.Sleep(1 * time.Millisecond)

	require.Len(t, out, 0)
}

func TestUniqueChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)
	outCapacity := 1

	out := channels.Unique(in, 5, 0,
		channels.ChannelCapacityOption[channels.BatchConfig](outCapacity),
	)

	require.Equal(t, outCapacity, cap(out))
}

func TestUniqueProviderOptionWithStatsReporting(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.BatchStats](0)
	defer provider.Close()

	out := channels.Unique(in, 2, 2*time.Millisecond,
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
	require.Equal(t, 0, stats[0].QueueLength)
}
