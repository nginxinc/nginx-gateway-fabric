package graph

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
)

// Rule represents a rule of an HTTPRoute.
type Rule struct {
	// BackendGroup is the BackendGroup of the rule.
	BackendGroup BackendGroup
	// ValidMatches indicates whether the matches of the rule are valid.
	// If the matches are invalid, NGK should not generate any configuration for the rule.
	ValidMatches bool
	// ValidFilters indicates whether the filters of the rule are valid.
	// If the filters are invalid, the data-plane should return 500 error provided that the matches are valid.
	ValidFilters bool
}

// ParentRef describes a reference to a parent in an HTTPRoute.
type ParentRef struct {
	// Gateway is the NamespacedName of the referenced Gateway
	Gateway types.NamespacedName
	// Idx is the index of the reference in the HTTPRoute.
	Idx int
}

// Route represents an HTTPRoute.
type Route struct {
	// Source is the source resource of the Route.
	// FIXME(pleshakov)
	// For now, we assume that the source is only HTTPRoute.
	// Later we can support more types - TLSRoute, TCPRoute and UDPRoute.
	Source *v1beta1.HTTPRoute
	// SectionNameRefs is a map of section names to the referenced NKG Gateways
	SectionNameRefs map[string]ParentRef
	// UnboundSectionNameRefs is a subset of SectionNameRefs that includes sections that could not be bound
	// to the referenced Gateway. For example, because section does not exist in the Gateway.
	UnboundSectionNameRefs map[string]conditions.Condition
	// Conditions include Conditions for the HTTPRoute.
	Conditions []conditions.Condition
	// Rules include Rules for the HTTPRoute. Each Rule[i] corresponds to the ith HTTPRouteRule.
	// If the Route is invalid, this field is nil
	Rules []Rule
	// Valid tells if the Route is valid.
	// If it is invalid, NGK should not generate any configuration for it.
	Valid bool
}

// GetAllBackendGroups returns BackendGroups for all rules with valid matches and filters in the Route.
func (r *Route) GetAllBackendGroups() []BackendGroup {
	count := 0

	for _, rule := range r.Rules {
		if rule.ValidMatches && rule.ValidFilters {
			count++
		}
	}

	if count == 0 {
		return nil
	}

	groups := make([]BackendGroup, 0, count)

	for _, rule := range r.Rules {
		if rule.ValidMatches && rule.ValidFilters {
			groups = append(groups, rule.BackendGroup)
		}
	}

	return groups
}

// GetAllConditionsForSectionName returns all Conditions for the referenced section name.
// It panics if the section name does not exist.
func (r *Route) GetAllConditionsForSectionName(name string) []conditions.Condition {
	if _, exist := r.SectionNameRefs[name]; !exist {
		panic(fmt.Errorf("section name %q does not exist", name))
	}

	count := len(r.Conditions)

	unboundCond, sectionIsUnbound := r.UnboundSectionNameRefs[name]
	if sectionIsUnbound {
		count++
	}

	if count == 0 {
		return nil
	}

	conds := make([]conditions.Condition, 0, count)

	if sectionIsUnbound {
		conds = append(conds, unboundCond)
	}

	conds = append(conds, r.Conditions...)

	return conds
}

// buildRoutes builds routes from HTTPRoutes excluding the ones that don't reference any of the NKG Gateways.
func buildRoutes(
	validator validation.HTTPFieldsValidator,
	httpRoutes map[types.NamespacedName]*v1beta1.HTTPRoute,
	gatewayNsNames []types.NamespacedName,
) map[types.NamespacedName]*Route {
	if len(gatewayNsNames) == 0 {
		return nil
	}

	routes := make(map[types.NamespacedName]*Route)

	for _, ghr := range httpRoutes {
		r := buildRoute(validator, ghr, gatewayNsNames)
		if r != nil {
			routes[client.ObjectKeyFromObject(ghr)] = r
		}
	}

	return routes
}

