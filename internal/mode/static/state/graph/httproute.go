package graph

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/http"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

var (
	add    = "add"
	set    = "set"
	remove = "remove"
)

func buildHTTPRoute(
	validator validation.HTTPFieldsValidator,
	ghr *v1.HTTPRoute,
	gatewayNsNames []types.NamespacedName,
) *L7Route {
	r := &L7Route{
		Source:    ghr,
		RouteType: RouteTypeHTTP,
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

	r.Spec.Hostnames = ghr.Spec.Hostnames

	r.Valid = true
	r.Attachable = true

	rules, atLeastOneValid, allRulesErrs := processHTTPRouteRules(ghr.Spec.Rules, validator)

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

func processHTTPRouteRules(
	specRules []v1.HTTPRouteRule,
	validator validation.HTTPFieldsValidator,
) (rules []RouteRule, atLeastOneValid bool, allRulesErrs field.ErrorList) {
	rules = make([]RouteRule, len(specRules))

	for i, rule := range specRules {
		rulePath := field.NewPath("spec").Child("rules").Index(i)

		var matchesErrs field.ErrorList
		for j, match := range rule.Matches {
			matchPath := rulePath.Child("matches").Index(j)
			matchesErrs = append(matchesErrs, validateMatch(validator, match, matchPath)...)
		}

		var filtersErrs field.ErrorList
		for j, filter := range rule.Filters {
			filterPath := rulePath.Child("filters").Index(j)
			filtersErrs = append(filtersErrs, validateFilter(validator, filter, filterPath)...)
		}

		var allErrs field.ErrorList
		allErrs = append(allErrs, matchesErrs...)
		allErrs = append(allErrs, filtersErrs...)
		allRulesErrs = append(allRulesErrs, allErrs...)

		if len(allErrs) == 0 {
			atLeastOneValid = true
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

		rules[i] = RouteRule{
			ValidMatches:     len(matchesErrs) == 0,
			ValidFilters:     len(filtersErrs) == 0,
			Matches:          rule.Matches,
			Filters:          rule.Filters,
			RouteBackendRefs: backendRefs,
		}
	}
	return rules, atLeastOneValid, allRulesErrs
}

func validateMatch(
	validator validation.HTTPFieldsValidator,
	match v1.HTTPRouteMatch,
	matchPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	pathPath := matchPath.Child("path")
	allErrs = append(allErrs, validatePathMatch(validator, match.Path, pathPath)...)

	for j, h := range match.Headers {
		headerPath := matchPath.Child("headers").Index(j)
		allErrs = append(allErrs, validateHeaderMatch(validator, h.Type, string(h.Name), h.Value, headerPath)...)
	}

	for j, q := range match.QueryParams {
		queryParamPath := matchPath.Child("queryParams").Index(j)
		allErrs = append(allErrs, validateQueryParamMatch(validator, q, queryParamPath)...)
	}

	if err := validateMethodMatch(
		validator,
		match.Method,
		matchPath.Child("method"),
	); err != nil {
		allErrs = append(allErrs, err)
	}

	return allErrs
}

func validateMethodMatch(
	validator validation.HTTPFieldsValidator,
	method *v1.HTTPMethod,
	methodPath *field.Path,
) *field.Error {
	if method == nil {
		return nil
	}

	if valid, supportedValues := validator.ValidateMethodInMatch(string(*method)); !valid {
		return field.NotSupported(methodPath, *method, supportedValues)
	}

	return nil
}

func validateQueryParamMatch(
	validator validation.HTTPFieldsValidator,
	q v1.HTTPQueryParamMatch,
	queryParamPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if q.Type == nil {
		allErrs = append(allErrs, field.Required(queryParamPath.Child("type"), "cannot be empty"))
	} else if *q.Type != v1.QueryParamMatchExact {
		valErr := field.NotSupported(queryParamPath.Child("type"), *q.Type, []string{string(v1.QueryParamMatchExact)})
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidateQueryParamNameInMatch(string(q.Name)); err != nil {
		valErr := field.Invalid(queryParamPath.Child("name"), q.Name, err.Error())
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidateQueryParamValueInMatch(q.Value); err != nil {
		valErr := field.Invalid(queryParamPath.Child("value"), q.Value, err.Error())
		allErrs = append(allErrs, valErr)
	}

	return allErrs
}

func validatePathMatch(
	validator validation.HTTPFieldsValidator,
	path *v1.HTTPPathMatch,
	fieldPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if path == nil {
		return allErrs
	}

	if path.Type == nil {
		return field.ErrorList{field.Required(fieldPath.Child("type"), "path type cannot be nil")}
	}
	if path.Value == nil {
		return field.ErrorList{field.Required(fieldPath.Child("value"), "path value cannot be nil")}
	}

	if strings.HasPrefix(*path.Value, http.InternalRoutePathPrefix) {
		msg := fmt.Sprintf("path cannot start with %s. This prefix is reserved for internal use",
			http.InternalRoutePathPrefix)
		return field.ErrorList{field.Invalid(fieldPath.Child("value"), *path.Value, msg)}
	}

	if *path.Type != v1.PathMatchPathPrefix && *path.Type != v1.PathMatchExact {
		valErr := field.NotSupported(fieldPath.Child("type"), *path.Type,
			[]string{string(v1.PathMatchExact), string(v1.PathMatchPathPrefix)})
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidatePathInMatch(*path.Value); err != nil {
		valErr := field.Invalid(fieldPath.Child("value"), *path.Value, err.Error())
		allErrs = append(allErrs, valErr)
	}

	return allErrs
}

func validateFilter(
	validator validation.HTTPFieldsValidator,
	filter v1.HTTPRouteFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	switch filter.Type {
	case v1.HTTPRouteFilterRequestRedirect:
		return validateFilterRedirect(validator, filter.RequestRedirect, filterPath)
	case v1.HTTPRouteFilterURLRewrite:
		return validateFilterRewrite(validator, filter.URLRewrite, filterPath)
	case v1.HTTPRouteFilterRequestHeaderModifier:
		return validateFilterHeaderModifier(validator, filter.RequestHeaderModifier, filterPath.Child(string(filter.Type)))
	case v1.HTTPRouteFilterResponseHeaderModifier:
		return validateFilterResponseHeaderModifier(
			validator, filter.ResponseHeaderModifier, filterPath.Child(string(filter.Type)),
		)
	default:
		valErr := field.NotSupported(
			filterPath.Child("type"),
			filter.Type,
			[]string{
				string(v1.HTTPRouteFilterRequestRedirect),
				string(v1.HTTPRouteFilterURLRewrite),
				string(v1.HTTPRouteFilterRequestHeaderModifier),
				string(v1.HTTPRouteFilterResponseHeaderModifier),
			},
		)
		allErrs = append(allErrs, valErr)
		return allErrs
	}
}

func validateFilterRedirect(
	validator validation.HTTPFieldsValidator,
	redirect *v1.HTTPRequestRedirectFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	redirectPath := filterPath.Child("requestRedirect")

	if redirect == nil {
		return field.ErrorList{field.Required(redirectPath, "requestRedirect cannot be nil")}
	}

	if redirect.Scheme != nil {
		if valid, supportedValues := validator.ValidateRedirectScheme(*redirect.Scheme); !valid {
			valErr := field.NotSupported(redirectPath.Child("scheme"), *redirect.Scheme, supportedValues)
			allErrs = append(allErrs, valErr)
		}
	}

	if redirect.Hostname != nil {
		if err := validator.ValidateHostname(string(*redirect.Hostname)); err != nil {
			valErr := field.Invalid(redirectPath.Child("hostname"), *redirect.Hostname, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}

	if redirect.Port != nil {
		if err := validator.ValidateRedirectPort(int32(*redirect.Port)); err != nil {
			valErr := field.Invalid(redirectPath.Child("port"), *redirect.Port, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}

	if redirect.Path != nil {
		valErr := field.Forbidden(redirectPath.Child("path"), "path is not supported")
		allErrs = append(allErrs, valErr)
	}

	if redirect.StatusCode != nil {
		if valid, supportedValues := validator.ValidateRedirectStatusCode(*redirect.StatusCode); !valid {
			valErr := field.NotSupported(redirectPath.Child("statusCode"), *redirect.StatusCode, supportedValues)
			allErrs = append(allErrs, valErr)
		}
	}

	return allErrs
}

func validateFilterRewrite(
	validator validation.HTTPFieldsValidator,
	rewrite *v1.HTTPURLRewriteFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	rewritePath := filterPath.Child("urlRewrite")

	if rewrite == nil {
		return field.ErrorList{field.Required(rewritePath, "urlRewrite cannot be nil")}
	}

	if rewrite.Hostname != nil {
		if err := validator.ValidateHostname(string(*rewrite.Hostname)); err != nil {
			valErr := field.Invalid(rewritePath.Child("hostname"), *rewrite.Hostname, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}

	if rewrite.Path != nil {
		var path string
		switch rewrite.Path.Type {
		case v1.FullPathHTTPPathModifier:
			path = *rewrite.Path.ReplaceFullPath
		case v1.PrefixMatchHTTPPathModifier:
			path = *rewrite.Path.ReplacePrefixMatch
		default:
			msg := fmt.Sprintf("urlRewrite path type %s not supported", rewrite.Path.Type)
			valErr := field.Invalid(rewritePath.Child("path"), *rewrite.Path, msg)
			return append(allErrs, valErr)
		}

		if err := validator.ValidateRewritePath(path); err != nil {
			valErr := field.Invalid(rewritePath.Child("path"), *rewrite.Path, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}

	return allErrs
}
