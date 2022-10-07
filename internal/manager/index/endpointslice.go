package index

import (
	"fmt"

	discoveryV1 "k8s.io/api/discovery/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// KubernetesServiceNameIndexField is the name of the Index Field used to index EndpointSlices by their service
	// owners.
	KubernetesServiceNameIndexField = "k8sServiceName"
	// KubernetesServiceNameLabel is the label used to identify the Kubernetes service name on an EndpointSlice.
	KubernetesServiceNameLabel = "kubernetes.io/service-name"
)

// CreateEndpointSliceFieldIndices creates a FieldIndices map for the EndpointSlice resource.
func CreateEndpointSliceFieldIndices() FieldIndices {
	return FieldIndices{
		KubernetesServiceNameIndexField: serviceNameIndexFunc,
	}
}

// serviceNameIndexFunc is a client.IndexerFunc that parses a Kubernetes object and returns the value of the
// Kubernetes service-name label.
// Used to index EndpointSlices by their service owners.
func serviceNameIndexFunc(obj client.Object) []string {
	slice, ok := obj.(*discoveryV1.EndpointSlice)
	if !ok {
		panic(fmt.Sprintf("expected an EndpointSlice; got %T", obj))
	}

	name := GetServiceNameFromEndpointSlice(slice)
	if name == "" {
		return nil
	}

	return []string{name}
}

// GetServiceNameFromEndpointSlice returns the value of the Kubernetes service-name label from an EndpointSlice.
func GetServiceNameFromEndpointSlice(slice *discoveryV1.EndpointSlice) string {
	if slice.Labels == nil {
		return ""
	}

	return slice.Labels[KubernetesServiceNameLabel]
}
