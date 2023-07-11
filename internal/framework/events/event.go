package events

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EventBatch is a batch of events to be handled at once.
type EventBatch []interface{}

// UpsertEvent represents upserting a resource.
type UpsertEvent struct {
	// Resource is the resource that is being upserted.
	Resource client.Object
}

// DeleteEvent representing deleting a resource.
type DeleteEvent struct {
	// Type is the resource type. For example, if the event is for *v1beta1.HTTPRoute, pass &v1beta1.HTTPRoute{} as Type.
	Type client.Object
	// NamespacedName is the namespace & name of the deleted resource.
	NamespacedName types.NamespacedName
}
