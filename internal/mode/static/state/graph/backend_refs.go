package graph

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

// BackendRef is an internal representation of a backendRef in an HTTPRoute.
type BackendRef struct {
	// BackendTLSPolicy is the BackendTLSPolicy of the Service which is referenced by the backendRef.
	BackendTLSPolicy *gatewayv1alpha2.BackendTLSPolicy
	// SvcNsName is the NamespacedName of the Service referenced by the backendRef.
	SvcNsName types.NamespacedName
	// ServicePort is the ServicePort of the Service which is referenced by the backendRef.
	ServicePort v1.ServicePort
	// Weight is the weight of the backendRef.
	Weight int32
	// Valid indicates whether the backendRef is valid.
	// No configuration should be generated for an invalid BackendRef.
	Valid bool
}

// ServicePortReference returns a string representation for the service and port that is referenced by the BackendRef.
func (b BackendRef) ServicePortReference() string {
	if !b.Valid {
		return ""
	}
	return fmt.Sprintf("%s_%s_%d", b.SvcNsName.Namespace, b.SvcNsName.Name, b.ServicePort.Port)
}

func addBackendRefsToRouteRules(
	routes map[types.NamespacedName]*Route,
	refGrantResolver *referenceGrantResolver,
	services map[types.NamespacedName]*v1.Service,
	backendTLSPolicies map[types.NamespacedName]*gatewayv1alpha2.BackendTLSPolicy,
	configMapResolver *configMapResolver,
) {
	for _, r := range routes {
		addBackendRefsToRules(r, refGrantResolver, services, backendTLSPolicies, configMapResolver)
	}
}

// addBackendRefsToRules iterates over the rules of a route and adds a list of BackendRef to each rule.
// The route is modified in place.
// If a reference in a rule is invalid, the function will add a condition to the rule.
func addBackendRefsToRules(
	route *Route,
	refGrantResolver *referenceGrantResolver,
	services map[types.NamespacedName]*v1.Service,
	backendTLSPolicies map[types.NamespacedName]*gatewayv1alpha2.BackendTLSPolicy,
	configMapResolver *configMapResolver,
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

			ref, cond := createBackendRef(
				ref,
				route.Source.Namespace,
				refGrantResolver,
				services,
				refPath,
				backendTLSPolicies,
				configMapResolver,
			)

			backendRefs = append(backendRefs, ref)
			if cond != nil {
				route.Conditions = append(route.Conditions, *cond)
			}
		}

		route.Rules[idx].BackendRefs = backendRefs
	}
}

func createBackendRef(
	ref gatewayv1.HTTPBackendRef,
	sourceNamespace string,
	refGrantResolver *referenceGrantResolver,
	services map[types.NamespacedName]*v1.Service,
	refPath *field.Path,
	backendTLSPolicies map[types.NamespacedName]*gatewayv1alpha2.BackendTLSPolicy,
	configMapResolver *configMapResolver,
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

	svcNsName, svcPort, err := getServiceAndPortFromRef(ref.BackendRef, sourceNamespace, services, refPath)
	if err != nil {
		backendRef = BackendRef{
			SvcNsName:   svcNsName,
			ServicePort: svcPort,
			Weight:      weight,
			Valid:       false,
		}

		cond := staticConds.NewRouteBackendRefRefBackendNotFound(err.Error())
		return backendRef, &cond
	}

	backendTLSPolicy, err := findBackendTLSPolicyForService(
		backendTLSPolicies,
		ref,
		sourceNamespace,
		configMapResolver,
	)
	if err != nil {
		backendRef = BackendRef{
			SvcNsName:   svcNsName,
			ServicePort: svcPort,
			Weight:      weight,
			Valid:       false,
		}

		cond := staticConds.NewRouteBackendRefUnsupportedValue(err.Error())
		return backendRef, &cond
	}

	backendRef = BackendRef{
		SvcNsName:        svcNsName,
		BackendTLSPolicy: backendTLSPolicy,
		ServicePort:      svcPort,
		Valid:            true,
		Weight:           weight,
	}

	return backendRef, nil
}

