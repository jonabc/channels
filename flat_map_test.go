package channels_test

import (
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestFlatMap(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.FlatMap(in, func(i int) ([]int, bool) {
		return []int{i * 10, i*10 + 1}, i < 3
	})
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2
	in <- 3

	require.Equal(t, <-out, 10)
	require.Equal(t, <-out, 11)
	require.Equal(t, <-out, 20)
	require.Equal(t, <-out, 21)
	require.Len(t, out, 0)
}

func TestFlatMapValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := channels.FlatMapValues(in, func(i int) ([]int, bool) {
		return []int{i * 10, i*10 + 1}, i < 3
	})

	require.Len(t, out, 4)
	require.Equal(t, out, []int{10, 11, 20, 21})
}

func TestFlatMapAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	errs := make(chan any)
	defer close(errs)

	out := channels.FlatMap(in,
		func(i int) ([]bool, bool) { panic("panic!") },
		channels.ChannelCapacityOption[channels.FlatMapConfig](5),
		channels.ErrorChannelOption[channels.FlatMapConfig](errs),
	)

	require.Equal(t, 5, cap(out))

	in <- 1
	require.Equal(t, "panic!", <-errs)
}
