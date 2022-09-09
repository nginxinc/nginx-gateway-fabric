package resolver

import (
	"context"
	"fmt"

	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceResolver

// ServiceResolver resolves a Service and Service Port to a list of Endpoints.
// Returns an error if the Service or Service Port cannot be resolved.
type ServiceResolver interface {
	Resolve(ctx context.Context, svc *v1.Service, svcPort int32) ([]Endpoint, error)
}

// Endpoint is the internal representation of a Kubernetes endpoint.
type Endpoint struct {
	// Address is the IP address of the endpoint.
	Address string
	// Port is the port of the endpoint.
	Port int32
}

// ServiceResolverImpl implements ServiceResolver.
type ServiceResolverImpl struct {
	client client.Client
}

// NewServiceResolverImpl creates a new instance of a ServiceResolverImpl.
func NewServiceResolverImpl(client client.Client) *ServiceResolverImpl {
	return &ServiceResolverImpl{client: client}
}

// Resolve resolves a Service and Port to a list of Endpoints.
// Returns an error if the Service or Port cannot be resolved.
func (e *ServiceResolverImpl) Resolve(ctx context.Context, svc *v1.Service, port int32) ([]Endpoint, error) {
	if svc == nil {
		return nil, fmt.Errorf("cannot resolve a nil Service")
	}

	// We list EndpointSlices using the Service Name Index Field we added as an index to the EndpointSlice cache.
	// This allows us to perform a quick lookup of all EndpointSlices for a Service.
	var endpointSliceList discoveryV1.EndpointSliceList
	err := e.client.List(
		ctx,
		&endpointSliceList,
		client.MatchingFields{index.KubernetesServiceNameIndexField: svc.Name},
		client.InNamespace(svc.Namespace),
	)

	if err != nil || len(endpointSliceList.Items) == 0 {
		return nil, fmt.Errorf("no endpoints found for Service %s", client.ObjectKeyFromObject(svc))
	}

	return resolveEndpoints(svc, port, endpointSliceList)
}

func resolveEndpoints(
	svc *v1.Service,
	port int32,
	endpointSliceList discoveryV1.EndpointSliceList,
) ([]Endpoint, error) {
	svcPort, err := getServicePort(svc, port)
	if err != nil {
		return nil, err
	}

	filteredSlices := filterEndpointSliceList(endpointSliceList, svcPort)

	if len(filteredSlices) == 0 {
		svcNsName := client.ObjectKeyFromObject(svc)
		return nil, fmt.Errorf("no valid endpoints found for Service %s and port %+v", svcNsName, svcPort)
	}

	// Endpoints may be duplicated across multiple EndpointSlices.
	// Using a set to prevent returning duplicate endpoints.
	endpointSet := make(map[Endpoint]struct{})

	for _, eps := range filteredSlices {
		for _, endpoint := range eps.Endpoints {

			if !endpointReady(endpoint) {
				continue
			}

			// We don't check for a zero port value here because we are only working with EndpointSlices
			// that have a matching port.
			endpointPort := findPort(eps.Ports, svcPort)

			for _, address := range endpoint.Addresses {
				ep := Endpoint{Address: address, Port: endpointPort}
				endpointSet[ep] = struct{}{}
			}
		}
	}

	endpoints := make([]Endpoint, 0, len(endpointSet))
	for ep := range endpointSet {
		endpoints = append(endpoints, ep)
	}

	return endpoints, nil
}

func getServicePort(svc *v1.Service, port int32) (v1.ServicePort, error) {
	for _, p := range svc.Spec.Ports {
		if p.Port == port {
			return p, nil
		}
	}

	return v1.ServicePort{}, fmt.Errorf("no matching port for Service %s and port %d", svc.Name, port)
}

// getDefaultPort returns the default port for a ServicePort.
// This default port is used when the EndpointPort has a nil port which indicates all ports are valid.
// If the ServicePort has a non-zero integer TargetPort, the TargetPort integer value is returned.
// Otherwise, the ServicePort port value is returned.
func getDefaultPort(svcPort v1.ServicePort) int32 {
	switch svcPort.TargetPort.Type {
	case intstr.Int:
		if svcPort.TargetPort.IntVal != 0 {
			return svcPort.TargetPort.IntVal
		}
	}

	return svcPort.Port
}

func ignoreEndpointSlice(endpointSlice discoveryV1.EndpointSlice, port v1.ServicePort) bool {
	if endpointSlice.AddressType != discoveryV1.AddressTypeIPv4 {
		return true
	}

	// ignore endpoint slices that don't have a matching port.
	return findPort(endpointSlice.Ports, port) == 0
}

func endpointReady(endpoint discoveryV1.Endpoint) bool {
	ready := endpoint.Conditions.Ready
	return ready != nil && *ready
}

func filterEndpointSliceList(
	endpointSliceList discoveryV1.EndpointSliceList,
	port v1.ServicePort,
) []discoveryV1.EndpointSlice {
	filtered := make([]discoveryV1.EndpointSlice, 0, len(endpointSliceList.Items))

	for _, endpointSlice := range endpointSliceList.Items {
		if !ignoreEndpointSlice(endpointSlice, port) {
			filtered = append(filtered, endpointSlice)
		}
	}

	return filtered
}

// findPort locates the port in the slice of EndpointPort that matches the ServicePort name.
// The Kubernetes EndpointSlice controller handles matching the TargetPort of a ServicePort to the container port of
// an endpoint. All we have to do is find the port with the same name as the ServicePort.
// If a ServicePort is unnamed, then the EndpointPort will also be unnamed (empty string).
//
// If an EndpointPort port is nil -- indicating all ports are valid --
// the default port for the ServicePort is returned.
// If no matching port is found, 0 is returned.
func findPort(ports []discoveryV1.EndpointPort, svcPort v1.ServicePort) int32 {
	portName := svcPort.Name

	for _, p := range ports {

		if p.Port == nil {
			return getDefaultPort(svcPort)
		}

		if p.Name != nil && *p.Name == portName {
			return *p.Port
		}
	}

	return 0
}
