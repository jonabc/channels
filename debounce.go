package channels

import (
	"sync"
	"time"

	"github.com/jonabc/channels/providers"
)

// DebounceConfig contains user configurable options for the Debounce functions
type DebounceConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[DebounceStats]
	capacity      int
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

// The channel returned by Debounce is unbuffered by default.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.

// Debounce also returns a function which returns the number of debounced values
// that are currently being delayed
func Debounce[T comparable](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)

	inBridge := make(chan *debounceInputComparable[T])
	go func() {
		defer close(inBridge)
		for in := range inc {
			inBridge <- &debounceInputComparable[T]{val: in, delay: delay}
		}
	}()

	outBridge, getDebouncedCount := DebounceCustom(inBridge,
		PanicProviderOption[DebounceConfig](cfg.panicProvider),
		DebounceStatsProviderOption(cfg.statsProvider),
	)
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

// The channel returned by DebounceCustom is unbuffered by default.
// When the input channel is closed, any remaining values being delayed/debounced
// will be flushed to the output channel and the output channel will be closed.

// DebounceCustom also returns a function which returns the number of debounced values
// that are currently being delayed
func DebounceCustom[K comparable, T DebounceInput[K, T]](inc <-chan T, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	cfg := parseOpts(opts...)

	outc := make(chan T, cfg.capacity)
	done := make(chan struct{})
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	// the buffer stores a map of key value pairs of
	// items from the input channel currently being debounced
	buffer := debounceBuffer[K, T]{
		data:          make(map[K]*debounceItem[K, T]),
		done:          done,
		panicProvider: panicProvider,
		statsProvider: statsProvider,
	}

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		for next := range inc {
			buffer.process(next, func(t T) {
				outc <- t
			})
		}

		close(done)
		buffer.wg.Wait()
	}()

	return outc, buffer.len
}

type debounceItem[K comparable, T DebounceInput[K, T]] struct {
	count uint
	value T
}

// debounceBuffer stores debounced values and counts
type debounceBuffer[K comparable, T DebounceInput[K, T]] struct {
	sync.Mutex
	data map[K]*debounceItem[K, T]

	panicProvider providers.Provider[any]
	statsProvider providers.Provider[DebounceStats]

	wg   sync.WaitGroup
	done <-chan struct{}
}

func (buffer *debounceBuffer[K, T]) process(item T, onDebounce func(T)) {
	buffer.Lock()
	defer buffer.Unlock()

	key := item.Key()

	if existing, hasExistingValue := buffer.data[key]; hasExistingValue {
		existing.count++

		value, ok := existing.value.Reduce(item)
		if ok {
			existing.value = value
		}
		return
	}

	// for new items, start a routine to push the item
	buffer.data[key] = &debounceItem[K, T]{
		value: item,
		count: 1,
	}
	buffer.wg.Add(1)
	delay := item.Delay()
	go func(key K, delay time.Duration) {
		defer tryHandlePanic(buffer.panicProvider)
		defer buffer.wg.Done()

		start := time.Now()
		// call onDebounce after the specified delay
		// or when the done channel is closed
		select {
		case <-buffer.done:
		case <-time.After(delay):
		}

		duration := time.Since(start)

		buffer.Lock()
		defer buffer.Unlock()

		item := buffer.data[key]
		delete(buffer.data, key)
		onDebounce(item.value)

		tryProvideStats(DebounceStats{Delay: duration, Count: item.count}, buffer.statsProvider)
	}(key, delay)
}

func (buffer *debounceBuffer[K, T]) len() int {
	buffer.Lock()
	defer buffer.Unlock()

	return len(buffer.data)
}
