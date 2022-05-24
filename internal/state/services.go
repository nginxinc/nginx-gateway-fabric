package state

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . ServiceStore

// ServiceStore stores services and can be queried for the cluster IP of a service.
type ServiceStore interface {
	// Upsert upserts the service into the store.
	Upsert(svc *v1.Service)
	// Delete deletes the service from the store.
	Delete(nsname types.NamespacedName)
	// Resolve returns the cluster IP  the service specified by its namespace and name.
	// If the service doesn't have a cluster IP or it doesn't exist, resolve will return an error.
	// FIXME(pleshakov): later, we will start using the Endpoints rather than cluster IPs.
	Resolve(nsname types.NamespacedName) (string, error)
}

// NewServiceStore creates a new ServiceStore.
func NewServiceStore() ServiceStore {
	return &serviceStoreImpl{
		services: make(map[string]*v1.Service),
	}
}

type serviceStoreImpl struct {
	services map[string]*v1.Service
}

func (s *serviceStoreImpl) Upsert(svc *v1.Service) {
	s.services[getResourceKey(&svc.ObjectMeta)] = svc
}

func (s *serviceStoreImpl) Delete(nsname types.NamespacedName) {
	delete(s.services, nsname.String())
}

func (s *serviceStoreImpl) Resolve(nsname types.NamespacedName) (string, error) {
	svc, exist := s.services[nsname.String()]
	if !exist {
		return "", fmt.Errorf("service %s doesn't exist", nsname.String())
	}

	if svc.Spec.ClusterIP == "" || svc.Spec.ClusterIP == "None" {
		return "", fmt.Errorf("service %s doesn't have ClusterIP", nsname.String())
	}

	return svc.Spec.ClusterIP, nil
}

func getResourceKey(meta *metav1.ObjectMeta) string {
	return fmt.Sprintf("%s/%s", meta.Namespace, meta.Name)
}
