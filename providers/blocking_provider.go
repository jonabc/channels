package providers

// The default provider blocks on calls to Observe when
// the underlying channel blocks.
func NewProvider[T any](size int) (Provider[T], Receiver[T]) {
	provider := newProviderReceiver(size, func(val T, provider *providerReceiver[T, T]) bool {
		select {
		case <-provider.done:
			return false
		case provider.outc <- val:
			return true
		}
	})

	return provider, provider
}
