package graph

import (
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

func buildTLSRoute(
	gtr *v1alpha2.TLSRoute,
	gatewayNsNames []types.NamespacedName,
	services map[types.NamespacedName]*apiv1.Service,
	npCfg *NginxProxy,
) *L4Route {
	r := &L4Route{
		Source: gtr,
	}

	sectionNameRefs, err := buildSectionNameRefs(gtr.Spec.ParentRefs, gtr.Namespace, gatewayNsNames)
	if err != nil {
		r.Valid = false

		return r
	}
	// route doesn't belong to any of the Gateways
	if len(sectionNameRefs) == 0 {
		return nil
	}
	r.ParentRefs = sectionNameRefs

	if err := validateHostnames(
		gtr.Spec.Hostnames,
		field.NewPath("spec").Child("hostnames"),
	); err != nil {
		r.Valid = false
		r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedValue(err.Error()))
		return r
	}

	r.Spec.Hostnames = gtr.Spec.Hostnames

	if len(gtr.Spec.Rules) != 1 || len(gtr.Spec.Rules[0].BackendRefs) != 1 {
		r.Valid = false
		cond := staticConds.NewRouteBackendRefUnsupportedValue(
			"Must have exactly one Rule and BackendRef",
		)
		r.Conditions = append(r.Conditions, cond)
		return r
	}

	cond, br := validateBackendRefTLSRoute(gtr, services, npCfg)

	r.Spec.BackendRef = br
	r.Valid = true
	r.Attachable = true

	if cond != nil {
		r.Conditions = append(r.Conditions, *cond)
		r.Valid = false
	}

	return r
}

func validateBackendRefTLSRoute(gtr *v1alpha2.TLSRoute,
	services map[types.NamespacedName]*apiv1.Service,
	npCfg *NginxProxy,
) (*conditions.Condition, BackendRef) {
	// Length of BackendRefs and Rules is guaranteed to be one due to earlier check in buildTLSRoute
	refPath := field.NewPath("spec").Child("rules").Index(0).Child("backendRefs").Index(0)

	ref := gtr.Spec.Rules[0].BackendRefs[0]

	ns := gtr.Namespace
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	svcNsName := types.NamespacedName{
		Namespace: ns,
		Name:      string(gtr.Spec.Rules[0].BackendRefs[0].Name),
	}

	backendRef := BackendRef{
		Valid: true,
	}
	var cond *conditions.Condition

	if ref.Port == nil {
		valErr := field.Required(refPath.Child("port"), "port cannot be nil")
		backendRef.Valid = false
		cond = helpers.GetPointer(staticConds.NewRouteBackendRefUnsupportedValue(valErr.Error()))

		return cond, backendRef
	}

	svcIPFamily, svcPort, err := getIPFamilyAndPortFromRef(
		ref,
		svcNsName,
		services,
		refPath,
	)

	backendRef.ServicePort = svcPort
	backendRef.SvcNsName = svcNsName

	if err != nil {
		backendRef.Valid = false
		cond = helpers.GetPointer(staticConds.NewRouteBackendRefRefBackendNotFound(err.Error()))
	} else if err := verifyIPFamily(npCfg, svcIPFamily); err != nil {
		backendRef.Valid = false
		cond = helpers.GetPointer(staticConds.NewRouteInvalidIPFamily(err.Error()))
	} else if ref.Group != nil && !(*ref.Group == "core" || *ref.Group == "") {
		valErr := field.NotSupported(refPath.Child("group"), *ref.Group, []string{"core", ""})
		backendRef.Valid = false
		cond = helpers.GetPointer(staticConds.NewRouteBackendRefInvalidKind(valErr.Error()))
	} else if ref.Kind != nil && *ref.Kind != "Service" {
		valErr := field.NotSupported(refPath.Child("kind"), *ref.Kind, []string{"Service"})
		backendRef.Valid = false
		cond = helpers.GetPointer(staticConds.NewRouteBackendRefInvalidKind(valErr.Error()))
	} else if ref.Namespace != nil && string(*ref.Namespace) != gtr.Namespace {
		msg := "Cross-namespace routing is not supported"
		backendRef.Valid = false
		cond = helpers.GetPointer(staticConds.NewRouteBackendRefUnsupportedValue(msg))
	}
	// FIXME(sarthyparty): Add check for invalid weights, we removed checks to pass the conformance test
	return cond, backendRef
}
