package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

// GRPCRoute represents a GRPCRoute.
type GRPCRoute struct {
	// Source is the source resource of the Route.
	Source *v1alpha2.GRPCRoute
	// ParentRefs includes ParentRefs with NGF Gateways only.
	ParentRefs []ParentRef
	// Conditions include Conditions for the GRPCRoute.
	Conditions []conditions.Condition
	// Rules include Rules for the HTTPRoute. Each Rule[i] corresponds to the ith HTTPRouteRule.
	// If the Route is invalid, this field is nil
	Rules []Rule
	// Valid tells if the Route is valid.
	// If it is invalid, NGF should not generate any configuration for it.
	Valid bool
	// Attachable tells if the Route can be attached to any of the Gateways.
	// Route can be invalid but still attachable.
	Attachable bool
}

// buildGRPCRoutesForGateways builds routes from GRPCRoutes that reference any of the specified Gateways.
func buildGRPCRoutesForGateways(
	validator validation.HTTPFieldsValidator,
	grpcRoutes map[types.NamespacedName]*v1alpha2.GRPCRoute,
	gatewayNsNames []types.NamespacedName,
) map[types.NamespacedName]*GRPCRoute {
	if len(gatewayNsNames) == 0 {
		return nil
	}

	routes := make(map[types.NamespacedName]*GRPCRoute)

	for _, ghr := range grpcRoutes {
		r := buildGRPCRoute(validator, ghr, gatewayNsNames)
		if r != nil {
			routes[client.ObjectKeyFromObject(ghr)] = r
		}
	}

	return routes
}

func buildGRPCRoute(
	validator validation.HTTPFieldsValidator,
	ghr *v1alpha2.GRPCRoute,
	gatewayNsNames []types.NamespacedName,
) *GRPCRoute {
	r := &GRPCRoute{
		Source: ghr,
	}

	sectionNameRefs, err := buildSectionNameRefs(ghr.Spec.ParentRefs, ghr.Namespace, gatewayNsNames)
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
		ghr.Spec.Hostnames,
		field.NewPath("spec").Child("hostnames"),
	); err != nil {
		r.Valid = false
		r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedValue(err.Error()))

		return r
	}

	r.Valid = true
	r.Attachable = true
	var rules []Rule
	var atLeastOneValid bool
	var allRulesErrs field.ErrorList

	rules, atLeastOneValid, allRulesErrs = processGRPCRouteRules(ghr.Spec.Rules, validator)

	r.Rules = rules

	if len(allRulesErrs) > 0 {
		msg := allRulesErrs.ToAggregate().Error()

		if atLeastOneValid {
			r.Conditions = append(r.Conditions, staticConds.NewRoutePartiallyInvalid(msg))
		} else {
			msg = "All rules are invalid: " + msg
			r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedValue(msg))

			r.Valid = false
		}
	}

	return r
}

func processGRPCRouteRules(
	specRules []v1alpha2.GRPCRouteRule,
	validator validation.HTTPFieldsValidator,
) ([]Rule, bool, field.ErrorList) {
	rules := make([]Rule, len(specRules))
	var allRulesErrs field.ErrorList
	atLeastOneValid := false
	validFilters := true

	for i, rule := range specRules {
		rulePath := field.NewPath("spec").Child("rules").Index(i)

		var allErrs field.ErrorList

		var matchesErrs field.ErrorList
		for j, match := range rule.Matches {
			matchPath := rulePath.Child("matches").Index(j)
			matchesErrs = append(matchesErrs, validateGRPCMatch(validator, match, matchPath)...)
		}

		if len(rule.Filters) > 0 {
			filterPath := rulePath.Child("filters")
			allErrs = append(
				allErrs,
				field.NotSupported(filterPath, rule.Filters, []string{"gRPC filters are not yet supported"}),
			)
			validFilters = false
		}

		// rule.BackendRefs are validated separately because of their special requirements

		allErrs = append(allErrs, matchesErrs...)
		allRulesErrs = append(allRulesErrs, allErrs...)

		if len(allErrs) == 0 {
			atLeastOneValid = true
		}

		rules[i] = Rule{
			ValidMatches: len(matchesErrs) == 0,
			ValidFilters: validFilters,
		}
	}
	return rules, atLeastOneValid, allRulesErrs
}

func validateGRPCMatch(
	validator validation.HTTPFieldsValidator,
	match v1alpha2.GRPCRouteMatch,
	matchPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	methodPath := matchPath.Child("method")
	allErrs = append(allErrs, validateGRPCMethodMatch(match.Method, methodPath)...)

	for j, h := range match.Headers {
		headerPath := matchPath.Child("headers").Index(j)
		allErrs = append(allErrs, validateHeaderMatch(validator, h.Type, string(h.Name), h.Value, headerPath)...)
	}

	return allErrs
}

func validateGRPCMethodMatch(
	method *v1alpha2.GRPCMethodMatch,
	methodPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if method != nil {
		if method.Type == nil {
			allErrs = append(allErrs, field.Required(methodPath.Child("type"), "cannot be empty"))
		} else if *method.Type != v1alpha2.GRPCMethodMatchExact {
			allErrs = append(
				allErrs,
				field.NotSupported(methodPath.Child("type"), *method.Type, []string{string(v1alpha2.GRPCMethodMatchExact)}),
			)
		}
		if method.Service == nil || *method.Service == "" {
			methodServicePath := methodPath.Child("service")
			allErrs = append(
				allErrs,
				field.Required(methodServicePath, "service is required"),
			)
		}
		if method.Method == nil || *method.Method == "" {
			methodMethodPath := methodPath.Child("method")
			allErrs = append(allErrs, field.Required(methodMethodPath, "method is required"))
		}
	}
	return allErrs
}
