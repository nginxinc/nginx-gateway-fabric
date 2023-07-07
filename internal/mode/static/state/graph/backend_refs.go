package graph

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
)

// BackendRef is an internal representation of a backendRef in an HTTPRoute.
type BackendRef struct {
	// Svc is the service referenced by the backendRef.
	Svc *v1.Service
	// Port is the port of the backendRef.
	Port int32
	// Weight is the weight of the backendRef.
	Weight int32
	// Valid indicates whether the backendRef is valid.
	Valid bool
}

// ServicePortReference returns a string representation for the service and port that is referenced by the BackendRef.
func (b BackendRef) ServicePortReference() string {
	if b.Svc == nil {
		return ""
	}
	return fmt.Sprintf("%s_%s_%d", b.Svc.Namespace, b.Svc.Name, b.Port)
}

func addBackendRefsToRouteRules(
	routes map[types.NamespacedName]*Route,
	refGrantResolver *referenceGrantResolver,
	services map[types.NamespacedName]*v1.Service,
) {
	for _, r := range routes {
		addBackendRefsToRules(r, refGrantResolver, services)
	}
}

// addBackendRefsToRules iterates over the rules of a route and adds a list of BackendRef to each rule.
// The route is modified in place.
// If a reference in a rule is invalid, the function will add a condition to the rule.
func addBackendRefsToRules(
	route *Route,
	refGrantResolver *referenceGrantResolver,
	services map[types.NamespacedName]*v1.Service,
) {
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

		backendRefs := make([]BackendRef, 0, len(rule.BackendRefs))

		for refIdx, ref := range rule.BackendRefs {
			refPath := field.NewPath("spec").Child("rules").Index(idx).Child("backendRefs").Index(refIdx)

			ref, cond := createBackendRef(ref, route.Source.Namespace, refGrantResolver, services, refPath)

			backendRefs = append(backendRefs, ref)
			if cond != nil {
				route.Conditions = append(route.Conditions, *cond)
			}
		}

		route.Rules[idx].BackendRefs = backendRefs
	}
}

func createBackendRef(
	ref v1beta1.HTTPBackendRef,
	sourceNamespace string,
	refGrantResolver *referenceGrantResolver,
	services map[types.NamespacedName]*v1.Service,
	refPath *field.Path,
) (BackendRef, *conditions.Condition) {
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

	var backendRef BackendRef

	valid, cond := validateHTTPBackendRef(ref, sourceNamespace, refGrantResolver, refPath)
	if !valid {
		backendRef = BackendRef{
			Weight: weight,
			Valid:  false,
		}

		return backendRef, &cond
	}

	svc, port, err := getServiceAndPortFromRef(ref.BackendRef, sourceNamespace, services, refPath)
	if err != nil {
		backendRef = BackendRef{
			Weight: weight,
			Valid:  false,
		}

		cond := conditions.NewRouteBackendRefRefBackendNotFound(err.Error())
		return backendRef, &cond
	}

	backendRef = BackendRef{
		Svc:    svc,
		Port:   port,
		Valid:  true,
		Weight: weight,
	}

	return backendRef, nil
}

func getServiceAndPortFromRef(
	ref v1beta1.BackendRef,
	routeNamespace string,
	services map[types.NamespacedName]*v1.Service,
	refPath *field.Path,
) (*v1.Service, int32, error) {
	ns := routeNamespace
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	svcNsName := types.NamespacedName{Name: string(ref.Name), Namespace: ns}

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
	refGrantResolver *referenceGrantResolver,
	path *field.Path,
) (valid bool, cond conditions.Condition) {
	// Because all errors cause the same condition but different reasons, we return as soon as we find an error

	if len(ref.Filters) > 0 {
		valErr := field.TooMany(path.Child("filters"), len(ref.Filters), 0)
		return false, conditions.NewRouteBackendRefUnsupportedValue(valErr.Error())
	}

	return validateBackendRef(ref.BackendRef, routeNs, refGrantResolver, path)
}

func validateBackendRef(
	ref v1beta1.BackendRef,
	routeNs string,
	refGrantResolver *referenceGrantResolver,
	path *field.Path,
) (valid bool, cond conditions.Condition) {
	// Because all errors cause same condition but different reasons, we return as soon as we find an error

	if ref.Group != nil && !(*ref.Group == "core" || *ref.Group == "") {
		valErr := field.NotSupported(path.Child("group"), *ref.Group, []string{"core", ""})
		return false, conditions.NewRouteBackendRefInvalidKind(valErr.Error())
	}

	if ref.Kind != nil && *ref.Kind != "Service" {
		valErr := field.NotSupported(path.Child("kind"), *ref.Kind, []string{"Service"})
		return false, conditions.NewRouteBackendRefInvalidKind(valErr.Error())
	}

	// no need to validate ref.Name

	if ref.Namespace != nil && string(*ref.Namespace) != routeNs {
		refNsName := types.NamespacedName{Namespace: string(*ref.Namespace), Name: string(ref.Name)}

		if !refGrantResolver.refAllowed(toService(refNsName), fromHTTPRoute(routeNs)) {
			msg := fmt.Sprintf("Backend ref to Service %s not permitted by any ReferenceGrant", refNsName)

			return false, conditions.NewRouteBackendRefRefNotPermitted(msg)
		}
	}

	if ref.Port == nil {
		panicForBrokenWebhookAssumption(fmt.Errorf("port is nil for ref %q", ref.Name))
	}

	// any value of port is OK

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
