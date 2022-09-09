package sdk

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Implementation[T ObjectConstraint] interface {
	Upsert(obj T)
	Remove(nsname types.NamespacedName)
}

type ObjectConstraint interface {
	client.Object
}
