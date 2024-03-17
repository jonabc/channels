package channels_test

import (
	"context"
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	ctx := context.Background()
	out := channels.Map(ctx, in, func(fnCtx context.Context, i int) (bool, bool) {
		require.Equal(t, ctx, fnCtx)
		return i%2 == 0, i < 3
	})
	require.Equal(t, cap(in), cap(out))

	in <- 1
	in <- 2
	in <- 3

	require.Equal(t, <-out, false)
	require.Equal(t, <-out, true)
	require.Len(t, out, 0)
}

func TestMapValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)

	in <- 1
	in <- 2
	in <- 3
	close(in)

	ctx := context.Background()
	out := channels.MapValues(ctx, in, func(fnCtx context.Context, i int) (bool, bool) {
		require.Equal(t, ctx, fnCtx)
		return i%2 == 0, i < 3
	})

	require.Len(t, out, 2)
	require.Equal(t, out, []bool{false, true})
}
