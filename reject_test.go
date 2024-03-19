package channels_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
)

func TestReject(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Reject(in, func(i int) bool { return i%2 == 0 })
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2

	time.Sleep(1 * time.Millisecond)

	require.Len(t, out, 1)
	require.Equal(t, <-out, 1)
}

func TestRejectValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	close(in)

	out := channels.RejectValues(in, func(i int) bool { return i%2 == 0 })

	require.Len(t, out, 1)
	require.Equal(t, out, []int{1})
}

func TestRejectAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	errs := make(chan any)
	defer close(errs)

	out := channels.Reject(in,
		func(i int) bool { panic("panic!") },
		channels.ChannelCapacityOption[channels.RejectConfig](5),
		channels.ErrorChannelOption[channels.RejectConfig](errs),
	)

	require.Equal(t, 5, cap(out))

	in <- 1
	require.Equal(t, "panic!", <-errs)
}
