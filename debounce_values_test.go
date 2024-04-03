package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestDebounceValues(t *testing.T) {
	t.Parallel()

	inc := make(chan int, 100)

	delay := 20 * time.Millisecond
	outc, getDebouncedCount := channels.DebounceValues(inc, delay)
	require.Equal(t, 0, cap(outc))

	start := time.Now()

	inc <- 1
	inc <- 1

	time.Sleep(2 * time.Millisecond)

	inc <- 2
	inc <- 1

	time.Sleep(2 * time.Millisecond)
	require.Equal(t, 2, getDebouncedCount())

	// still waiting for debounce, out channel should be empty
	require.Len(t, outc, 0)

	results := make([]int, 0)
	results = append(results, <-outc)

	require.ElementsMatch(t, results, []int{1})
	require.GreaterOrEqual(t, time.Since(start), delay)

	results = append(results, <-outc)
	require.ElementsMatch(t, results, []int{1, 2})
	require.GreaterOrEqual(t, time.Since(start), delay+(2*time.Millisecond))

	inc <- 1
	inc <- 2
	close(inc)

	results = results[:0]
	for result := range outc {
		results = append(results, result)
	}

	require.ElementsMatch(t, results, []int{1, 2})
}

func TestDebounceValuesAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	panicProvider, _ := providers.NewProvider[any](0)
	defer panicProvider.Close()

	out, _ := channels.DebounceValues(in, 2*time.Millisecond,
		channels.ChannelCapacityOption[channels.DebounceConfig](5),
		channels.PanicProviderOption[channels.DebounceConfig](panicProvider),
	)

	require.Equal(t, 5, cap(out))
}

func TestDebounceValuesProviderOptionWithStatsReporting(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewCollectingProvider[channels.DebounceStats](0)
	defer provider.Close()

	delay := 5 * time.Millisecond
	out, _ := channels.DebounceValues(in, delay,
		channels.DebounceStatsProviderOption(provider),
	)

	in <- 1
	in <- 1
	<-out

	stats, ok := <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.Greater(t, stats[0].Delay, 5*time.Millisecond)
	require.Equal(t, uint(2), stats[0].Count)

	in <- 1
	<-out

	stats, ok = <-receiver.Channel()
	require.True(t, ok)
	require.Len(t, stats, 1)
	require.GreaterOrEqual(t, stats[0].Delay, 2*time.Millisecond)
	require.Equal(t, uint(1), stats[0].Count)
}

func TestDebounceValuesLeadDebounceTypeOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getDebouncedCount := channels.DebounceValues(in, delay,
		channels.DebounceTypeOption(channels.LeadDebounceType),
	)

	in <- 1
	in <- 2
	start := time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getDebouncedCount())

	in <- 1
	in <- 2
	select {
	case <-out:
		require.FailNow(t, "Unexpected value during debounce period")
	case <-time.After(delay + 1*time.Millisecond):
	}
	require.Equal(t, 0, getDebouncedCount())

	in <- 1
	in <- 2
	start = time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getDebouncedCount())
}

func TestDebounceValuesLeadTailDebounceTypeOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getDebouncedCount := channels.DebounceValues(in, delay,
		channels.DebounceTypeOption(channels.LeadTailDebounceType),
	)

	in <- 1
	in <- 2
	start := time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getDebouncedCount())

	<-out
	<-out
	require.Greater(t, time.Since(start), delay)
	require.Equal(t, 0, getDebouncedCount())
}
