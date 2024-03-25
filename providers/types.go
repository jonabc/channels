package providers

// Providers provide values to receives
type Provider[T any] interface {
	// Returns true if the provider is closed
	IsClosed() bool

	// Close the provider
	Close()

	// Provide a value to receivers
	Provide(T) bool
}

// Receivers receive values from providers
type Receiver[T any] interface {
	// Returns true if the receive is closed
	IsClosed() bool

	// Close the receiver
	Close()

	// Returns a channel which receives values from a provider
	Channel() <-chan T
}
