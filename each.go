package channels

import (
	"time"

	"github.com/jonabc/channels/providers"
)

type EachConfig struct {
	panicProvider providers.Provider[any]
	statsProvider providers.Provider[Stats]
}

func Each[T any](inc <-chan T, eachFn func(T), opts ...Option[EachConfig]) {
	cfg := parseOpts(opts...)

	panicProvider := cfg.panicProvider
	statsProvider := cfg.statsProvider

	go func() {
		defer tryHandlePanic(panicProvider)

		for in := range inc {
			start := time.Now()
			eachFn(in)
			duration := time.Since(start)

			tryProvideStats(Stats{Duration: duration}, statsProvider)
		}
	}()
}
