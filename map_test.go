package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Map(in, func(i int) (bool, bool) {
		return i%2 == 0, i < 3
	})
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2
	in <- 3

	require.Equal(t, <-out, false)
	require.Equal(t, <-out, true)
	require.Len(t, out, 0)
}

func TestMapValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := channels.MapValues(in, func(i int) (bool, bool) {
		return i%2 == 0, i < 3
	})

	require.Len(t, out, 2)
	require.Equal(t, out, []bool{false, true})
}

func TestMapChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Map(in,
		func(i int) (bool, bool) { return i%2 == 0, i < 3 },
		channels.ChannelCapacityOption[channels.MapConfig](5),
	)

	require.Equal(t, 5, cap(out))
}

func TestMapProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Map(in,
		func(i int) (bool, bool) { panic("panic!") },
		channels.PanicProviderOption[channels.MapConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}

func TestMapProviderOptionWithReportStats(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.Stats](0)
	defer provider.Close()

	out := channels.Map(in,
		func(i int) (bool, bool) {
			time.Sleep(2 * time.Millisecond)
			return true, true
		},
		channels.StatsProviderOption[channels.MapConfig](provider),
	)

	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, 2*time.Millisecond)
}
