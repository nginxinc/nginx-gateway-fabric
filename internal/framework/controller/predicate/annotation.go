package predicate

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// AnnotationPredicate implements a predicate function based on the Annotation.
//
// This predicate will skip the following events:
// 1. Create events that do not contain the Annotation.
// 2. Update events where the Annotation value has not changed.
type AnnotationPredicate struct {
	predicate.Funcs
	Annotation string
}

// Create filters CreateEvents based on the Annotation.
func (cp AnnotationPredicate) Create(e event.CreateEvent) bool {
	if e.Object == nil {
		return false
	}

	_, ok := e.Object.GetAnnotations()[cp.Annotation]
	return ok
}

// Update filters UpdateEvents based on the Annotation.
func (cp AnnotationPredicate) Update(e event.UpdateEvent) bool {
	if e.ObjectOld == nil || e.ObjectNew == nil {
		// this case should not happen
		return false
	}

	oldAnnotationVal := e.ObjectOld.GetAnnotations()[cp.Annotation]
	newAnnotationVal := e.ObjectNew.GetAnnotations()[cp.Annotation]

	return oldAnnotationVal != newAnnotationVal
}
