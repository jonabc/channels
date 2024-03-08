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
	outc := channels.Debounce(inc, delay)
	require.Equal(t, cap(inc), cap(outc))

	start := time.Now()

	inc <- 1
	inc <- 1

	time.Sleep(2 * time.Millisecond)

	inc <- 2
	inc <- 1

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

func (d *customDebouncingType) Reduce(other *customDebouncingType) *customDebouncingType {
	d.value += "," + other.value
	return d
}

func TestDebounceKeyed(t *testing.T) {
	inc := make(chan *customDebouncingType, 100)
	defer close(inc)

	delay := 5 * time.Millisecond
	outc := channels.DebounceCustom(inc)
	require.Equal(t, cap(inc), cap(outc))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		inc <- &customDebouncingType{key: "1", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "1", value: "val2", delay: 1 * time.Millisecond}
		inc <- &customDebouncingType{key: "2", value: "val1", delay: delay}
		inc <- &customDebouncingType{key: "2", value: "val2", delay: 1 * time.Millisecond}
		inc <- &customDebouncingType{key: "1", value: "val3", delay: 2 * time.Millisecond}
		wg.Wait()
	}()

	start := time.Now()
	for time.Since(start) < 5*time.Millisecond {
		require.Len(t, outc, 0)
		time.Sleep(1 * time.Millisecond)
	}

	require.Len(t, outc, 2)

	results := make([]*customDebouncingType, 0)
	results = append(results, <-outc)
	results = append(results, <-outc)
	// only the delay for the first element seen on any key counts
	require.ElementsMatch(t, results, []*customDebouncingType{
		{key: "1", value: "val1,val2,val3", delay: delay},
		{key: "2", value: "val1,val2", delay: delay},
	})
	require.Len(t, outc, 0)
	wg.Done()
}
