package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

func buildGRPCRoute(
	validator validation.HTTPFieldsValidator,
	ghr *v1.GRPCRoute,
	gatewayNsNames []types.NamespacedName,
	http2disabled bool,
) *L7Route {
	r := &L7Route{
		Source:    ghr,
		RouteType: RouteTypeGRPC,
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

	if http2disabled {
		r.Valid = false
		msg := "HTTP2 is disabled - cannot configure GRPCRoutes"
		r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedConfiguration(msg))

		return r
	}

	if err := validateHostnames(
		ghr.Spec.Hostnames,
		field.NewPath("spec").Child("hostnames"),
	); err != nil {
		r.Valid = false
		r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedValue(err.Error()))

		return r
	}

	r.Spec.Hostnames = ghr.Spec.Hostnames

	r.Valid = true
	r.Attachable = true

	rules, atLeastOneValid, allRulesErrs := processGRPCRouteRules(ghr.Spec.Rules, validator)

	r.Spec.Rules = rules

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
	specRules []v1.GRPCRouteRule,
	validator validation.HTTPFieldsValidator,
) (rules []RouteRule, atLeastOneValid bool, allRulesErrs field.ErrorList) {
	rules = make([]RouteRule, len(specRules))

	for i, rule := range specRules {
		rulePath := field.NewPath("spec").Child("rules").Index(i)

		var allErrs field.ErrorList
		var matchesErrs field.ErrorList
		var filtersErrs field.ErrorList

		for j, match := range rule.Matches {
			matchPath := rulePath.Child("matches").Index(j)
			matchesErrs = append(matchesErrs, validateGRPCMatch(validator, match, matchPath)...)
		}

		for j, filter := range rule.Filters {
			filterPath := rulePath.Child("filters").Index(j)
			filtersErrs = append(filtersErrs, validateGRPCFilter(validator, filter, filterPath)...)
		}

		backendRefs := make([]RouteBackendRef, 0, len(rule.BackendRefs))

		// rule.BackendRefs are validated separately because of their special requirements
		for _, b := range rule.BackendRefs {
			var interfaceFilters []interface{}
			if len(b.Filters) > 0 {
				interfaceFilters = make([]interface{}, 0, len(b.Filters))
				for i, v := range b.Filters {
					interfaceFilters[i] = v
				}
			}
			rbr := RouteBackendRef{
				BackendRef: b.BackendRef,
				Filters:    interfaceFilters,
			}
			backendRefs = append(backendRefs, rbr)
		}

		allErrs = append(allErrs, matchesErrs...)
		allErrs = append(allErrs, filtersErrs...)
		allRulesErrs = append(allRulesErrs, allErrs...)

		if len(allErrs) == 0 {
			atLeastOneValid = true
		}

		validFilters := len(filtersErrs) == 0

		var convertedFilters []v1.HTTPRouteFilter
		if validFilters {
			convertedFilters = convertGRPCFilters(rule.Filters)
		}

		rules[i] = RouteRule{
			ValidMatches:     len(matchesErrs) == 0,
			ValidFilters:     validFilters,
			Matches:          convertGRPCMatches(rule.Matches),
			Filters:          convertedFilters,
			RouteBackendRefs: backendRefs,
		}
	}
	return rules, atLeastOneValid, allRulesErrs
}

func convertGRPCMatches(grpcMatches []v1.GRPCRouteMatch) []v1.HTTPRouteMatch {
	pathValue := "/"
	pathType := v1.PathMatchType("PathPrefix")
	// If no matches are specified, the implementation MUST match every gRPC request.
	if len(grpcMatches) == 0 {
		return []v1.HTTPRouteMatch{
			{
				Path: &v1.HTTPPathMatch{
					Type:  &pathType,
					Value: helpers.GetPointer(pathValue),
				},
			},
		}
	}

	hms := make([]v1.HTTPRouteMatch, 0, len(grpcMatches))

	for _, gm := range grpcMatches {
		var hm v1.HTTPRouteMatch
		hmHeaders := make([]v1.HTTPHeaderMatch, 0, len(gm.Headers))
		for _, head := range gm.Headers {
			hmHeaders = append(hmHeaders, v1.HTTPHeaderMatch{
				Name:  v1.HTTPHeaderName(head.Name),
				Value: head.Value,
			})
		}
		hm.Headers = hmHeaders

		if gm.Method != nil && gm.Method.Service != nil && gm.Method.Method != nil {
			// if method match is provided, service and method are required
			// as the only method type supported is exact.
			// Validation has already been done at this point, and the condition will
			// have been added there if required.
			pathValue = "/" + *gm.Method.Service + "/" + *gm.Method.Method
			pathType = v1.PathMatchType("Exact")
		}
		hm.Path = &v1.HTTPPathMatch{
			Type:  &pathType,
			Value: helpers.GetPointer(pathValue),
		}

		hms = append(hms, hm)
	}
	return hms
}

