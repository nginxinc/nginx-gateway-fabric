package state

import "sigs.k8s.io/controller-runtime/pkg/client"

// stateChangedPredicate determines whether upsert and delete events constitute a change in state.
type stateChangedPredicate interface {
	// upsert returns true if the newObject changes state.
	upsert(oldObject, newObject client.Object) bool
	// delete returns true if the deletion of the object changes state.
	delete(object client.Object) bool
}

// funcs is a function that implements stateChangedPredicate.
type funcs struct {
	upsertStateChangeFunc func(oldObject, newObject client.Object) bool
	deleteStateChangeFunc func(object client.Object) bool
}

func (f funcs) upsert(oldObject, newObject client.Object) bool {
	return f.upsertStateChangeFunc(oldObject, newObject)
}

func (f funcs) delete(object client.Object) bool {
	return f.deleteStateChangeFunc(object)
}

// newStateChangedPredicateFuncs returns a predicate funcs that applies the given function on calls to upsert and
// delete.
func newStateChangedPredicateFuncs(stateChangedFunc func(object client.Object) bool) funcs {
	return funcs{
		upsertStateChangeFunc: func(oldObject, newObject client.Object) bool {
			return stateChangedFunc(newObject)
		},
		deleteStateChangeFunc: func(object client.Object) bool {
			return stateChangedFunc(object)
		},
	}
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
