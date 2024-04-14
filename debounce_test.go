package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestDebounce(t *testing.T) {
	t.Parallel()

	inc := make(chan int, 100)

	delay := 20 * time.Millisecond
	outc, getDebouncedCount := channels.Debounce(inc, delay)
	require.Equal(t, 0, cap(outc))

	start := time.Now()

	inc <- 1
	inc <- 1

	time.Sleep(2 * time.Millisecond)

	inc <- 2
	inc <- 1

	time.Sleep(2 * time.Millisecond)
	require.Equal(t, 1, getDebouncedCount())

	// still waiting for debounce, out channel should be empty
	require.Len(t, outc, 0)

	require.Equal(t, 1, <-outc)
	require.GreaterOrEqual(t, time.Since(start), delay)

	// sleep for an additional delay period and ensure that no other values
	// have been written to outc
	time.Sleep(delay)
	require.Len(t, outc, 0)

	close(inc)
}

func TestDebounceAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	panicProvider, _ := providers.NewProvider[any](0)
	defer panicProvider.Close()

	out, _ := channels.Debounce(in, 2*time.Millisecond,
		channels.ChannelCapacityOption[channels.DebounceConfig](5),
		channels.PanicProviderOption[channels.DebounceConfig](panicProvider),
	)

	require.Equal(t, 5, cap(out))
}

func TestDebounceProviderOptionWithStatsReporting(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.DebounceStats](0)
	defer provider.Close()

	delay := 5 * time.Millisecond
	out, _ := channels.Debounce(in, delay,
		channels.DebounceStatsProviderOption(provider),
	)

	in <- 1
	in <- 2
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.Greater(t, stats[0].Delay, 5*time.Millisecond)
	require.Equal(t, uint(2), stats[0].Count)
}

func TestDebounceLeadDebounceTypeOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getDebouncedCount := channels.Debounce(in, delay,
		channels.DebounceTypeOption(channels.LeadDebounceType),
	)

	in <- 1
	in <- 2
	start := time.Now()
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 1, getDebouncedCount())

	in <- 1
	select {
	case <-out:
		require.FailNow(t, "Unexpected value during debounce period")
	case <-time.After(delay + 2*time.Millisecond):
	}
	require.Equal(t, 0, getDebouncedCount())

	in <- 1
	in <- 2
	start = time.Now()
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 1, getDebouncedCount())
}

func TestDebounceLeadTailDebounceTypeOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getDebouncedCount := channels.Debounce(in, delay,
		channels.DebounceTypeOption(channels.LeadTailDebounceType),
	)

	in <- 1
	in <- 2
	start := time.Now()
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 1, getDebouncedCount())

	<-out
	require.Greater(t, time.Since(start), delay)
	require.Equal(t, 0, getDebouncedCount())
}
