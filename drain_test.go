package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestDrainReturnsTrueWhenChannelIsDrainedAndClosed(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	go func() {
		time.Sleep(20 * time.Millisecond)
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()

	count, drained := channels.Drain(in, 0)
	require.Equal(t, 3, count)
	require.True(t, drained)
}

func TestDrainReturnsFalseWhenMaxWaitElapses(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	go func() {
		time.Sleep(20 * time.Millisecond)
		in <- 1
		in <- 2
		in <- 3
		close(in)
	}()

	count, drained := channels.Drain(in, 2*time.Millisecond)
	require.Equal(t, 0, count)
	require.False(t, drained)
}
