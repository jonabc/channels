package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestEach(t *testing.T) {
	t.Parallel()

	provider, receiver := providers.NewCollectingProvider[channels.Stats](0)
	defer provider.Close()

	in := make(chan int, 100)
	defer close(in)

	results := []int{}
	channels.Each(in,
		func(i int) { results = append(results, i) },
		channels.StatsProviderOption[channels.EachConfig](provider),
	)

	in <- 1
	in <- 2

	collectedStats := 0
	for collectedStats < 2 {
		collectedStats += len(<-receiver.Channel())
	}

	require.Equal(t, []int{1, 2}, results)
}

func TestEachProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Each(in,
		func(i int) { panic("panic!") },
		channels.PanicProviderOption[channels.EachConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}

func TestEachProviderOptionWithReportStats(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.Stats](0)
	defer provider.Close()

	channels.Each(in,
		func(i int) { time.Sleep(2 * time.Millisecond) },
		channels.StatsProviderOption[channels.EachConfig](provider),
	)

	in <- 1

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, 2*time.Millisecond)
}
