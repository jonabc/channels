package channels

import (
	"sync"
	"time"

	"github.com/jonabc/channels/providers"
)

type DebounceType byte

const (
	TailDebounceType DebounceType = 1 << iota
	LeadDebounceType
	LeadTailDebounceType = TailDebounceType | LeadDebounceType
)

// DebounceConfig contains user configurable options for the Debounce functions
type DebounceConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[DebounceStats]
	capacity      int
	debounceType  DebounceType
}

func defaultDebounceOptions() []Option[DebounceConfig] {
	return []Option[DebounceConfig]{
		DebounceTypeOption(TailDebounceType),
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

// Debounce reads values from the input channel and pushes them to the returned
// output channel before, after, or before and after a `delay` debounce period.
// The `DebounceType` value used in the function controls when the value is pushed
// to the output channel.  Any duplicate values read from the input channel during
// the `delay` debounce period are ignored.

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
		append(opts, ChannelCapacityOption[DebounceConfig](0))...,
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
	cfg := parseOpts(append(defaultDebounceOptions(), opts...)...)

	outc := make(chan T, cfg.capacity)
	done := make(chan struct{})
	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider
	debounceType := cfg.debounceType

	// the buffer stores a map of key value pairs of
	// items from the input channel currently being debounced
	buffer := debounceBuffer[K, T]{
		data: make(map[K]*debounceItem[K, T]),
	}

	go func() {
		defer tryHandlePanic(panicProvider)
		defer close(outc)

		var wg sync.WaitGroup
		for next := range inc {
			key := next.Key()
			if buffer.add(key, next) {
				wg.Add(1)

				go func(key K, delay time.Duration) {
					defer tryHandlePanic(panicProvider)
					defer wg.Done()

					start := time.Now()

					timer := time.NewTimer(delay)
					select {
					case <-done:
						timer.Stop()
					case <-timer.C:
					}

					duration := time.Since(start)
					item, count := buffer.remove(key)

					if debounceType&TailDebounceType == TailDebounceType {
						outc <- item
					}
					tryProvideStats(DebounceStats{Delay: duration, Count: count}, statsProvider)
				}(key, next.Delay())

				if debounceType&LeadDebounceType == LeadDebounceType {
					outc <- next
				}
			}
		}

		close(done)
		wg.Wait()
	}()

	return outc, buffer.len
}

type debounceItem[K comparable, T DebounceInput[K, T]] struct {
	value T
	count uint
}

// debounceBuffer stores debounced values and counts
type debounceBuffer[K comparable, T DebounceInput[K, T]] struct {
	data map[K]*debounceItem[K, T]
	sync.Mutex
}

func (buffer *debounceBuffer[K, T]) add(key K, value T) bool {
	buffer.Lock()
	defer buffer.Unlock()

	if existing, hasExistingValue := buffer.data[key]; hasExistingValue {
		existing.count++

		value, ok := existing.value.Reduce(value)
		if ok {
			existing.value = value
		}
		return false
	}

	buffer.data[key] = &debounceItem[K, T]{
		value: value,
		count: 1,
	}

	return true
}

func (buffer *debounceBuffer[K, T]) remove(key K) (T, uint) {
	buffer.Lock()
	defer buffer.Unlock()

	item := buffer.data[key]
	delete(buffer.data, key)

	return item.value, item.count
}

func (buffer *debounceBuffer[K, T]) len() int {
	buffer.Lock()
	defer buffer.Unlock()

	return len(buffer.data)
}
