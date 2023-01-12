package reconciler

import "k8s.io/apimachinery/pkg/runtime"

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . EventRecorder

// EventRecorder records events for a resource.
// It allows us to mock the record.EventRecorder.Eventf method.
type EventRecorder interface {
	// Eventf is a method of k8s.io/client-go/tools/record.EventRecorder
	Eventf(object runtime.Object, eventtype, reason, messageFmt string, args ...interface{})
}
