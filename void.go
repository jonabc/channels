package channels

// Void consumes a channel until it is closed, doing nothing with the channel items.
func Void[T any](inc <-chan T) {
	go func() {
		for range inc {
		}
	}()
}
