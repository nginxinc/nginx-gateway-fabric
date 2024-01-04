package state

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// stateChangedPredicate determines whether upsert and delete events constitute a change in state.
type stateChangedPredicate interface {
	// upsert returns true if the newObject changes state.
	upsert(oldObject, newObject client.Object) bool
	// delete returns true if the deletion of the object changes state.
	delete(object client.Object, nsname types.NamespacedName) bool
}

// funcPredicate applies the stateChanged function on upsert and delete. On upsert, the newObject is passed.
// Implements stateChangedPredicate.
type funcPredicate struct {
	stateChanged func(object client.Object, nsname types.NamespacedName) bool
}

func (f funcPredicate) upsert(_, newObject client.Object) bool {
	if newObject == nil {
		panic("New object cannot be nil")
	}

	return f.stateChanged(newObject, types.NamespacedName{
		Namespace: newObject.GetNamespace(),
		Name:      newObject.GetName(),
	})
}

func (f funcPredicate) delete(object client.Object, nsname types.NamespacedName) bool {
	return f.stateChanged(object, nsname)
}

// FIXME(kevin85421): We should remove this predicate and update changeTrackingUpdater once #1432 is merged.
type alwaysProcess struct{}

func (alwaysProcess) delete(_ client.Object, _ types.NamespacedName) bool { return true }

func (alwaysProcess) upsert(_, _ client.Object) bool { return true }

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

func (a annotationChangedPredicate) delete(_ client.Object, _ types.NamespacedName) bool { return true }
