package providers

import "sync"

type providerFunc[TIn any, TOut any] func(TIn, *providerReceiver[TIn, TOut]) bool

// Providers wrap channels to provide a few quality of life benefits:
// 1. Writing to the providerReceiver while it is closing or after it is closed will not produce a panic
// 2. An providerReceiver can be configured for different behaviors on calling `Observe`
type providerReceiver[TIn any, TOut any] struct {
	outc      chan TOut
	done      chan struct{}
	closeOnce sync.Once

	observingFn providerFunc[TIn, TOut]
}

func newProviderReceiver[TIn any, TOut any](size int, fn providerFunc[TIn, TOut]) *providerReceiver[TIn, TOut] {
	return &providerReceiver[TIn, TOut]{
		outc:        make(chan TOut, size),
		done:        make(chan struct{}),
		observingFn: fn,
	}
}

// Returns true if the provider has been closed, false otherwise
func (o *providerReceiver[TIn, TOut]) IsClosed() bool {
	select {
	case <-o.done:
		return true
	default:
		return false
	}
}

// Close the provider
func (o *providerReceiver[TIn, TOut]) Close() {
	o.closeOnce.Do(func() {
		close(o.done)
		close(o.outc)
	})
}

// Returns a read-only channel managed by the provider
func (o *providerReceiver[TIn, TOut]) Channel() <-chan TOut {
	return o.outc
}

// Observe a value, sending it to the output channel if possible
func (o *providerReceiver[TIn, TOut]) Provide(val TIn) bool {
	// Give preference to the done channel to signal the provider is closed
	select {
	case <-o.done:
		return false
	default:
		return o.observingFn(val, o)
	}
}
