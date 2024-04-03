package channels

import "time"

// Throttle is equivalent to Debounce with `channels.LeadDebounceType`.
// See Debounce for usage details.
func Throttle[T comparable](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	return Debounce(inc, delay, append(opts, DebounceTypeOption(LeadDebounceType))...)
}

// ThrottleValues is equivalent to DebounceValues with `channels.LeadDebounceType`.
// See DebounceValues for usage details.
func ThrottleValues[T comparable](inc <-chan T, delay time.Duration, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	return DebounceValues(inc, delay, append(opts, DebounceTypeOption(LeadDebounceType))...)
}

// ThrottleCustom is equivalent to DebounceCustom with `channels.LeadDebounceType`.
// See DebounceCustom for usage details.
func ThrottleCustom[K comparable, T DebounceInput[K, T]](inc <-chan T, opts ...Option[DebounceConfig]) (<-chan T, func() int) {
	return DebounceCustom(inc, append(opts, DebounceTypeOption(LeadDebounceType))...)
}
