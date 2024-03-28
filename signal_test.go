package channels_test

import (
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestWithDone(t *testing.T) {
	t.Parallel()

	inc := make(chan int, 4)

	outc, done := channels.WithDone(inc)
	require.Equal(t, 0, cap(outc))
	require.Equal(t, 0, cap(done))
	require.Empty(t, outc)

	go func() {
		inc <- 1
		inc <- 2
		close(inc)
	}()

	results := make([]int, 0, 2)
	for done != nil || outc != nil {
		select {
		case val, ok := <-outc:
			if !ok {
				outc = nil
				break
			}
			results = append(results, val)
		case <-done:
			done = nil
		}
	}

	require.Equal(t, []int{1, 2}, results)
	require.Nil(t, done)
}
