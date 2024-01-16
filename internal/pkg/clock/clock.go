package clock

import "time"

// Clock  - interface for testing purpose
type Clock interface {
	Now() time.Time
}

type SystemClock struct {
}

func (s SystemClock) Now() time.Time {
	return time.Now()
}
