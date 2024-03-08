package time

import "time"

type Timer struct {
	internal *time.Timer
	C        <-chan time.Time
}

func NewTimer(interval time.Duration) *Timer {
	var internal *time.Timer
	var c <-chan time.Time
	if interval > 0 {
		internal = time.NewTimer(interval)
		c = internal.C
	}

	return &Timer{
		internal: internal,
		C:        c,
	}
}

func (t *Timer) Reset(d time.Duration) {
	if t.internal != nil {
		t.internal.Reset(d)
	}
}

func (t *Timer) Stop() {
	if t.internal != nil {
		t.internal.Stop()
	}
}
