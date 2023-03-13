package graph

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func addBackendGroupsToRoutes(
	routes map[types.NamespacedName]*Route,
	services map[types.NamespacedName]*v1.Service,
) {
	for _, r := range routes {
		addBackendGroupsToRoute(r, services)
	}
}

// addBackendGroupsToRoute iterates over the rules of a route and adds BackendGroups to the rules.
// The route is modified in place.
// If a reference in a rule is invalid, the function will add a condition to the rule.
func addBackendGroupsToRoute(route *Route, services map[types.NamespacedName]*v1.Service) {
	if !route.Valid {
		return
	}

	for idx, rule := range route.Source.Spec.Rules {
		if !route.Rules[idx].ValidMatches {
			continue
		}
		if !route.Rules[idx].ValidFilters {
			continue
		}

		// zero backendRefs is OK. For example, a rule can include a redirect filter.
		if len(rule.BackendRefs) == 0 {
			continue
		}

		group := &route.Rules[idx].BackendGroup

		group.Backends = make([]BackendRef, 0, len(rule.BackendRefs))

		for refIdx, ref := range rule.BackendRefs {
			refPath := field.NewPath("spec").Child("rules").Index(idx).Child("backendRefs").Index(refIdx)

			backend, conds := createBackend(ref, route.Source.Namespace, services, refPath)

			group.Backends = append(group.Backends, backend)
			route.Conditions = append(route.Conditions, conds...)
		}
	}
}

func createBackend(
	ref v1beta1.HTTPBackendRef,
	sourceNamespace string,
	services map[types.NamespacedName]*v1.Service,
	refPath *field.Path,
) (BackendRef, []conditions.Condition) {
	// Data plane will handle invalid ref by responding with 500.
	// Because of that, we always need to add a BackendRef to group.Backends, even if the ref is invalid.
	// Additionally, we always calculate the weight, even if it is invalid.
	weight := int32(1)
	if ref.Weight != nil {
		if validateWeight(*ref.Weight) != nil {
			// We don't need to add a condition because validateHTTPBackendRef will do that.
			weight = 0 // 0 will get no traffic
		} else {
			weight = *ref.Weight
		}
	}

	var backend BackendRef

	valid, cond := validateHTTPBackendRef(ref, sourceNamespace, refPath)
	if !valid {
		backend = BackendRef{
			Weight: weight,
			Valid:  false,
		}

		return backend, []conditions.Condition{cond}
	}

	svc, port, err := getServiceAndPortFromRef(ref.BackendRef, sourceNamespace, services, refPath)
	if err != nil {
		backend = BackendRef{
			Weight: weight,
			Valid:  false,
		}

		cond := conditions.NewRouteBackendRefRefBackendNotFound(err.Error())
		return backend, []conditions.Condition{cond}
	}

	backend = BackendRef{
		Name:   fmt.Sprintf("%s_%s_%d", svc.Namespace, svc.Name, port),
		Svc:    svc,
		Port:   port,
		Valid:  true,
		Weight: weight,
	}

	return backend, nil
}

func getServiceAndPortFromRef(
	ref v1beta1.BackendRef,
	routeNamespace string,
	services map[types.NamespacedName]*v1.Service,
	refPath *field.Path,
) (*v1.Service, int32, error) {
	svcNsName := types.NamespacedName{Name: string(ref.Name), Namespace: routeNamespace}

	svc, ok := services[svcNsName]
	if !ok {
		return nil, 0, field.NotFound(refPath.Child("name"), ref.Name)
	}

	// safe to dereference port here because we already validated that the port is not nil.
	return svc, int32(*ref.Port), nil
}

func validateHTTPBackendRef(
	ref v1beta1.HTTPBackendRef,
	routeNs string,
	path *field.Path,
) (valid bool, cond conditions.Condition) {
	// Because all errors cause the same condition but different reasons, we return as soon as we find an error

	if len(ref.Filters) > 0 {
		valErr := field.TooMany(path.Child("filters"), len(ref.Filters), 0)
		return false, conditions.NewRouteBackendRefUnsupportedValue(valErr.Error())
	}

	return validateBackendRef(ref.BackendRef, routeNs, path)
}

func validateBackendRef(
	ref v1beta1.BackendRef,
	routeNs string,
	path *field.Path,
) (valid bool, cond conditions.Condition) {
	// Because all errors cause same condition but different reasons, we return as soon as we find an error

	if ref.Group != nil && !(*ref.Group == "core" || *ref.Group == "") {
		valErr := field.NotSupported(path.Child("group"), *ref.Group, []string{"core", ""})
		return false, conditions.NewRouteBackendRefUnsupportedValue(valErr.Error())
	}

	if ref.Kind != nil && *ref.Kind != "Service" {
		valErr := field.NotSupported(path.Child("kind"), *ref.Kind, []string{"Service"})
		return false, conditions.NewRouteBackendRefInvalidKind(valErr.Error())
	}

	// no need to validate ref.Name

	if ref.Namespace != nil && string(*ref.Namespace) != routeNs {
		valErr := field.Invalid(path.Child("namespace"), *ref.Namespace, "cross-namespace routing is not permitted")
		return false, conditions.NewRouteBackendRefRefNotPermitted(valErr.Error())
	}

	// The imported Webhook validation ensures ref.Port is set
	// any value is OK
	// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

	if ref.Weight != nil {
		if err := validateWeight(*ref.Weight); err != nil {
			valErr := field.Invalid(path.Child("weight"), *ref.Weight, err.Error())
			return false, conditions.NewRouteBackendRefUnsupportedValue(valErr.Error())
		}
	}

	return true, conditions.Condition{}
}

func validateWeight(weight int32) error {
	const (
		minWeight = 0
		maxWeight = 1_000_000
	)

	if weight < minWeight || weight > maxWeight {
		return fmt.Errorf("must be in the range [%d, %d]", minWeight, maxWeight)
	}

	return nil
}
