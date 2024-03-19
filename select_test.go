package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
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

func TestSelectAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	errs := make(chan any)
	defer close(errs)

	out := channels.Select(in,
		func(i int) bool { panic("panic!") },
		channels.ChannelCapacityOption[channels.SelectConfig](5),
		channels.ErrorChannelOption[channels.SelectConfig](errs),
	)

	require.Equal(t, 5, cap(out))

	in <- 1
	require.Equal(t, "panic!", <-errs)
}
