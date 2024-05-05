package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestVoid(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	channels.Void(in)

	in <- 1
	in <- 2

	// Void starts a goroutine, sleep for a small period of time here
	// to ensure that the goroutine has time to consume the channel
	time.Sleep(100 * time.Microsecond)
	require.Len(t, in, 0)
}
