package state

import (
	"errors"
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// addBackendGroupsToRoutes iterates over the routes and adds BackendGroups to the routes.
// The routes are modified in place.
// If a backend ref is invalid it will store an error message in the BackendGroup.Errors field.
// A backend ref is invalid if:
// - the Kind is not Service
// - the Namespace is not the same as the HTTPRoute namespace
// - the Port is nil
func addBackendGroupsToRoutes(
	routes map[types.NamespacedName]*route,
	services map[types.NamespacedName]*v1.Service,
) {
	for _, r := range routes {
		r.BackendGroups = make([]BackendGroup, len(r.Source.Spec.Rules))

		for idx, rule := range r.Source.Spec.Rules {

			group := BackendGroup{
				Source:  client.ObjectKeyFromObject(r.Source),
				RuleIdx: idx,
			}

			if len(rule.BackendRefs) == 0 {

				r.BackendGroups[idx] = group
				continue
			}

			group.Errors = make([]string, 0, len(rule.BackendRefs))
			group.Backends = make([]BackendRef, 0, len(rule.BackendRefs))

			for _, ref := range rule.BackendRefs {

				weight := int32(1)
				if ref.Weight != nil {
					weight = *ref.Weight
				}

				svc, port, err := getServiceAndPortFromRef(ref.BackendRef, r.Source.Namespace, services)
				if err != nil {
					group.Backends = append(group.Backends, BackendRef{Weight: weight})

					group.Errors = append(group.Errors, err.Error())

					continue
				}

				group.Backends = append(group.Backends, BackendRef{
					Name:   fmt.Sprintf("%s_%s_%d", svc.Namespace, svc.Name, port),
					Svc:    svc,
					Port:   port,
					Valid:  true,
					Weight: weight,
				})
			}

			r.BackendGroups[idx] = group
		}
	}
}

func getServiceAndPortFromRef(
	ref v1beta1.BackendRef,
	routeNamespace string,
	services map[types.NamespacedName]*v1.Service,
) (*v1.Service, int32, error) {
	err := validateBackendRef(ref, routeNamespace)
	if err != nil {
		return nil, 0, err
	}

	svcNsName := types.NamespacedName{Name: string(ref.Name), Namespace: routeNamespace}

	svc, ok := services[svcNsName]
	if !ok {
		return nil, 0, fmt.Errorf("the Service %s does not exist", svcNsName)
	}

	// safe to dereference port here because we already validated that the port is not nil.
	return svc, int32(*ref.Port), nil
}

func validateBackendRef(ref v1beta1.BackendRef, routeNs string) error {
	if ref.Kind != nil && *ref.Kind != "Service" {
		return fmt.Errorf("the Kind must be Service; got %s", *ref.Kind)
	}

	if ref.Namespace != nil && string(*ref.Namespace) != routeNs {
		return fmt.Errorf("cross-namespace routing is not permitted; namespace %s does not match the HTTPRoute namespace %s", *ref.Namespace, routeNs)
	}

	if ref.Port == nil {
		return errors.New("port is missing")
	}

	return nil
}
