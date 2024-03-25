package channels

import (
	"time"

	"github.com/jonabc/channels/providers"
)

// Stats provides an operation duration.
type Stats struct {
	Duration time.Duration
}

// BatchStats provides a batch operation's duration and batch size.
type BatchStats struct {
	Duration  time.Duration
	BatchSize uint
}

// DebounceStats provides a debounce operation's delay and debounced count.
type DebounceStats struct {
	Delay time.Duration
	Count uint
}

// SelectStats provides a select or reject operation's duration and
// whether the item was selected or not.
type SelectStats struct {
	Duration time.Duration
	Selected bool
}

// TapStats provides the duration of a tap operations pre and post functions.
type TapStats struct {
	PreDuration  time.Duration
	PostDuration time.Duration
}

type statsProviderInput interface {
	Stats | BatchStats | DebounceStats | SelectStats | TapStats
}

func tryProvideStats[T statsProviderInput](stats T, provider providers.Provider[T]) {
	if provider == nil {
		return
	}

	provider.Provide(stats)
}
