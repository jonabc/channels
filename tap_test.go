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
