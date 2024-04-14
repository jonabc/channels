package channels_test

import (
	"sync"
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

type customDebouncingType struct {
	key   string
	value string
	delay time.Duration
}

func (d *customDebouncingType) Key() string {
	return d.key
}

func (d *customDebouncingType) Delay() time.Duration {
	return d.delay
}

func (d *customDebouncingType) Reduce(other *customDebouncingType) (*customDebouncingType, bool) {
	if d.key == "3" {
		return d, false
	}

	d.value += "," + other.value
	return d, true
}

func TestDebounceCustom(t *testing.T) {
	t.Parallel()

	inc := make(chan *customDebouncingType, 100)
	defer close(inc)

	delay := 5 * time.Millisecond
	outc, getDebouncedCount := channels.DebounceCustom(inc)
	require.Equal(t, 0, cap(outc))

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		inc <- &customDebouncingType{key: "1", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "1", value: "val2", delay: 1 * time.Millisecond}
		time.Sleep(2 * time.Millisecond)

		// should have one debounced item after pushing "key:1" twice
		require.Equal(t, 1, getDebouncedCount())
		inc <- &customDebouncingType{key: "2", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "2", value: "val2", delay: 1 * time.Millisecond}
		inc <- &customDebouncingType{key: "1", value: "val3", delay: 2 * time.Millisecond}

		time.Sleep(2 * time.Millisecond)
		inc <- &customDebouncingType{key: "3", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "3", value: "shouldnotreduce", delay: 1 * time.Millisecond}
		wg.Wait()
	}()

	require.Len(t, outc, 0)

	// only the delay for the first element seen on any key counts
	require.Equal(t, &customDebouncingType{key: "1", value: "val1,val2,val3", delay: delay}, <-outc)

	require.GreaterOrEqual(t, time.Since(start), delay)

	require.Equal(t, 2, getDebouncedCount())
	require.Equal(t, &customDebouncingType{key: "2", value: "val1,val2", delay: delay}, <-outc)
	require.GreaterOrEqual(t, time.Since(start), delay+(2*time.Millisecond))

	require.Equal(t, 1, getDebouncedCount())
	require.Equal(t, &customDebouncingType{key: "3", value: "val1", delay: delay}, <-outc)
	require.GreaterOrEqual(t, time.Since(start), delay+(4*time.Millisecond))
}

func TestDebounceCustomAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan *customDebouncingType, 100)
	defer close(in)

	panicProvider, _ := providers.NewProvider[any](0)
	defer panicProvider.Close()

	out, _ := channels.DebounceCustom(in,
		channels.ChannelCapacityOption[channels.DebounceConfig](5),
		channels.PanicProviderOption[channels.DebounceConfig](panicProvider),
	)

	require.Equal(t, 5, cap(out))
}

func TestDebounceCustomLeadDebounceTypeOption(t *testing.T) {
	t.Parallel()

	in := make(chan *customDebouncingType, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getDebouncedCount := channels.DebounceCustom(in,
		channels.DebounceTypeOption(channels.LeadDebounceType),
	)

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
	start := time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getDebouncedCount())

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
	select {
	case <-out:
		require.FailNow(t, "Unexpected value during throttling period")
	case <-time.After(delay + 2*time.Millisecond):
	}
	require.Equal(t, 0, getDebouncedCount())

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
	start = time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getDebouncedCount())
}

func TestDebounceCustomLeadTailDebounceTypeOption(t *testing.T) {
	t.Parallel()

	in := make(chan *customDebouncingType, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getDebouncedCount := channels.DebounceCustom(in,
		channels.DebounceTypeOption(channels.LeadTailDebounceType),
	)

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
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
