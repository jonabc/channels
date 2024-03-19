package channels

func handlePanicIfErrc(panc chan<- any) {
	// don't handle the panic if a panic channel isn't provided
	// is not provided
	if panc == nil {
		return
	}

	if err := recover(); err != nil {
		panc <- err
	}
}
