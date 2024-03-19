package channels_test

import (
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestTap(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)

	pre := make([]int, 0)
	post := make([]int, 0)
	out := channels.Tap(in,
		func(i int) { pre = append(pre, i) },
		func(i int) { post = append(post, i) },
	)

	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2
	close(in)

	results := []int{}
	for result := range out {
		results = append(results, result)
	}
	require.Equal(t, []int{1, 2}, results)
	require.Equal(t, []int{1, 2}, pre)
	require.Equal(t, []int{1, 2}, post)
	require.Len(t, out, 0)
}

func TestTapAcceptsOptions(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	errs := make(chan any)
	defer close(errs)

	out := channels.Tap(in,
		func(i int) { panic("panic!") },
		nil,
		channels.ErrorChannelOption[channels.TapConfig](errs),
		channels.ChannelCapacityOption[channels.TapConfig](1),
	)

	require.Equal(t, 1, cap(out))

	in <- 1
	require.Equal(t, "panic!", <-errs)
}
