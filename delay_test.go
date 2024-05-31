package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestDelay(t *testing.T) {
	t.Parallel()

	inc := make(chan int, 10)
	defer close(inc)

	delay := 5 * time.Millisecond
	outc, getDelayedCount := channels.Delay(inc, delay)
	require.Equal(t, 0, cap(outc))

	start := time.Now()

	inc <- 1

	// sleep to allow delay subroutine to intake the value
	time.Sleep(1 * time.Millisecond)
	require.Equal(t, getDelayedCount(), 1)

	val, ok := <-outc
	require.True(t, ok)
	require.Equal(t, 1, val)
	require.GreaterOrEqual(t, time.Since(start), delay)
}

func TestDelayAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	panicProvider, _ := providers.NewProvider[any](0)
	defer panicProvider.Close()

	out, _ := channels.Delay(in, 2*time.Millisecond,
		channels.ChannelCapacityOption[channels.DelayConfig](5),
		channels.PanicProviderOption[channels.DelayConfig](panicProvider),
	)

	require.Equal(t, 5, cap(out))
}

func TestDelayProviderOptionWithStatsReporting(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.Stats](0)
	defer provider.Close()

	delay := 5 * time.Millisecond
	out, _ := channels.Delay(in, delay,
		channels.StatsProviderOption[channels.DelayConfig](provider),
	)

	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Duration, delay)
}
