package sdk

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Implementation interface {
	Upsert(obj client.Object)
	Remove(nsname types.NamespacedName)
}