func validateGRPCMatch(
	validator validation.HTTPFieldsValidator,
	match v1.GRPCRouteMatch,
	matchPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	methodPath := matchPath.Child("method")
	allErrs = append(allErrs, validateGRPCMethodMatch(validator, match.Method, methodPath)...)

	for j, h := range match.Headers {
		headerPath := matchPath.Child("headers").Index(j)
		allErrs = append(allErrs, validateHeaderMatch(validator, h.Type, string(h.Name), h.Value, headerPath)...)
	}

	return allErrs
}

func validateGRPCMethodMatch(
	validator validation.HTTPFieldsValidator,
	method *v1.GRPCMethodMatch,
	methodPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if method != nil {
		methodServicePath := methodPath.Child("service")
		methodMethodPath := methodPath.Child("method")
		if method.Type == nil {
			allErrs = append(allErrs, field.Required(methodPath.Child("type"), "cannot be empty"))
		} else if *method.Type != v1.GRPCMethodMatchExact {
			allErrs = append(
				allErrs,
				field.NotSupported(methodPath.Child("type"), *method.Type, []string{string(v1.GRPCMethodMatchExact)}),
			)
		}
		if method.Service == nil || *method.Service == "" {
			allErrs = append(allErrs, field.Required(methodServicePath, "service is required"))
		} else {
			pathValue := "/" + *method.Service
			if err := validator.ValidatePathInMatch(pathValue); err != nil {
				valErr := field.Invalid(methodServicePath, *method.Service, err.Error())
				allErrs = append(allErrs, valErr)
			}
		}
		if method.Method == nil || *method.Method == "" {
			allErrs = append(allErrs, field.Required(methodMethodPath, "method is required"))
		} else {
			pathValue := "/" + *method.Method
			if err := validator.ValidatePathInMatch(pathValue); err != nil {
				valErr := field.Invalid(methodMethodPath, *method.Method, err.Error())
				allErrs = append(allErrs, valErr)
			}
		}
	}
	return allErrs
}

func validateGRPCFilter(
	validator validation.HTTPFieldsValidator,
	filter v1.GRPCRouteFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	switch filter.Type {
	case v1.GRPCRouteFilterRequestHeaderModifier:
		return validateFilterHeaderModifier(validator, filter.RequestHeaderModifier, filterPath.Child(string(filter.Type)))
	case v1.GRPCRouteFilterResponseHeaderModifier:
		return validateFilterHeaderModifier(validator, filter.ResponseHeaderModifier, filterPath.Child(string(filter.Type)))
	default:
		valErr := field.NotSupported(
			filterPath.Child("type"),
			filter.Type,
			[]string{
				string(v1.GRPCRouteFilterRequestHeaderModifier),
				string(v1.GRPCRouteFilterResponseHeaderModifier),
			},
		)
		allErrs = append(allErrs, valErr)
		return allErrs
	}
}

// convertGRPCFilters converts GRPCRouteFilters (a subset of HTTPRouteFilter) to HTTPRouteFilters
// so we can reuse the logic from HTTPRoute filter validation and processing.
func convertGRPCFilters(filters []v1.GRPCRouteFilter) []v1.HTTPRouteFilter {
	if len(filters) == 0 {
		return nil
	}
	httpFilters := make([]v1.HTTPRouteFilter, 0, len(filters))
	for _, filter := range filters {
		switch filter.Type {
		case v1.GRPCRouteFilterRequestHeaderModifier:
			httpRequestHeaderFilter := v1.HTTPRouteFilter{
				Type:                  v1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: filter.RequestHeaderModifier,
			}
			httpFilters = append(httpFilters, httpRequestHeaderFilter)
		case v1.GRPCRouteFilterResponseHeaderModifier:
			httpResponseHeaderFilter := v1.HTTPRouteFilter{
				Type:                   v1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: filter.ResponseHeaderModifier,
			}
			httpFilters = append(httpFilters, httpResponseHeaderFilter)
		default:
			continue
		}
	}
	return httpFilters
}
