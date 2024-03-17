package channels_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/jonabc/channels"
)

func TestSelect(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	ctx := context.Background()
	out := channels.Select(ctx, in, func(fnCtx context.Context, i int) bool {
		require.Equal(t, ctx, fnCtx)
		return i%2 == 0
	})
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

	ctx := context.Background()
	out := channels.SelectValues(ctx, in, func(fnCtx context.Context, i int) bool {
		require.Equal(t, ctx, fnCtx)
		return i%2 == 0
	})

	require.Len(t, out, 1)
	require.Equal(t, out, []int{2})
}
