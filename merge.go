package channels

import (
	"sync"
)

// Merge merges multiple input channels into a single output channel.  The
// order of values in the output channel is not guaranteed to match the
// order that values are written to the input channels.  The output channel
// has the same capacity as the input channel and is closed when all input
// channels are closed.
func Merge[T any](chans ...<-chan T) <-chan T {
	switch len(chans) {
	case 0:
		return nil
	case 1:
		return chans[0]
	default:
		var wg sync.WaitGroup
		outc := make(chan T)
		i := 0

		for len(chans)-i >= 4 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				merge4(outc, chans[i], chans[i+1], chans[i+2], chans[i+3])
			}(i)
			i += 4
		}

		for len(chans)-i >= 2 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				merge2(outc, chans[i], chans[i+1])
			}(i)
			i += 2
		}

		for len(chans)-i >= 1 {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				for v := range chans[i] {
					outc <- v
				}
			}(i)
			i++
		}

		go func() {
			wg.Wait()
			close(outc)
		}()
		return outc
	}
}

func merge2[T any](outc chan<- T, inc1, inc2 <-chan T) {
	for inc1 != nil || inc2 != nil {
		select {
		case val, ok := <-inc1:
			if !ok {
				inc1 = nil
			} else {
				outc <- val
			}
		case val, ok := <-inc2:
			if !ok {
				inc2 = nil
			} else {
				outc <- val
			}
		}
	}
}

func merge4[T any](outc chan<- T, inc1, inc2, inc3, inc4 <-chan T) {
	for inc1 != nil || inc2 != nil || inc3 != nil || inc4 != nil {
		select {
		case val, ok := <-inc1:
			if !ok {
				inc1 = nil
			} else {
				outc <- val
			}
		case val, ok := <-inc2:
			if !ok {
				inc2 = nil
			} else {
				outc <- val
			}
		case val, ok := <-inc3:
			if !ok {
				inc3 = nil
			} else {
				outc <- val
			}
		case val, ok := <-inc4:
			if !ok {
				inc4 = nil
			} else {
				outc <- val
			}
		}
	}
}
