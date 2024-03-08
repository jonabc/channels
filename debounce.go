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
	Reduce(T) T
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

func (i *debounceInputComparable[T]) Reduce(*debounceInputComparable[T]) *debounceInputComparable[T] {
	return i
}

// Debounce reads values from the input channel and pushes them to the
// returned output channel after a delay.  If a value is read from the
// input channel multiple times during the debounce period it will only
// be pushed to the output channel once, after a `delay` started from
// when the first of the value multiple values is read.

// The channel returned by Debounce has the same capacity as the input channel.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.
func Debounce[T comparable](inc <-chan T, delay time.Duration) <-chan T {
	outc := make(chan T, cap(inc))

	inBridge := make(chan *debounceInputComparable[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceInputComparable[T]{val: in, delay: delay}
		}
	}()

	outBridge := DebounceCustom[T, *debounceInputComparable[T]](inBridge)
	go func() {
		defer close(outc)
		for out := range outBridge {
			outc <- out.val
		}
	}()

	return outc
}

// DebounceCustom is like Debounce but with per-item configurability over
// comparisons, delays, and reducing multiple values to a single debounced value.
// Where Debounce requires `comparable` values in the input channel,
// DebounceCustom requires types that implement the `DebounceInput[K comparable, T Keyable[K]]`
// interface.

// The channel returned by DebounceCustom has the same capacity as the input channel.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.
func DebounceCustom[K comparable, T DebounceInput[K, T]](inc <-chan T) <-chan T {
	outc := make(chan T, cap(inc))

	go func() {
		defer close(outc)

		var wg sync.WaitGroup
		done := make(chan struct{})

		buffer := struct {
			sync.Mutex
			data map[K]T
		}{
			data: make(map[K]T),
		}

		var waitAny <-chan K

		for inc != nil || waitAny != nil {
			select {
			case next, ok := <-inc:
				if !ok {
					close(done)
					inc = nil
					break
				}

				key := next.Key()
				value, ok := buffer.data[key]
				if ok {
					next = value.Reduce(next)
				} else {
					// add 1 to the waitgroup, the value will be decremented when the
					// debounced value is read from the waitAny channel
					wg.Add(1)
					debounceChan := make(chan K)
					delay := next.Delay()
					if delay == 0 {
						delay = 1 * time.Nanosecond
					}

					go func(key K, delay time.Duration) {
						defer close(debounceChan)

						// push the key to the values debounce channel
						// after a delay, or when the done channel is closed
						// signalling the input channel is closed and values
						// should be flushed
						select {
						case <-done:
						case <-time.After(delay):
						}
						debounceChan <- key
					}(key, delay)

					// merge the new debounce channel with existing waiters to build a tree of
					// debouncer notifications that can be handled one at a time.
					waitAny = Merge(waitAny, debounceChan)
				}

				buffer.data[key] = next
			case completed, ok := <-waitAny:
				if !ok {
					waitAny = nil
					break
				}

				wg.Done()
				outc <- buffer.data[completed]
				delete(buffer.data, completed)
			}
		}

		wg.Wait()
	}()

	return outc
}
