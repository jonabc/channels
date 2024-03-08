package time

import "time"

type Ticker struct {
	internal *time.Ticker
	C        <-chan time.Time
}

func NewTicker(interval time.Duration) *Ticker {
	var internal *time.Ticker
	var c <-chan time.Time
	if interval > 0 {
		internal = time.NewTicker(interval)
		c = internal.C
	}

	return &Ticker{
		internal: internal,
		C:        c,
	}
}

func (t *Ticker) Reset(d time.Duration) {
	if t.internal != nil {
		t.internal.Reset(d)
	}
}

func (t *Ticker) Stop() {
	if t.internal != nil {
		t.internal.Stop()
	}
}
