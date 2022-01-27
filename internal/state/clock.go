package state

import "time"

// Clock returns the current local time.
type Clock interface {
	Now() time.Time
}

// Real clock returns the current local time.
type RealClock struct {
}

// NewRealClock creates a new RealClock.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current local time.
func (c *RealClock) Now() time.Time {
	return time.Now()
}

// FakeClock allows you to control the returned time.
type FakeClock struct {
	time time.Time
}

// NewFakeClock creates a FakeClock. The clock will always return the specified time.
func NewFakeClock(time time.Time) *FakeClock {
	return &FakeClock{time: time}
}

// Now is a fake implementation of Now().
func (c FakeClock) Now() time.Time {
	return c.time
}
