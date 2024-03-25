package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Select(in, func(i int) bool { return i%2 == 0 })
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2

	time.Sleep(1 * time.Millisecond)

	require.Len(t, out, 1)
	require.Equal(t, <-out, 2)
}

func TestSelectValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	close(in)

	out := channels.SelectValues(in, func(i int) bool { return i%2 == 0 })

	require.Len(t, out, 1)
	require.Equal(t, out, []int{2})
}

func TestSelectChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Select(in,
		func(i int) bool { return i%2 == 0 },
		channels.ChannelCapacityOption[channels.SelectConfig](5),
	)

	require.Equal(t, 5, cap(out))
}

func TestSelectProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Select(in,
		func(i int) bool { panic("panic!") },
		channels.PanicProviderOption[channels.SelectConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}
