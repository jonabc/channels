package channels_test

import (
	"fmt"
	"testing"

	"github.com/jonabc/channels"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	// tests cover some base cases:
	// - 1 channels explicit handling
	// - 2 channels merge2
	// - 3 channels merge2 + merge1
	// - 4 channels merge4
	// - 5 channels multiple merge2 + merge1
	// - 11 channels multiple merge4 + merge2 + merge1
	channelCounts := []int{1, 2, 3, 4, 5, 11}
	for _, count := range channelCounts {
		count := count
		t.Run(fmt.Sprintf("With%dChannels", count), func(t *testing.T) {
			chans := make([](<-chan int), 0, count)
			for i := 0; i < count; i++ {
				channel := make(chan int)
				chans = append(chans, channel)

				go func(channel chan int, val int) {
					defer close(channel)

					// send one value, block, then send another
					channel <- val
				}(channel, i)
			}

			merged := channels.Merge[int](chans...)

			results := make([]int, 0, count)
			for out := range merged {
				results = append(results, out)
			}

			require.Len(t, results, count)
			for i := 0; i < count; i++ {
				require.Contains(t, results, i)
			}
		})
	}
}