func findBackendTLSPolicyForService(
	backendTLSPolicies map[types.NamespacedName]*gatewayv1alpha2.BackendTLSPolicy,
	ref gatewayv1.HTTPBackendRef,
	routeNamespace string,
	configMapResolver *configMapResolver,
) (*gatewayv1alpha2.BackendTLSPolicy, error) {
	var beTLSPolicy *gatewayv1alpha2.BackendTLSPolicy
	refNs := routeNamespace
	if ref.Namespace != nil {
		refNs = string(*ref.Namespace)
	}

	for _, btp := range backendTLSPolicies {
		btpNs := btp.Namespace
		if btp.Spec.TargetRef.Namespace != nil {
			btpNs = string(*btp.Spec.TargetRef.Namespace)
		}
		if btp.Spec.TargetRef.Name == ref.Name && btpNs == refNs {
			beTLSPolicy = btp
			break
		}
	}

	if beTLSPolicy == nil {
		return nil, nil
	}

	err := validateBackendTLSPolicy(beTLSPolicy, configMapResolver)

	return beTLSPolicy, err
}

func validateBackendTLSPolicy(
	btp *gatewayv1alpha2.BackendTLSPolicy,
	configMapResolver *configMapResolver,
) error {
	if err := validateBackendTLSHostname(btp); err != nil {
		return err
	}
	if btp.Spec.TLS.CACertRefs != nil && len(btp.Spec.TLS.CACertRefs) > 0 {
		return validateBackendTLSCACertRef(btp, configMapResolver)
	} else if btp.Spec.TLS.WellKnownCACerts != nil {
		return validateBackendTLSWellKnownCACerts(btp)
	}
	return fmt.Errorf("must specify either CACertRefs or WellKnownCACerts")
}

func validateBackendTLSHostname(btp *gatewayv1alpha2.BackendTLSPolicy) error {
	h := string(btp.Spec.TLS.Hostname)

	if err := validateHostname(h); err != nil {
		path := field.NewPath("tls.hostname")
		valErr := field.Invalid(path, btp.Spec.TLS.Hostname, err.Error())
		return valErr
	}
	return nil
}

func validateBackendTLSCACertRef(btp *gatewayv1alpha2.BackendTLSPolicy, configMapResolver *configMapResolver) error {
	if len(btp.Spec.TLS.CACertRefs) != 1 {
		path := field.NewPath("tls.cacertrefs")
		valErr := field.TooMany(path, len(btp.Spec.TLS.CACertRefs), 1)
		return valErr
	}
	if btp.Spec.TLS.CACertRefs[0].Kind != "ConfigMap" {
		path := field.NewPath("tls.cacertrefs[0].kind")
		valErr := field.NotSupported(path, btp.Spec.TLS.CACertRefs[0].Kind, []string{"ConfigMap"})
		return valErr
	}
	if btp.Spec.TLS.CACertRefs[0].Group != "" && btp.Spec.TLS.CACertRefs[0].Group != "core" {
		path := field.NewPath("tls.cacertrefs[0].group")
		valErr := field.NotSupported(path, btp.Spec.TLS.CACertRefs[0].Kind, []string{"", "core"})
		return valErr
	}
	nsName := types.NamespacedName{Namespace: btp.Namespace, Name: string(btp.Spec.TLS.CACertRefs[0].Name)}
	if err := configMapResolver.resolve(nsName); err != nil {
		path := field.NewPath("tls.cacertrefs[0]")
		return field.Invalid(path, btp.Spec.TLS.CACertRefs[0], err.Error())
	}
	return nil
}

func validateBackendTLSWellKnownCACerts(btp *gatewayv1alpha2.BackendTLSPolicy) error {
	if *btp.Spec.TLS.WellKnownCACerts != gatewayv1alpha2.WellKnownCACertSystem {
		path := field.NewPath("tls.wellknowncacerts")
		return field.Invalid(path, btp.Spec.TLS.WellKnownCACerts, "unsupported value")
	}
	return nil
}

