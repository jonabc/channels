package channels_test

import (
	"sync"
	"testing"
	"time"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestDebounce(t *testing.T) {
	inc := make(chan int, 100)

	delay := 20 * time.Millisecond
	outc, getDebouncedCount := channels.Debounce(inc, delay)
	require.Equal(t, cap(inc), cap(outc))

	start := time.Now()

	inc <- 1
	inc <- 1

	time.Sleep(2 * time.Millisecond)

	inc <- 2
	inc <- 1

	time.Sleep(2 * time.Millisecond)
	require.Equal(t, 2, getDebouncedCount())

	// still waiting for debounce, out channel should be empty
	require.Len(t, outc, 0)

	results := make([]int, 0)
	results = append(results, <-outc)

	require.ElementsMatch(t, results, []int{1})
	require.GreaterOrEqual(t, time.Since(start), delay)

	results = append(results, <-outc)
	require.ElementsMatch(t, results, []int{1, 2})
	require.GreaterOrEqual(t, time.Since(start), delay+(2*time.Millisecond))

	inc <- 1
	inc <- 2
	close(inc)

	results = results[:0]
	for result := range outc {
		results = append(results, result)
	}

	require.ElementsMatch(t, results, []int{1, 2})
}

type customDebouncingType struct {
	key   string
	value string
	delay time.Duration
}

func (d *customDebouncingType) Key() string {
	return d.key
}

func (d *customDebouncingType) Delay() time.Duration {
	return d.delay
}

func (d *customDebouncingType) Reduce(other *customDebouncingType) (*customDebouncingType, bool) {
	if d.key == "3" {
		return d, false
	}

	d.value += "," + other.value
	return d, true
}

func TestDebounceKeyed(t *testing.T) {
	inc := make(chan *customDebouncingType, 100)
	defer close(inc)

	delay := 5 * time.Millisecond
	outc, getDebouncedCount := channels.DebounceCustom(inc)
	require.Equal(t, cap(inc), cap(outc))

	start := time.Now()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		inc <- &customDebouncingType{key: "1", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "1", value: "val2", delay: 1 * time.Millisecond}
		time.Sleep(2 * time.Millisecond)

		// should have one debounced item after pushing "key:1" twice
		require.Equal(t, 1, getDebouncedCount())
		inc <- &customDebouncingType{key: "2", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "2", value: "val2", delay: 1 * time.Millisecond}
		inc <- &customDebouncingType{key: "1", value: "val3", delay: 2 * time.Millisecond}

		time.Sleep(2 * time.Millisecond)
		inc <- &customDebouncingType{key: "3", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "3", value: "shouldnotreduce", delay: 1 * time.Millisecond}
		wg.Wait()
	}()

	require.Len(t, outc, 0)

	// only the delay for the first element seen on any key counts
	require.Equal(t, &customDebouncingType{key: "1", value: "val1,val2,val3", delay: delay}, <-outc)

	require.GreaterOrEqual(t, time.Since(start), delay)

	require.Equal(t, 2, getDebouncedCount())
	require.Equal(t, &customDebouncingType{key: "2", value: "val1,val2", delay: delay}, <-outc)
	require.GreaterOrEqual(t, time.Since(start), delay+(2*time.Millisecond))

	require.Equal(t, 1, getDebouncedCount())
	require.Equal(t, &customDebouncingType{key: "3", value: "val1", delay: delay}, <-outc)
	require.GreaterOrEqual(t, time.Since(start), delay+(4*time.Millisecond))
}
