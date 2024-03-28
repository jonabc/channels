package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
)

func TestReject(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	out := channels.Reject(in, func(i int) bool { return i%2 == 0 })
	require.Equal(t, 0, cap(out))

	in <- 2
	in <- 1
	close(in)

	require.Equal(t, <-out, 1)
	_, ok := <-out
	require.False(t, ok)
}

func TestRejectValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	close(in)

	out := channels.RejectValues(in, func(i int) bool { return i%2 == 0 })

	require.Len(t, out, 1)
	require.Equal(t, out, []int{1})
}

func TestRejectChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Reject(in,
		func(i int) bool { return i%2 == 0 },
		channels.ChannelCapacityOption[channels.SelectConfig](5),
	)

	require.Equal(t, 5, cap(out))
}

func TestRejectProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Reject(in,
		func(i int) bool { panic("panic!") },
		channels.PanicProviderOption[channels.SelectConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}

func TestRejectProviderOptionWithReportStats(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.SelectStats](0)
	defer provider.Close()

	out := channels.Reject(in,
		func(i int) bool {
			time.Sleep(2 * time.Millisecond)
			return i%2 == 0
		},
		channels.SelectStatsProviderOption(provider),
	)

	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, 2*time.Millisecond)
	require.True(t, stats[0].Selected)
}
