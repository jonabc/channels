package channels

import (
	"github.com/jonabc/channels/providers"
)

type channelConfiguration interface {
	BatchConfig |
		DebounceConfig |
		EachConfig |
		FlatMapConfig |
		MapConfig |
		MergeConfig |
		ReduceConfig |
		SelectConfig |
		SplitConfig |
		TapConfig
}

type Option[T channelConfiguration] func(*T)

func PanicProviderOption[T channelConfiguration](provider providers.Provider[any]) Option[T] {
	return func(cfg *T) {
		switch cfg := any(cfg).(type) {
		case *BatchConfig:
			cfg.panicProvider = provider
		case *DebounceConfig:
			cfg.panicProvider = provider
		case *EachConfig:
			cfg.panicProvider = provider
		case *FlatMapConfig:
			cfg.panicProvider = provider
		case *MapConfig:
			cfg.panicProvider = provider
		case *MergeConfig:
			cfg.panicProvider = provider
		case *ReduceConfig:
			cfg.panicProvider = provider
		case *SelectConfig:
			cfg.panicProvider = provider
		case *SplitConfig:
			cfg.panicProvider = provider
		case *TapConfig:
			cfg.panicProvider = provider
		}
	}
}

type singleOutputConfiguration interface {
	BatchConfig |
		DebounceConfig |
		FlatMapConfig |
		MapConfig |
		MergeConfig |
		ReduceConfig |
		SelectConfig |
		TapConfig
}

// Specify the capacity for a single output channel returned by a channels function.
func ChannelCapacityOption[T singleOutputConfiguration](capacity int) Option[T] {
	return func(cfg *T) {
		switch cfg := any(cfg).(type) {
		case *BatchConfig:
			cfg.capacity = capacity
		case *DebounceConfig:
			cfg.capacity = capacity
		case *FlatMapConfig:
			cfg.capacity = capacity
		case *MapConfig:
			cfg.capacity = capacity
		case *MergeConfig:
			cfg.capacity = capacity
		case *ReduceConfig:
			cfg.capacity = capacity
		case *SelectConfig:
			cfg.capacity = capacity
		case *TapConfig:
			cfg.capacity = capacity
		}
	}
}

type multiOutputConfiguration interface {
	SplitConfig
}

// Specify the capacities for output channels created from functions which return multiple channels.
func MultiChannelCapacitiesOption[T multiOutputConfiguration](capacities []int) Option[T] {
	return func(cfg *T) {
		switch cfg := any(cfg).(type) {
		case *SplitConfig:
			cfg.capacities = capacities
		}
	}
}

func parseOpts[C any, O ~func(*C)](opts ...O) *C {
	cfg := new(C)
	for _, opt := range opts {
		opt(cfg)
	}

	return cfg
}

type statsConfiguration interface {
	EachConfig |
		FlatMapConfig |
		MapConfig |
		ReduceConfig |
		SplitConfig
}

// Specify a stats provider to receive information about operations that emit a duration.
func StatsProviderOption[T statsConfiguration](provider providers.Provider[Stats]) Option[T] {
	return func(cfg *T) {
		switch cfg := any(cfg).(type) {
		case *EachConfig:
			cfg.statsProvider = provider
		case *FlatMapConfig:
			cfg.statsProvider = provider
		case *MapConfig:
			cfg.statsProvider = provider
		case *ReduceConfig:
			cfg.statsProvider = provider
		case *SplitConfig:
			cfg.statsProvider = provider
		}
	}
}

// Specify a stats provider to receive information about batch operations.
func BatchStatsProviderOption(provider providers.Provider[BatchStats]) Option[BatchConfig] {
	return func(cfg *BatchConfig) {
		cfg.statsProvider = provider
	}
}

// Specify a stats provider to receive information about debounce operations.
func DebounceStatsProviderOption(provider providers.Provider[DebounceStats]) Option[DebounceConfig] {
	return func(cfg *DebounceConfig) {
		cfg.statsProvider = provider
	}
}

// Specify a stats provider to receive information about select and reject operations.
func SelectStatsProviderOption(provider providers.Provider[SelectStats]) Option[SelectConfig] {
	return func(cfg *SelectConfig) {
		cfg.statsProvider = provider
	}
}

// Specify a stats provider to receive information about tap operations.
func TapStatsProviderOption(provider providers.Provider[TapStats]) Option[TapConfig] {
	return func(cfg *TapConfig) {
		cfg.statsProvider = provider
	}
}

// Specify the type of debouncing to a debouncer function - lead, tail, or both.
func DebounceTypeOption(debounceType DebounceType) Option[DebounceConfig] {
	return func(cfg *DebounceConfig) {
		cfg.debounceType = debounceType
	}
}
