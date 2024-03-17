package channels_test

import (
	"context"
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestReduce(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	ctx := context.Background()
	out := channels.Reduce(ctx, in, func(fnCtx context.Context, current int, i int) (int, bool) {
		require.Equal(t, ctx, fnCtx)
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

	ctx := context.Background()
	out := channels.ReduceValues(ctx, in, func(fnCtx context.Context, current int, i int) (int, bool) {
		require.Equal(t, ctx, fnCtx)
		if i == 3 {
			return 10, false
		}
		return current + i, true
	})
	require.Equal(t, out, 7)
}
