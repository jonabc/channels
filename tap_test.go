package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestTap(t *testing.T) {
	in := make(chan int, 10)
	defer close(in)

	pre := make([]int, 0)
	post := make([]int, 0)
	out := channels.Tap(in,
		func(i int) { pre = append(pre, i) },
		func(i int) { post = append(post, i) },
	)

	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2

	time.Sleep(1 * time.Millisecond)

	require.Equal(t, 1, <-out)
	require.Equal(t, 2, <-out)
	require.Equal(t, []int{1, 2}, pre)
	require.Equal(t, []int{1, 2}, post)
	require.Len(t, out, 0)
}
