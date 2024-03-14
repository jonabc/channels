package channels

import (
	"sync"
	"time"
)

type Keyable[K comparable] interface {
	Key() K
}

type DebounceInput[K comparable, T Keyable[K]] interface {
	Keyable[K]
	Delay() time.Duration
	Reduce(T) (T, bool)
}

type debounceInputComparable[T comparable] struct {
	val   T
	delay time.Duration
}

func (i *debounceInputComparable[T]) Key() T {
	return i.val
}

func (i *debounceInputComparable[T]) Delay() time.Duration {
	return i.delay
}

func (i *debounceInputComparable[T]) Reduce(*debounceInputComparable[T]) (*debounceInputComparable[T], bool) {
	return i, true
}

// Debounce reads values from the input channel and pushes them to the
// returned output channel after a delay.  If a value is read from the
// input channel multiple times during the debounce period it will only
// be pushed to the output channel once, after a `delay` started from
// when the first of the value multiple values is read.

// The channel returned by Debounce has the same capacity as the input channel.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.

// Debounce also returns a function which returns the number of debounced values
// that are currently being delayed
func Debounce[T comparable](inc <-chan T, delay time.Duration) (<-chan T, func() int) {
	outc := make(chan T, cap(inc))

	inBridge := make(chan *debounceInputComparable[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceInputComparable[T]{val: in, delay: delay}
		}
	}()

	outBridge, getDebouncedCount := DebounceCustom[T, *debounceInputComparable[T]](inBridge)
	go func() {
		defer close(outc)
		for out := range outBridge {
			outc <- out.val
		}
	}()

	return outc, getDebouncedCount
}

// DebounceCustom is like Debounce but with per-item configurability over
// comparisons, delays, and reducing multiple values to a single debounced value.
// Where Debounce requires `comparable` values in the input channel,
// DebounceCustom requires types that implement the `DebounceInput[K comparable, T Keyable[K]]`
// interface.

// The channel returned by DebounceCustom has the same capacity as the input channel.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.

// DebounceCustom also returns a function which returns the number of debounced values
// that are currently being delayed
func DebounceCustom[K comparable, T DebounceInput[K, T]](inc <-chan T) (<-chan T, func() int) {
	outc := make(chan T, cap(inc))

	// the buffer stores a map of key value pairs of
	// items from the input channel currently being debounced
	buffer := struct {
		sync.Mutex
		data map[K]T
	}{
		data: make(map[K]T),
	}

	go func() {
		defer close(outc)

		var wg sync.WaitGroup
		done := make(chan struct{})

		for next := range inc {
			key := next.Key()

			if value, ok := buffer.data[key]; ok {
				next, ok = value.Reduce(next)
				if !ok {
					next = value
				}
			} else {
				// add 1 to the waitgroup, the value will be decremented when the
				// debounced value is read from the waitAny channel
				wg.Add(1)
				delay := next.Delay()

				go func(key K, delay time.Duration) {
					defer wg.Done()

					// push the key to the output channel after the delay,
					// or when the done channel is closed, signaling
					// to flush all remaining values
					select {
					case <-done:
					case <-time.After(delay):
					}

					value := buffer.data[key]

					// lock on write
					buffer.Lock()
					delete(buffer.data, key)
					buffer.Unlock()

					outc <- value
				}(key, delay)
			}

			// lock only on write
			buffer.Lock()
			buffer.data[key] = next
			buffer.Unlock()
		}

		close(done)
		wg.Wait()
	}()

	return outc, func() int {
		buffer.Lock()
		defer buffer.Unlock()

		return len(buffer.data)
	}
}
