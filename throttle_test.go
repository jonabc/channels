package channels_test

import (
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestThrottle(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getThrottledCount := channels.Throttle(in, delay)

	in <- 1
	in <- 2
	start := time.Now()
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 1, getThrottledCount())

	in <- 1
	select {
	case <-out:
		require.FailNow(t, "Unexpected value during throttling period")
	case <-time.After(delay + 1*time.Millisecond):
	}
	require.Equal(t, 0, getThrottledCount())

	in <- 1
	in <- 2
	start = time.Now()
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 1, getThrottledCount())
}

func TestThrottleValues(t *testing.T) {
	t.Parallel()

	in := make(chan int, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getThrottledCount := channels.ThrottleValues(in, delay)

	in <- 1
	in <- 2
	start := time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getThrottledCount())

	in <- 1
	in <- 2
	select {
	case <-out:
		require.FailNow(t, "Unexpected value during throttling period")
	case <-time.After(delay + 1*time.Millisecond):
	}
	require.Equal(t, 0, getThrottledCount())

	in <- 1
	in <- 2
	start = time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getThrottledCount())
}

func TestThrottleCustom(t *testing.T) {
	t.Parallel()

	in := make(chan *customDebouncingType, 100)
	defer close(in)

	delay := 5 * time.Millisecond
	out, getThrottledCount := channels.ThrottleCustom(in)

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
	start := time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getThrottledCount())

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
	select {
	case <-out:
		require.FailNow(t, "Unexpected value during throttling period")
	case <-time.After(delay + 1*time.Millisecond):
	}
	require.Equal(t, 0, getThrottledCount())

	in <- &customDebouncingType{key: "1", value: "1", delay: delay}
	in <- &customDebouncingType{key: "2", value: "2", delay: delay}
	start = time.Now()
	<-out
	<-out
	require.Less(t, time.Since(start), delay)
	require.Equal(t, 2, getThrottledCount())
}