func buildSectionNameRefs(
	parentRefs []v1beta1.ParentReference,
	routeNamespace string,
	gatewayNsNames []types.NamespacedName,
) map[string]ParentRef {
	sectionNameRefs := make(map[string]ParentRef)

	for i, p := range parentRefs {
		gw, found := findGatewayForParentRef(p, routeNamespace, gatewayNsNames)
		if !found {
			continue
		}

		// The imported Webhook validation ensures unique section names.
		// Additionally, it ensures that if multiple refs reference the same Gateway, their section names
		// are not empty
		// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

		// FIXME(pleshakov): SectionNames across multiple Gateways might collide. Fix that.
		var sectionName string
		if p.SectionName != nil {
			sectionName = string(*p.SectionName)
		}

		sectionNameRefs[sectionName] = ParentRef{
			Idx:     i,
			Gateway: gw,
		}
	}

	return sectionNameRefs
}

func findGatewayForParentRef(
	ref v1beta1.ParentReference,
	routeNamespace string,
	gatewayNsNames []types.NamespacedName,
) (gwNsName types.NamespacedName, found bool) {
	if ref.Kind != nil && *ref.Kind != "Gateway" {
		return types.NamespacedName{}, false
	}
	if ref.Group != nil && *ref.Group != v1beta1.GroupName {
		return types.NamespacedName{}, false
	}

	// if the namespace is missing, assume the namespace of the HTTPRoute
	ns := routeNamespace
	if ref.Namespace != nil {
		ns = string(*ref.Namespace)
	}

	for _, gw := range gatewayNsNames {
		if gw.Namespace == ns && gw.Name == string(ref.Name) {
			return gw, true
		}
	}

	return types.NamespacedName{}, false
}

func buildRoute(
	validator validation.HTTPFieldsValidator,
	ghr *v1beta1.HTTPRoute,
	gatewayNsNames []types.NamespacedName,
) *Route {
	sectionNameRefs := buildSectionNameRefs(ghr.Spec.ParentRefs, ghr.Namespace, gatewayNsNames)
	// route doesn't belong to any of the Gateways
	if len(sectionNameRefs) == 0 {
		return nil
	}

	r := &Route{
		Source:                 ghr,
		SectionNameRefs:        sectionNameRefs,
		UnboundSectionNameRefs: map[string]conditions.Condition{},
	}

	err := validateHostnames(validator, ghr.Spec.Hostnames, field.NewPath("spec").Child("hostnames"))
	if err != nil {
		r.Valid = false
		r.Conditions = append(r.Conditions, conditions.NewRouteUnsupportedValue(err.Error()))

		return r
	}

	r.Valid = true

	r.Rules = make([]Rule, len(ghr.Spec.Rules))

	atLeastOneValid := false
	var allRulesErrs field.ErrorList

	for i, rule := range ghr.Spec.Rules {
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

		// rule.BackendRefs are validated separately because of their special requirements

		var allErrs field.ErrorList
		allErrs = append(allErrs, matchesErrs...)
		allErrs = append(allErrs, filtersErrs...)
		allRulesErrs = append(allRulesErrs, allErrs...)

		if len(allErrs) == 0 {
			atLeastOneValid = true
		}

		r.Rules[i] = Rule{
			ValidMatches: len(matchesErrs) == 0,
			ValidFilters: len(filtersErrs) == 0,
			BackendGroup: BackendGroup{
				Source:  client.ObjectKeyFromObject(r.Source),
				RuleIdx: i,
			},
		}
	}

	if len(allRulesErrs) > 0 {
		msg := allRulesErrs.ToAggregate().Error()

		if atLeastOneValid {
			// FIXME(pleshakov): Partial validity for HTTPRoute rules is not defined in the Gateway API spec yet.
			// See https://github.com/kubernetes-sigs/gateway-api/issues/1696
			msg = "Some rules are invalid: " + msg
			r.Conditions = append(r.Conditions, conditions.NewTODO(msg))
		} else {
			msg = "All rules are invalid: " + msg
			r.Conditions = append(r.Conditions, conditions.NewRouteUnsupportedValue(msg))
		}
	}

	return r
}

func bindRoutesToListeners(
	routes map[types.NamespacedName]*Route,
	gw *Gateway,
) {
	if gw == nil {
		return
	}

	for _, r := range routes {
		bindRouteToListeners(r, gw)
	}
}

