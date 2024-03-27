package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestTap(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)

	pre := make([]int, 0)
	post := make([]int, 0)
	out := channels.Tap(in,
		func(i int) { pre = append(pre, i) },
		func(i int) { post = append(post, i) },
	)

	require.Equal(t, 0, cap(out))

	in <- 1
	in <- 2
	close(in)

	results := []int{}
	for result := range out {
		results = append(results, result)
	}
	require.Equal(t, []int{1, 2}, results)
	require.Equal(t, []int{1, 2}, pre)
	require.Equal(t, []int{1, 2}, post)

	_, ok := <-out
	require.False(t, ok)
}

func TestTapAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	out := channels.Tap(in,
		nil,
		nil,
		channels.ChannelCapacityOption[channels.TapConfig](1),
	)

	require.Equal(t, 1, cap(out))
}

func TestTapProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Tap(in,
		func(i int) { panic("panic!") },
		nil,
		channels.PanicProviderOption[channels.TapConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}

func TestTapProviderOptionWithReportStats(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.TapStats](0)
	defer provider.Close()

	out := channels.Tap(in,
		func(i int) { time.Sleep(2 * time.Millisecond) },
		func(i int) { time.Sleep(4 * time.Millisecond) },
		channels.TapStatsProviderOption(provider),
	)

	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].PreDuration, 2*time.Millisecond)
	require.GreaterOrEqual(t, stats[0].PostDuration, 4*time.Millisecond)
}
