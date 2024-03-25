package channels_test

import (
	"testing"

	"github.com/jonabc/channels"
	"github.com/jonabc/channels/providers"
	"github.com/stretchr/testify/require"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Reduce(in, func(current int, i int) (int, bool) {
		if i == 3 {
			return 10, false
		}
		return current + i, true
	})
	require.Equal(t, cap(in), cap(out))
	in <- 1
	in <- 2
	in <- 3
	in <- 4

	require.Equal(t, 1, <-out)
	require.Equal(t, 3, <-out)
	require.Equal(t, 7, <-out)

	require.Equal(t, 0, len(out))
}

func TestReduceValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	in <- 3
	in <- 4
	close(in)

	out := channels.ReduceValues(in, func(current int, i int) (int, bool) {
		if i == 3 {
			return 10, false
		}
		return current + i, true
	})
	require.Equal(t, out, 7)
}

func TestReduceChannelCapacityOption(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Reduce(in,
		func(current int, i int) (int, bool) {
			if i == 3 {
				return 10, false
			}
			return current + i, true
		},
		channels.ChannelCapacityOption[channels.ReduceConfig](5),
	)

	require.Equal(t, 5, cap(out))
}

func TesReduceProviderOptionWithReportPanics(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	provider, receiver := providers.NewProvider[any](0)
	defer provider.Close()

	channels.Reduce(in,
		func(current int, i int) (int, bool) { panic("panic!") },
		channels.PanicProviderOption[channels.ReduceConfig](provider),
	)

	in <- 1
	require.Equal(t, "panic!", <-receiver.Channel())
}
