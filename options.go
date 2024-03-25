package channels

type channelConfiguration interface {
	BatchConfig |
		DebounceConfig |
		FlatMapConfig |
		MapConfig |
		MergeConfig |
		ReduceConfig |
		RejectConfig |
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
		case *FlatMapConfig:
			cfg.panicProvider = provider
		case *MapConfig:
			cfg.panicProvider = provider
		case *MergeConfig:
			cfg.panicProvider = provider
		case *ReduceConfig:
			cfg.panicProvider = provider
		case *RejectConfig:
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
		ReduceConfig |
		RejectConfig |
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
		case *ReduceConfig:
			cfg.capacity = capacity
		case *RejectConfig:
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
