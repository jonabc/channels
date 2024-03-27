package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestSplit(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)

	outArray := channels.Split(in, 2, func(i int, chans []chan<- int) {
		require.Len(t, chans, 2)
		chans[i%2] <- i
	})
	require.Equal(t, 2, len(outArray))

	evens := outArray[0]
	require.Equal(t, 1, cap(evens))

	odds := outArray[1]
	require.Equal(t, 1, cap(odds))

	in <- 1
	in <- 2
	in <- 3
	in <- 4
	close(in)

	require.Equal(t, <-odds, 1)
	require.Equal(t, <-evens, 2)
	require.Equal(t, <-odds, 3)
	require.Equal(t, <-evens, 4)

	_, ok := <-odds
	require.False(t, ok)
	_, ok = <-odds
	require.False(t, ok)
}

func TestSplitValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	in <- 1
	in <- 2
	in <- 3
	in <- 4
	close(in)

	results := channels.SplitValues(in, 2, func(i int, chans []chan<- int) {
		require.Len(t, chans, 2)
		chans[i%2] <- i
	})
	require.Equal(t, 2, len(results))

	evens := results[0]
	require.ElementsMatch(t, evens, []int{2, 4})

	odds := results[1]
	require.ElementsMatch(t, odds, []int{1, 3})
}

func TestSplit2(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	evens, odds := channels.Split2(in, func(i int, chans []chan<- int) {
		require.Len(t, chans, 2)
		chans[i%2] <- i
	})

	in <- 1
	require.Equal(t, 1, <-odds)
	in <- 2
	require.Equal(t, 2, <-evens)
	in <- 3
	require.Equal(t, 3, <-odds)
	in <- 4
	require.Equal(t, 4, <-evens)
}

func TestSplit3(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	zeros, ones, twos := channels.Split3(in, func(i int, chans []chan<- int) {
		require.Len(t, chans, 3)
		chans[i%3] <- i
	})

	in <- 1
	require.Equal(t, 1, <-ones)
	in <- 2
	require.Equal(t, 2, <-twos)
	in <- 3
	require.Equal(t, 3, <-zeros)
	in <- 4
	require.Equal(t, 4, <-ones)
}

func TestSplitMultiChannelCapacitiesOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	outArray := channels.Split(in,
		2,
		func(i int, chans []chan<- int) { chans[i%2] <- i },
		channels.MultiChannelCapacitiesOption[channels.SplitConfig]([]int{2, 5}),
	)

	require.Equal(t, 2, cap(outArray[0]))
	require.Equal(t, 5, cap(outArray[1]))
}

func TestSplitProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Split(in,
		2,
		func(i int, c []chan<- int) { panic("panic!") },
		channels.PanicProviderOption[channels.SplitConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}

func TestSplitProviderOptionWithReportStats(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.Stats](0)
	defer provider.Close()

	out := channels.Split(in,
		2,
		func(i int, chans []chan<- int) {
			time.Sleep(2 * time.Millisecond)
			chans[i%2] <- i
		},
		channels.StatsProviderOption[channels.SplitConfig](provider),
	)

	in <- 1
	<-out[1]

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, 2*time.Millisecond)
}
