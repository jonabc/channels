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

type errorConfiguration interface {
	channelConfiguration
}

// Specify a channel which should receive the value returned by `recover` if
// a panic happens, for reporting and/or graceful handling.
func ErrorChannelOption[T errorConfiguration](panc chan<- any) Option[T] {
	return func(cfg *T) {
		switch cfg := any(cfg).(type) {
		case *BatchConfig:
			cfg.panc = panc
		case *DebounceConfig:
			cfg.panc = panc
		case *FlatMapConfig:
			cfg.panc = panc
		case *MapConfig:
			cfg.panc = panc
		case *MergeConfig:
			cfg.panc = panc
		case *ReduceConfig:
			cfg.panc = panc
		case *RejectConfig:
			cfg.panc = panc
		case *SelectConfig:
			cfg.panc = panc
		case *SplitConfig:
			cfg.panc = panc
		case *TapConfig:
			cfg.panc = panc
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
func ChannelCapacityOption[T singleOutputConfiguration](capacity int) func(*T) {
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
func MultiChannelCapacitiesOption[T multiOutputConfiguration](capacities []int) func(*T) {
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