// getServiceAndPortFromRef extracts the NamespacedName of the Service and the port from a BackendRef.
// It can return an error and an empty v1.ServicePort in two cases:
// 1. The Service referenced from the BackendRef does not exist in the cluster/state.
// 2. The Port on the BackendRef does not match any of the ServicePorts on the Service.
func getServiceAndPortFromRef(
	ref gatewayv1.BackendRef,
	routeNamespace string,
	services map[types.NamespacedName]*v1.Service,
	refPath *field.Path,
) (types.NamespacedName, v1.ServicePort, error) {
	ns := routeNamespace
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	svcNsName := types.NamespacedName{Name: string(ref.Name), Namespace: ns}

	svc, ok := services[svcNsName]
	if !ok {
		return svcNsName, v1.ServicePort{}, field.NotFound(refPath.Child("name"), ref.Name)
	}

	// safe to dereference port here because we already validated that the port is not nil in validateBackendRef.
	svcPort, err := getServicePort(svc, int32(*ref.Port))
	if err != nil {
		return svcNsName, v1.ServicePort{}, err
	}

	return svcNsName, svcPort, nil
}

func validateHTTPBackendRef(
	ref gatewayv1.HTTPBackendRef,
	routeNs string,
	refGrantResolver *referenceGrantResolver,
	path *field.Path,
) (valid bool, cond conditions.Condition) {
	// Because all errors cause the same condition but different reasons, we return as soon as we find an error

	if len(ref.Filters) > 0 {
		valErr := field.TooMany(path.Child("filters"), len(ref.Filters), 0)
		return false, staticConds.NewRouteBackendRefUnsupportedValue(valErr.Error())
	}

	return validateBackendRef(ref.BackendRef, routeNs, refGrantResolver, path)
}

func validateBackendRef(
	ref gatewayv1.BackendRef,
	routeNs string,
	refGrantResolver *referenceGrantResolver,
	path *field.Path,
) (valid bool, cond conditions.Condition) {
	// Because all errors cause same condition but different reasons, we return as soon as we find an error

	if ref.Group != nil && !(*ref.Group == "core" || *ref.Group == "") {
		valErr := field.NotSupported(path.Child("group"), *ref.Group, []string{"core", ""})
		return false, staticConds.NewRouteBackendRefInvalidKind(valErr.Error())
	}

	if ref.Kind != nil && *ref.Kind != "Service" {
		valErr := field.NotSupported(path.Child("kind"), *ref.Kind, []string{"Service"})
		return false, staticConds.NewRouteBackendRefInvalidKind(valErr.Error())
	}

	// no need to validate ref.Name

	if ref.Namespace != nil && string(*ref.Namespace) != routeNs {
		refNsName := types.NamespacedName{Namespace: string(*ref.Namespace), Name: string(ref.Name)}

		if !refGrantResolver.refAllowed(toService(refNsName), fromHTTPRoute(routeNs)) {
			msg := fmt.Sprintf("Backend ref to Service %s not permitted by any ReferenceGrant", refNsName)

			return false, staticConds.NewRouteBackendRefRefNotPermitted(msg)
		}
	}

	if ref.Port == nil {
		panicForBrokenWebhookAssumption(fmt.Errorf("port is nil for ref %q", ref.Name))
	}

	// any value of port is OK

	if ref.Weight != nil {
		if err := validateWeight(*ref.Weight); err != nil {
			valErr := field.Invalid(path.Child("weight"), *ref.Weight, err.Error())
			return false, staticConds.NewRouteBackendRefUnsupportedValue(valErr.Error())
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

func getServicePort(svc *v1.Service, port int32) (v1.ServicePort, error) {
	for _, p := range svc.Spec.Ports {
		if p.Port == port {
			return p, nil
		}
	}

	return v1.ServicePort{}, fmt.Errorf("no matching port for Service %s and port %d", svc.Name, port)
}
