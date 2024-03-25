package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestFlatMap(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.FlatMap(in, func(i int) ([]int, bool) {
		return []int{i * 10, i*10 + 1}, i < 3
	})
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2
	in <- 3

	require.Equal(t, <-out, 10)
	require.Equal(t, <-out, 11)
	require.Equal(t, <-out, 20)
	require.Equal(t, <-out, 21)
	require.Len(t, out, 0)
}

func TestFlatMapValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := channels.FlatMapValues(in, func(i int) ([]int, bool) {
		return []int{i * 10, i*10 + 1}, i < 3
	})

	require.Len(t, out, 4)
	require.Equal(t, out, []int{10, 11, 20, 21})
}

func TestFlatMapChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.FlatMap(in,
		func(i int) ([]int, bool) { return []int{i * 10, i*10 + 1}, i < 3 },
		channels.ChannelCapacityOption[channels.FlatMapConfig](5),
	)

	require.Equal(t, 5, cap(out))
}

func TestFlatMapProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.FlatMap(in,
		func(i int) ([]bool, bool) { panic("panic!") },
		channels.PanicProviderOption[channels.FlatMapConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}

func TestFlatMapProviderOptionWithReportStats(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.Stats](0)
	defer provider.Close()

	out := channels.FlatMap(in,
		func(i int) ([]bool, bool) {
			time.Sleep(2 * time.Millisecond)
			return []bool{true}, true
		},
		channels.StatsProviderOption[channels.FlatMapConfig](provider),
	)

	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, 2*time.Millisecond)
}
