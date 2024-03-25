package providers

// The collecting provider collects values when writing to the
// output channel is blocked.
func NewCollectingProvider[T any](size int) (Provider[T], Receiver[[]T]) {
	bridge := make(chan T)

	provider := newProviderReceiver(size, func(val T, provider *providerReceiver[T, []T]) bool {
		bridge <- val
		return true
	})

	go func() {
		defer close(bridge)

		for {
			val, ok := waitForReadFromIn(bridge, provider.done)
			if !ok {
				return
			}

			if !waitForWriteToOut(val, bridge, provider.outc, provider.done) {
				return
			}
		}
	}()

	return provider, provider
}

func waitForReadFromIn[T any](in <-chan T, done <-chan struct{}) (T, bool) {
	// Wait for an item read from the in channel or the done channel closed
	select {
	case <-done:
		return *new(T), false
	case next, ok := <-in:
		return next, ok
	}
}

func waitForWriteToOut[T any](val T, in <-chan T, out chan<- []T, done <-chan struct{}) bool {
	collector := []T{val}

	for {
		// Give preference to the done channel to signal the provider is closed
		select {
		case <-done:
			return false
		default:
			// Wait for either a write to the output channel, or the provider to be closed.
			select {
			case <-done:
				return false
			case out <- collector:
				return true
			case next := <-in:
				// If an additional item is sent while waiting to write to the output channel,
				// add it to the collector batch
				collector = append(collector, next)
			}
		}
	}
}
