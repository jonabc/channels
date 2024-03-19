package channels

import (
	"time"

	internalTime "github.com/jonabc/channels/internal/time"
)

// Drain blocks until either the input channel is fully drained and closed or `maxWait` duration has passed.
// Drain returns true when exiting due to the input channel being drained and closed, false when exiting
// due to waiting for the `maxWait` duration.
// When `maxWait <= 0`, Drain will wait forever, and only exit when the input channel is closed.
func Drain[T any](inc <-chan T, maxWait time.Duration) bool {
	ticker := internalTime.NewTicker(maxWait)
	defer ticker.Stop()

	for {
		select {
		case _, ok := <-inc:
			if !ok {
				return true
			}
		case <-ticker.C:
			return false
		}
	}
}
