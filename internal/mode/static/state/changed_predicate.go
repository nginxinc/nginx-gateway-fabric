package state

import "sigs.k8s.io/controller-runtime/pkg/client"

// stateChangedPredicate determines whether upsert and delete events constitute a change in state.
type stateChangedPredicate interface {
	// upsert returns true if the newObject changes state.
	upsert(oldObject, newObject client.Object) bool
	// delete returns true if the deletion of the object changes state.
	delete(object client.Object) bool
}

// funcPredicate applies the stateChanged function on upsert and delete. On upsert, the newObject is passed.
// Implements stateChangedPredicate.
type funcPredicate struct {
	stateChanged func(object client.Object) bool
}

func (f funcPredicate) upsert(_, newObject client.Object) bool {
	return f.stateChanged(newObject)
}

func (f funcPredicate) delete(object client.Object) bool {
	return f.stateChanged(object)
}

// generationChangedPredicate implements stateChangedPredicate based on the generation of the object.
// This predicate will return true on upsert if the object's generation has changed.
// It always returns true on delete.
type generationChangedPredicate struct{}

func (generationChangedPredicate) delete(_ client.Object) bool { return true }

func (generationChangedPredicate) upsert(oldObject, newObject client.Object) bool {
	if oldObject == nil {
		return true
	}

	if newObject == nil {
		panic("Cannot determine if generation has changed on upsert because new object is nil")
	}

	return newObject.GetGeneration() != oldObject.GetGeneration()
}

// annotationChangedPredicate implements stateChangedPredicate based on the value of the annotation provided.
// This predicate will return true on upsert if the annotation's value has changed.
// It always returns true on delete.
type annotationChangedPredicate struct {
	annotation string
}

func (a annotationChangedPredicate) upsert(oldObject, newObject client.Object) bool {
	if oldObject == nil {
		return true
	}

	if newObject == nil {
		panic("Cannot determine if annotation has changed on upsert because new object is nil")
	}

	oldAnnotation := oldObject.GetAnnotations()[a.annotation]
	newAnnotation := newObject.GetAnnotations()[a.annotation]

	return oldAnnotation != newAnnotation
}

func (a annotationChangedPredicate) delete(_ client.Object) bool { return true }
