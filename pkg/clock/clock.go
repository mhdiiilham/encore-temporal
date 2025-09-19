package clock

import "time"

// Clock defines an abstraction for retrieving the current time.
type Clock interface {
	Now() time.Time
}

// RealClock uses the system time.
type RealClock struct{}

// Now return time.Now()
func (RealClock) Now() time.Time {
	return time.Now()
}
