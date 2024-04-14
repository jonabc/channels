package channels

import (
	"time"

	internalTime "github.com/jonabc/channels/internal/time"
)

// Drain blocks until either the input channel is fully drained and closed or `maxWait` duration has passed.
// Drain returns the count of values drained from the channel, and a bool that is true when exiting due to
// the input channel being drained and closed or false when exiting due to waiting for the `maxWait` duration.
// When `maxWait <= 0`, Drain will wait forever, and only exit when the input channel is closed.
func Drain[T any](inc <-chan T, maxWait time.Duration) (int, bool) {
	ticker := internalTime.NewTicker(maxWait)
	defer ticker.Stop()

	count := 0

	for {
		select {
		case _, ok := <-inc:
			if !ok {
				return count, true
			}
			count++
		case <-ticker.C:
			return count, false
		}
	}
}

// DrainValues blocks until either the input channel is fully drained and closed or `maxWait` duration has passed.
// DrainValues returns the values drained from the channel, and a bool that is true when exiting due to
// the input channel being drained and closed or false when exiting due to waiting for the `maxWait` duration.
// When `maxWait <= 0`, Drain will wait forever, and only exit when the input channel is closed.
func DrainValues[T any](inc <-chan T, maxWait time.Duration) ([]T, bool) {
	ticker := internalTime.NewTicker(maxWait)
	defer ticker.Stop()

	values := []T{}

	for {
		select {
		case val, ok := <-inc:
			if !ok {
				return values, true
			}
			values = append(values, val)
		case <-ticker.C:
			return values, false
		}
	}
}
