package state

import (
	"context"
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceStore

// k8sServiceNameLabel is a Kubernetes label that holds the service name that owns the endpoint slice.
// Used to lookup endpoint slices for a given service.
const k8sServiceNameLabel = "kubernetes.io/service-name"

// Endpoint is the internal representation of a Kubernetes endpoint.
type Endpoint struct {
	Address string
	Port    int32
}

// ServiceStore stores services and can be queried for the endpoints of a service.
type ServiceStore interface {
	// Upsert upserts the service into the store.
	Upsert(svc *v1.Service)
	// Delete deletes the service from the store.
	Delete(nsname types.NamespacedName)
	// Resolve returns the endpoints for the service specified by its nsname and port.
	// If the service doesn't have endpoints, or it doesn't exist, resolve will return an error.
	Resolve(nsname types.NamespacedName, port int32) ([]Endpoint, error)
}

// NewServiceStore creates a new ServiceStore.
func NewServiceStore(k8sClient client.Client) *serviceStoreImpl {
	return &serviceStoreImpl{
		services:  make(map[string]*v1.Service),
		k8sClient: k8sClient,
	}
}

type serviceStoreImpl struct {
	services  map[string]*v1.Service
	k8sClient client.Client
}

func (s *serviceStoreImpl) Upsert(svc *v1.Service) {
	s.services[getResourceKey(&svc.ObjectMeta)] = svc
}

func (s *serviceStoreImpl) Delete(nsname types.NamespacedName) {
	delete(s.services, nsname.String())
}

func (s *serviceStoreImpl) Resolve(nsname types.NamespacedName, port int32) ([]Endpoint, error) {
	svc, exist := s.services[nsname.String()]
	if !exist {
		return nil, fmt.Errorf("service %s doesn't exist", nsname)
	}

	targetPort := getTargetPort(svc, port)
	if targetPort == 0 {
		return nil, fmt.Errorf("no matching target port for service %s and port %d", nsname, port)
	}

	endpoints, err := s.resolveEndpoints(svc, targetPort)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve endpoints for service %s: %w", nsname, err)
	}

	return endpoints, nil
}

func getTargetPort(svc *v1.Service, svcPort int32) int32 {
	for _, port := range svc.Spec.Ports {
		if port.Port == svcPort {
			return int32(port.TargetPort.IntValue())
		}
	}

	return 0
}

func (s *serviceStoreImpl) resolveEndpoints(svc *v1.Service, targetPort int32) ([]Endpoint, error) {
	var endpointSliceList discoveryV1.EndpointSliceList

	err := s.k8sClient.List(context.TODO(), &endpointSliceList, client.MatchingLabels{k8sServiceNameLabel: svc.Name})
	if err != nil {
		return nil, err
	}

	if len(endpointSliceList.Items) == 0 {
		return nil, errors.New("no endpoints found")
	}

	capacity := calculateEndpointSliceCapacity(endpointSliceList.Items, targetPort)

	if capacity == 0 {
		return nil, errors.New("no valid endpoints found")
	}

	endpoints := make([]Endpoint, 0, capacity)

	for _, eps := range endpointSliceList.Items {
		// FIXME(kate-osborn): only handling ipv4 addresses right now
		if eps.AddressType != discoveryV1.AddressTypeIPv4 {
			continue
		}

		if !targetPortExists(eps.Ports, targetPort) {
			continue
		}

		for _, endpoint := range eps.Endpoints {

			if !endpointReady(endpoint) {
				continue
			}

			for _, address := range endpoint.Addresses {
				ep := Endpoint{Address: address, Port: targetPort}
				endpoints = append(endpoints, ep)
			}
		}
	}

	return endpoints, nil
}

func calculateEndpointSliceCapacity(endpointSlices []discoveryV1.EndpointSlice, targetPort int32) (capacity int) {
	for _, es := range endpointSlices {
		if es.AddressType != discoveryV1.AddressTypeIPv4 {
			continue
		}

		if !targetPortExists(es.Ports, targetPort) {
			continue
		}

		for _, e := range es.Endpoints {
			if !endpointReady(e) {
				continue
			}
			capacity += len(e.Addresses)
		}
	}

	return
}

func endpointReady(endpoint discoveryV1.Endpoint) bool {
	ready := endpoint.Conditions.Ready
	return ready != nil && *ready
}

func targetPortExists(ports []discoveryV1.EndpointPort, targetPort int32) bool {
	for _, port := range ports {
		if port.Port == nil {
			continue
		}

		if *port.Port == targetPort {
			return true
		}
	}

	return false
}

func getResourceKey(meta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