func bindRouteToListeners(
	r *Route,
	gw *Gateway,
) {
	if !r.Valid {
		return
	}

	for name, ref := range r.SectionNameRefs {
		routeRef := r.Source.Spec.ParentRefs[ref.Idx]

		path := field.NewPath("spec").Child("parentRefs").Index(ref.Idx)

		if routeRef.SectionName == nil || *routeRef.SectionName == "" {
			valErr := field.Required(path.Child("sectionName"), "cannot be empty")
			r.UnboundSectionNameRefs[name] = conditions.NewRouteUnsupportedValue(valErr.Error())
			continue
		}

		if routeRef.Port != nil {
			valErr := field.Forbidden(path.Child("port"), "cannot be set")
			r.UnboundSectionNameRefs[name] = conditions.NewRouteUnsupportedValue(valErr.Error())
			continue
		}

		// Case 1 - winning Gateway
		if ref.Gateway.Namespace == gw.Source.Namespace && ref.Gateway.Name == gw.Source.Name {
			// Find a listener

			// FIXME(pleshakov)
			// For now, let's do simple matching.
			// However, we need to also support wildcard matching.
			// More over, we need to handle cases when a Route host matches multiple HTTP listeners on the same port
			// when sectionName is empty and only choose one listener.
			// For example:
			// - Route with host foo.example.com;
			// - listener 1 for port 80 with hostname foo.example.com
			// - listener 2 for port 80 with hostname *.example.com;
			// In this case, the Route host foo.example.com should choose listener 1, as it is a more specific match.

			l, exists := gw.Listeners[name]
			if !exists {
				// FIXME(pleshakov): Add a proper condition once it is available in the Gateway API.
				// https://github.com/nginxinc/nginx-kubernetes-gateway/issues/306
				r.UnboundSectionNameRefs[name] = conditions.NewTODO("listener is not found")
				continue
			}

			if !l.Valid {
				r.UnboundSectionNameRefs[name] = conditions.NewRouteInvalidListener()
				continue
			}

			accepted := findAcceptedHostnames(l.Source.Hostname, r.Source.Spec.Hostnames)

			if len(accepted) > 0 {
				for _, h := range accepted {
					l.AcceptedHostnames[h] = struct{}{}
				}
				l.Routes[client.ObjectKeyFromObject(r.Source)] = r
			} else {
				r.UnboundSectionNameRefs[name] = conditions.NewRouteNoMatchingListenerHostname()
			}

			continue
		}

		// Case 2: the parentRef references an ignored Gateway resource.
		r.UnboundSectionNameRefs[name] = conditions.NewTODO("Gateway is ignored")
	}
}

func findAcceptedHostnames(listenerHostname *v1beta1.Hostname, routeHostnames []v1beta1.Hostname) []string {
	hostname := getHostname(listenerHostname)

	match := func(h v1beta1.Hostname) bool {
		if hostname == "" {
			return true
		}
		return string(h) == hostname
	}

	var result []string

	for _, h := range routeHostnames {
		if match(h) {
			result = append(result, string(h))
		}
	}

	return result
}

func getHostname(h *v1beta1.Hostname) string {
	if h == nil {
		return ""
	}
	return string(*h)
}

func validateHostnames(validator validation.HTTPFieldsValidator, hostnames []v1beta1.Hostname, path *field.Path) error {
	var allErrs field.ErrorList

	for i := range hostnames {
		if err := validateHostname(string(hostnames[i])); err != nil {
			allErrs = append(allErrs, field.Invalid(path.Index(i), hostnames[i], err.Error()))
			continue
		}
		if err := validator.ValidateHostnameInServer(string(hostnames[i])); err != nil {
			allErrs = append(allErrs, field.Invalid(path.Index(i), hostnames[i], err.Error()))
		}
	}

	return allErrs.ToAggregate()
}

func validateMatch(
	validator validation.HTTPFieldsValidator,
	match v1beta1.HTTPRouteMatch,
	matchPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	pathPath := matchPath.Child("path")
	allErrs = append(allErrs, validatePathMatch(validator, match.Path, pathPath)...)

	for j, h := range match.Headers {
		headerPath := matchPath.Child("headers").Index(j)
		allErrs = append(allErrs, validateHeaderMatch(validator, h, headerPath)...)
	}

	for j, q := range match.QueryParams {
		queryParamPath := matchPath.Child("queryParams").Index(j)
		allErrs = append(allErrs, validateQueryParamMatch(validator, q, queryParamPath)...)
	}

	allErrs = append(allErrs, validateMethodMatch(validator, match.Method, matchPath.Child("method"))...)

	return allErrs
}

