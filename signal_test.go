package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestWithDone(t *testing.T) {
	t.Parallel()

	inc := make(chan int, 4)

	outc, done := channels.WithDone(inc)
	require.Equal(t, cap(inc), cap(outc))
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

func TestWithSignal(t *testing.T) {
	t.Parallel()

	inc := make(chan int, 4)

	outc, signalc := channels.WithSignal(inc, func(i int) (time.Time, bool) {
		return time.Now(), i > 10
	})

	inc <- 1
	inc <- 11
	inc <- 4
	inc <- 100
	close(inc)

	results := make([]int, 0, 4)
	for result := range outc {
		results = append(results, result)
	}
	require.Equal(t, []int{1, 11, 4, 100}, results)

	signals := make([]time.Time, 0)
	for signal := range signalc {
		signals = append(signals, signal)
	}
	require.Len(t, signals, 2)
}
