package channels_test

import (
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestFlatMap(t *testing.T) {
	in := make(chan int, 100)
	defer close(in)

	out := channels.FlatMap(in, func(i int) []int { return []int{i * 10, i*10 + 1} })
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2

	require.Equal(t, <-out, 10)
	require.Equal(t, <-out, 11)
	require.Equal(t, <-out, 20)
	require.Equal(t, <-out, 21)
	require.Len(t, out, 0)
}

func TestFlatMapValues(t *testing.T) {
	in := make(chan int, 100)

	in <- 1
	in <- 2
	close(in)

	out := channels.FlatMapValues(in, func(i int) []int { return []int{i * 10, i*10 + 1} })

	require.Len(t, out, 4)
	require.Equal(t, out, []int{10, 11, 20, 21})
}
