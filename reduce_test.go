package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	out := channels.Reduce(in, func(current int, i int) int { return current + i })
	require.Equal(t, cap(in), cap(out))

	current := 0
	for i := 0; i < 10; i++ {
		in <- i

		time.Sleep(1 * time.Millisecond)
		next, ok := <-out
		require.True(t, ok)

		current = current + i
		require.Equal(t, current, next)

	}
}

func TestReduceValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	in <- 3
	close(in)

	out := channels.ReduceValues(in, func(current int, i int) int { return current + i })
	require.Equal(t, out, 6)
}
