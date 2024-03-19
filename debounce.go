package channels

import (
	"sync"
	"time"
)

// DebounceConfig contains user configurable options for the Debounce functions
type DebounceConfig struct {
	panc     chan<- any
	capacity int
}

func defaultDebounceOptions[T any](inc <-chan T) []Option[DebounceConfig] {
	return []Option[DebounceConfig]{
		ChannelCapacityOption[DebounceConfig](cap(inc)),
	}
}

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
func Debounce[T comparable](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	cfg := parseOpts(append(defaultDebounceOptions(inc), opts...)...)

	outc := make(chan T, cfg.capacity)
	panc := cfg.panc

	inBridge := make(chan *debounceInputComparable[T])
	go func() {
		defer handlePanicIfErrc(panc)
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceInputComparable[T]{val: in, delay: delay}
		}
	}()

	outBridge, getDebouncedCount := DebounceCustom(inBridge, ErrorChannelOption[DebounceConfig](panc))
	go func() {
		defer handlePanicIfErrc(panc)
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
func DebounceCustom[K comparable, T DebounceInput[K, T]](inc <-chan T, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	cfg := parseOpts(append(defaultDebounceOptions(inc), opts...)...)

	outc := make(chan T, cfg.capacity)
	done := make(chan struct{})
	panc := cfg.panc

	// the buffer stores a map of key value pairs of
	// items from the input channel currently being debounced
	buffer := debounceBuffer[K, T]{
		data: make(map[K]T),
		done: done,
	}

	go func() {
		defer handlePanicIfErrc(panc)
		defer close(outc)

		for next := range inc {
			buffer.process(next, panc, func(t T) {
				outc <- t
			})
		}

		close(done)
		buffer.wg.Wait()
	}()

	return outc, buffer.len
}

// debounceBuffer stores debounced values and
type debounceBuffer[K comparable, T DebounceInput[K, T]] struct {
	sync.Mutex
	data map[K]T

	wg   sync.WaitGroup
	done <-chan struct{}
}

func (buffer *debounceBuffer[K, T]) process(item T, panc chan<- any, onDebounce func(T)) {
	buffer.Lock()
	defer buffer.Unlock()

	key := item.Key()

	if value, hasExistingValue := buffer.data[key]; hasExistingValue {
		value, ok := value.Reduce(item)
		if ok {
			buffer.data[key] = value
		}
		return
	}

	// for new items, start a routine to push the item
	buffer.data[key] = item
	buffer.wg.Add(1)
	delay := item.Delay()
	go func(key K, delay time.Duration) {
		defer handlePanicIfErrc(panc)
		defer buffer.wg.Done()

		// call onDebounce after the specified delay
		// or when the done channel is closed
		select {
		case <-buffer.done:
		case <-time.After(delay):
		}

		buffer.Lock()
		defer buffer.Unlock()

		item := buffer.data[key]
		delete(buffer.data, key)
		onDebounce(item)
	}(key, delay)
}

func (buffer *debounceBuffer[K, T]) len() int {
	buffer.Lock()
	defer buffer.Unlock()

	return len(buffer.data)
}