func validateMethodMatch(
	validator validation.HTTPFieldsValidator,
	method *v1beta1.HTTPMethod,
	methodPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if method == nil {
		return allErrs
	}

	if valid, supportedValues := validator.ValidateMethodInMatch(string(*method)); !valid {
		valErr := field.NotSupported(methodPath, *method, supportedValues)
		allErrs = append(allErrs, valErr)
	}

	return allErrs
}

func validateQueryParamMatch(
	validator validation.HTTPFieldsValidator,
	q v1beta1.HTTPQueryParamMatch,
	queryParamPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if q.Type == nil {
		allErrs = append(allErrs, field.Required(queryParamPath.Child("type"), "cannot be empty"))
	} else if *q.Type != v1beta1.QueryParamMatchExact {
		valErr := field.NotSupported(queryParamPath.Child("type"), *q.Type, []string{string(v1beta1.QueryParamMatchExact)})
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidateQueryParamNameInMatch(q.Name); err != nil {
		valErr := field.Invalid(queryParamPath.Child("name"), q.Name, err.Error())
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidateQueryParamValueInMatch(q.Value); err != nil {
		valErr := field.Invalid(queryParamPath.Child("value"), q.Value, err.Error())
		allErrs = append(allErrs, valErr)
	}

	return allErrs
}

func validateHeaderMatch(
	validator validation.HTTPFieldsValidator,
	header v1beta1.HTTPHeaderMatch,
	headerPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if header.Type == nil {
		allErrs = append(allErrs, field.Required(headerPath.Child("type"), "cannot be empty"))
	} else if *header.Type != v1beta1.HeaderMatchExact {
		valErr := field.NotSupported(
			headerPath.Child("type"),
			*header.Type,
			[]string{string(v1beta1.HeaderMatchExact)},
		)
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidateHeaderNameInMatch(string(header.Name)); err != nil {
		valErr := field.Invalid(headerPath.Child("name"), header.Name, err.Error())
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidateHeaderValueInMatch(header.Value); err != nil {
		valErr := field.Invalid(headerPath.Child("value"), header.Value, err.Error())
		allErrs = append(allErrs, valErr)
	}

	return allErrs
}

func validatePathMatch(
	validator validation.HTTPFieldsValidator,
	path *v1beta1.HTTPPathMatch,
	fieldPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if path == nil {
		return allErrs
	}

	// The imported Webhook validation ensures the path type and value are not nil
	// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

	if *path.Type != v1beta1.PathMatchPathPrefix {
		valErr := field.NotSupported(fieldPath, *path.Type, []string{string(v1beta1.PathMatchPathPrefix)})
		allErrs = append(allErrs, valErr)
	}

	if err := validator.ValidatePathInPrefixMatch(*path.Value); err != nil {
		valErr := field.Invalid(fieldPath, *path.Value, err.Error())
		allErrs = append(allErrs, valErr)
	}

	return allErrs
}

func validateFilter(
	validator validation.HTTPFieldsValidator,
	filter v1beta1.HTTPRouteFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if filter.Type != v1beta1.HTTPRouteFilterRequestRedirect {
		valErr := field.NotSupported(
			filterPath.Child("type"),
			filter.Type,
			[]string{string(v1beta1.HTTPRouteFilterRequestRedirect)},
		)
		allErrs = append(allErrs, valErr)
		return allErrs
	}

	// The imported Webhook validation ensures  filter.RequestRedirect is not nil
	// FIXME(pleshakov): Add a unit test for the imported Webhook validation code for this case.

	redirect := filter.RequestRedirect

	redirectPath := filterPath.Child("requestRedirect")

	if redirect.Scheme != nil {
		if valid, supportedValues := validator.ValidateRedirectScheme(*redirect.Scheme); !valid {
			valErr := field.NotSupported(redirectPath.Child("scheme"), *redirect.Scheme, supportedValues)
			allErrs = append(allErrs, valErr)
		}
	}

	if redirect.Hostname != nil {
		if err := validator.ValidateRedirectHostname(string(*redirect.Hostname)); err != nil {
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
