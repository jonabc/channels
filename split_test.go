package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestSplit(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	defer close(in)

	outArray := channels.Split(in, 2, func(i int, chans []chan<- int) {
		require.Len(t, chans, 2)
		chans[i%2] <- i
	})
	require.Equal(t, 2, len(outArray))

	evens := outArray[0]
	require.Equal(t, cap(in), cap(evens))

	odds := outArray[1]
	require.Equal(t, cap(in), cap(odds))

	in <- 1
	in <- 2
	in <- 3
	in <- 4

	time.Sleep(1 * time.Millisecond)

	require.Len(t, evens, 2)
	require.Equal(t, <-evens, 2)
	require.Equal(t, <-evens, 4)

	require.Len(t, odds, 2)
	require.Equal(t, <-odds, 1)
	require.Equal(t, <-odds, 3)
}

func TestSplitValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 10)
	in <- 1
	in <- 2
	in <- 3
	in <- 4
	close(in)

	results := channels.SplitValues(in, 2, func(i int, chans []chan<- int) {
		require.Len(t, chans, 2)
		chans[i%2] <- i
	})
	require.Equal(t, 2, len(results))

	evens := results[0]
	require.ElementsMatch(t, evens, []int{2, 4})

	odds := results[1]
	require.ElementsMatch(t, odds, []int{1, 3})
}
