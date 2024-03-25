package channels

import "github.com/jonabc/channels/providers"

func tryHandlePanic(provider providers.Provider[any]) {
	// don't handle the panic if a panic provider isn't provided
	if provider == nil {
		return
	}

	if err := recover(); err != nil {
		provider.Provide(err)
	}
}
