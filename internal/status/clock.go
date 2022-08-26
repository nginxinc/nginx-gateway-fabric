package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . Clock

// Clock returns the current local time.
type Clock interface {
	Now() metav1.Time
}

// Real clock returns the current local time.
type RealClock struct{}

// NewRealClock creates a new RealClock.
func NewRealClock() *RealClock {
	return &RealClock{}
}

// Now returns the current local time.
func (c *RealClock) Now() metav1.Time {
	return metav1.Now()
}
