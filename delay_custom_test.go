package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestDelayCustom(t *testing.T) {
	t.Parallel()

	inc := make(chan *customDebouncingType, 100)
	defer close(inc)

	delay := 5 * time.Millisecond
	outc, getDelayedCount := channels.DelayCustom(inc)
	require.Equal(t, 0, cap(outc))

	start := time.Now()

	input := &customDebouncingType{key: "1", value: "val1", delay: delay}
	inc <- input

	// sleep to allow delay subroutine to intake the value
	time.Sleep(1 * time.Millisecond)
	require.Equal(t, getDelayedCount(), 1)

	val, ok := <-outc
	require.True(t, ok)
	require.Equal(t, input, val)
	require.GreaterOrEqual(t, time.Since(start), delay)
}

func TestDelayCustomAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan *customDebouncingType, 100)
	defer close(in)

	panicProvider, _ := providers.NewProvider[any](0)
	defer panicProvider.Close()

	out, _ := channels.DelayCustom(in,
		channels.ChannelCapacityOption[channels.DelayConfig](5),
		channels.PanicProviderOption[channels.DelayConfig](panicProvider),
	)

	require.Equal(t, 5, cap(out))
}
