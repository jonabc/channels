package channels_test

import (
	"testing"

	"github.com/jonabc/channels"
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

func TestReduceAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	errs := make(chan any)
	defer close(errs)

	out := channels.Reduce(in,
		func(i int, j int) (int, bool) { panic("panic!") },
		channels.ChannelCapacityOption[channels.ReduceConfig](5),
		channels.ErrorChannelOption[channels.ReduceConfig](errs),
	)

	require.Equal(t, 5, cap(out))

	in <- 1
	require.Equal(t, "panic!", <-errs)
}
