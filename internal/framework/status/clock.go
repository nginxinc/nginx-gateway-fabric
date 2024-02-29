package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//counterfeiter:generate . Clock

// Clock returns the current local time.
type Clock interface {
	Now() metav1.Time
}

// RealClock returns the current local time.
type RealClock struct{}

// NewRealClock creates a new RealClock.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current local time.
func (c *RealClock) Now() metav1.Time {
	return metav1.Now()
}
