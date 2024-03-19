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

	require.True(t, channels.Drain(in, 0))
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

	require.False(t, channels.Drain(in, 2*time.Millisecond))
}
