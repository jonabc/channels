package providers

func NewDroppingProvider[T any](size int) (Provider[T], Receiver[T]) {
	provider := newProviderReceiver(size, func(val T, provider *providerReceiver[T, T]) bool {
		select {
		case provider.outc <- val:
		default:
			// drop the observed value if the output channel is blocked
		}
		return true
	})

	return provider, provider
}
