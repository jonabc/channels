package channels

import "time"

// Throttle is equivalent to Debounce with `channels.LeadDebounceType.
// Throttle reads values from the input channel and pushes them to the
// returned output channel followed by a period of `delay` duration during
// which any matching values will be throttled and not pushed to the output channel.

// The channel returned by Throttle is unbuffered by default and will be closd when
// the input channel is drained and closed.

// Throttle also returns a function which returns the number of values
// that are currently being throttled
func Throttle[T comparable](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	return Debounce(inc, delay, append(opts, DebounceTypeOption(LeadDebounceType))...)
}

// ThrottleCustom is equivalent to DebounceCustom with `channels.LeadDebounceType.
// ThrottleCustom is like Throttle but with per-item configurability over
// comparisons and delays. Where Throttle requires `comparable` values in
// the input channel, ThrottleCustom requires types that implement the
// `DebounceInput[K comparable, T Keyable[K]]` interface.

// The channel returned by ThrottleCustom is unbuffered by default and will be closd when
// the input channel is drained and closed.

// ThrottleCustom also returns a function which returns the number of values
// that are currently being throttled
func ThrottleCustom[K comparable, T DebounceInput[K, T]](inc <-chan T, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	return DebounceCustom(inc, append(opts, DebounceTypeOption(LeadDebounceType))...)
}
