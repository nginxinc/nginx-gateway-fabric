package graph

import (
	"errors"
	"fmt"
	"strings"

	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

const wildcardHostname = "~^"

// Rule represents a rule of an HTTPRoute.
type Rule struct {
	// BackendRefs is a list of BackendRefs for the rule.
	BackendRefs []BackendRef
	// ValidMatches indicates whether the matches of the rule are valid.
	// If the matches are invalid, NGF should not generate any configuration for the rule.
	ValidMatches bool
	// ValidFilters indicates whether the filters of the rule are valid.
	// If the filters are invalid, the data-plane should return 500 error provided that the matches are valid.
	ValidFilters bool
}

// ParentRef describes a reference to a parent in an HTTPRoute.
type ParentRef struct {
	// Attachment is the attachment status of the ParentRef. It could be nil. In that case, NGF didn't attempt to
	// attach because of problems with the Route.
	Attachment *ParentRefAttachmentStatus
	// Gateway is the NamespacedName of the referenced Gateway
	Gateway types.NamespacedName
	// Idx is the index of the corresponding ParentReference in the HTTPRoute.
	Idx int
}

// ParentRefAttachmentStatus describes the attachment status of a ParentRef.
type ParentRefAttachmentStatus struct {
	// AcceptedHostnames is an intersection between the hostnames supported by an attached Listener
	// and the hostnames from this Route. Key is listener name, value is list of hostnames.
	AcceptedHostnames map[string][]string
	// FailedCondition is the condition that describes why the ParentRef is not attached to the Gateway. It is set
	// when Attached is false.
	FailedCondition conditions.Condition
	// Attached indicates if the ParentRef is attached to the Gateway.
	Attached bool
}

// Route represents an HTTPRoute.
type Route struct {
	// Source is the source resource of the Route.
	Source *v1beta1.HTTPRoute
	// ParentRefs includes ParentRefs with NGF Gateways only.
	ParentRefs []ParentRef
	// Conditions include Conditions for the HTTPRoute.
	Conditions []conditions.Condition
	// Rules include Rules for the HTTPRoute. Each Rule[i] corresponds to the ith HTTPRouteRule.
	// If the Route is invalid, this field is nil
	Rules []Rule
	// Valid tells if the Route is valid.
	// If it is invalid, NGF should not generate any configuration for it.
	Valid bool
}

// buildRoutesForGateways builds routes from HTTPRoutes that reference any of the specified Gateways.
func buildRoutesForGateways(
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
) []ParentRef {
	sectionNameRefs := make([]ParentRef, 0, len(parentRefs))

	type key struct {
		gwNsName    types.NamespacedName
		sectionName string
	}
	uniqueSectionsPerGateway := make(map[key]struct{})

	for i, p := range parentRefs {
		gw, found := findGatewayForParentRef(p, routeNamespace, gatewayNsNames)
		if !found {
			continue
		}

		var sectionName string
		if p.SectionName != nil {
			sectionName = string(*p.SectionName)
		}

		k := key{
			gwNsName:    gw,
			sectionName: sectionName,
		}

		if _, exist := uniqueSectionsPerGateway[k]; exist {
			panicForBrokenWebhookAssumption(
				fmt.Errorf("duplicate section name %q for Gateway %s", sectionName, gw.String()),
			)
		}
		uniqueSectionsPerGateway[k] = struct{}{}

		sectionNameRefs = append(sectionNameRefs, ParentRef{
			Idx:     i,
			Gateway: gw,
		})
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
		Source:     ghr,
		ParentRefs: sectionNameRefs,
	}

	if err := validateHostnames(
		ghr.Spec.Hostnames,
		field.NewPath("spec").Child("hostnames"),
	); err != nil {
		r.Valid = false
		r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedValue(err.Error()))

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
		}
	}

	if len(allRulesErrs) > 0 {
		msg := allRulesErrs.ToAggregate().Error()

		if atLeastOneValid {
			msg = "Dropped Rule(s): " + msg
			r.Conditions = append(r.Conditions, staticConds.NewRoutePartiallyInvalid(msg))
		} else {
			msg = "All rules are invalid: " + msg
			r.Conditions = append(r.Conditions, staticConds.NewRouteUnsupportedValue(msg))

			r.Valid = false
		}
	}

	return r
}

func bindRoutesToListeners(
	routes map[types.NamespacedName]*Route,
	gw *Gateway,
	namespaces map[types.NamespacedName]*apiv1.Namespace,
) {
	if gw == nil {
		return
	}

	for _, r := range routes {
		bindRouteToListeners(r, gw, namespaces)
	}
}

func bindRouteToListeners(r *Route, gw *Gateway, namespaces map[types.NamespacedName]*apiv1.Namespace) {
	if !r.Valid {
		return
	}

	for i := 0; i < len(r.ParentRefs); i++ {
		attachment := &ParentRefAttachmentStatus{
			AcceptedHostnames: make(map[string][]string),
		}
		ref := &r.ParentRefs[i]
		ref.Attachment = attachment

		routeRef := r.Source.Spec.ParentRefs[ref.Idx]

		path := field.NewPath("spec").Child("parentRefs").Index(ref.Idx)

		// Case 1: Attachment is not possible due to unsupported configuration

		if routeRef.Port != nil {
			valErr := field.Forbidden(path.Child("port"), "cannot be set")
			attachment.FailedCondition = staticConds.NewRouteUnsupportedValue(valErr.Error())
			continue
		}

		// Case 2: the parentRef references an ignored Gateway resource.

		referencesWinningGw := ref.Gateway.Namespace == gw.Source.Namespace && ref.Gateway.Name == gw.Source.Name

		if !referencesWinningGw {
			attachment.FailedCondition = staticConds.NewTODO("Gateway is ignored")
			continue
		}

		// Case 3: Attachment is not possible because Gateway is invalid

		if !gw.Valid {
			attachment.FailedCondition = staticConds.NewRouteInvalidGateway()
			continue
		}

		// Case 4 - winning Gateway

		// Try to attach Route to all matching listeners
		cond, attached := tryToAttachRouteToListeners(ref.Attachment, routeRef.SectionName, r, gw, namespaces)
		if !attached {
			attachment.FailedCondition = cond
			continue
		}

		attachment.Attached = true
	}
}

// tryToAttachRouteToListeners tries to attach the route to the listeners that match the parentRef and the hostnames.
// If it succeeds in attaching at least one listener it will return true and the condition will be empty.
// If it fails to attach the route, it will return false and the failure condition.
func tryToAttachRouteToListeners(
	refStatus *ParentRefAttachmentStatus,
	sectionName *v1beta1.SectionName,
	route *Route,
	gw *Gateway,
	namespaces map[types.NamespacedName]*apiv1.Namespace,
) (conditions.Condition, bool) {
	validListeners, listenerExists := findValidListeners(getSectionName(sectionName), gw.Listeners)

	if !listenerExists {
		return staticConds.NewRouteNoMatchingParent(), false
	}

	if len(validListeners) == 0 {
		return staticConds.NewRouteInvalidListener(), false
	}

	bind := func(l *Listener) (allowed, attached bool) {
		if !routeAllowedByListener(l, route.Source.Namespace, gw.Source.Namespace, namespaces) {
			return false, false
		}

		hostnames := findAcceptedHostnames(l.Source.Hostname, route.Source.Spec.Hostnames)
		if len(hostnames) == 0 {
			return true, false
		}

		refStatus.AcceptedHostnames[string(l.Source.Name)] = hostnames
		l.Routes[client.ObjectKeyFromObject(route.Source)] = route

		return true, true
	}

	var allowed, attached bool
	for _, l := range validListeners {
		routeAllowed, routeAttached := bind(l)
		allowed = allowed || routeAllowed
		attached = attached || routeAttached
	}

	if !attached {
		if !allowed {
			return staticConds.NewRouteNotAllowedByListeners(), false
		}
		return staticConds.NewRouteNoMatchingListenerHostname(), false
	}

	return conditions.Condition{}, true
}

// findValidListeners returns a list of valid listeners and whether the listener exists for a non-empty sectionName.
func findValidListeners(sectionName string, listeners map[string]*Listener) ([]*Listener, bool) {
	if sectionName != "" {
		l, exists := listeners[sectionName]
		if !exists {
			return nil, false
		}

		if l.Valid {
			return []*Listener{l}, true
		}

		return nil, true
	}

	validListeners := make([]*Listener, 0, len(listeners))
	for _, l := range listeners {
		if !l.Valid {
			continue
		}

		validListeners = append(validListeners, l)
	}

	return validListeners, true
}

func findAcceptedHostnames(listenerHostname *v1beta1.Hostname, routeHostnames []v1beta1.Hostname) []string {
	hostname := getHostname(listenerHostname)

	if len(routeHostnames) == 0 {
		if hostname == "" {
			return []string{wildcardHostname}
		}
		return []string{hostname}
	}

	var result []string

	for _, h := range routeHostnames {
		routeHost := string(h)
		if match(hostname, routeHost) {
			result = append(result, GetMoreSpecificHostname(hostname, routeHost))
		}
	}

	return result
}

func match(listenerHost, routeHost string) bool {
	if listenerHost == "" {
		return true
	}

	if routeHost == listenerHost {
		return true
	}

	wildcardMatch := func(host1, host2 string) bool {
		return strings.HasPrefix(host1, "*.") && strings.HasSuffix(host2, strings.TrimPrefix(host1, "*"))
	}

	// check if listenerHost is a wildcard and routeHost matches
	if wildcardMatch(listenerHost, routeHost) {
		return true
	}

	// check if routeHost is a wildcard and listener matchess
	return wildcardMatch(routeHost, listenerHost)
}

// GetMoreSpecificHostname returns the more specific hostname between the two inputs.
//
// This function assumes that the two hostnames match each other, either:
// - Exactly
// - One as a substring of the other
func GetMoreSpecificHostname(hostname1, hostname2 string) string {
	if hostname1 == hostname2 {
		return hostname1
	}
	if hostname1 == "" {
		return hostname2
	}
	if hostname2 == "" {
		return hostname1
	}

	// Compare if wildcards are present
	if strings.HasPrefix(hostname1, "*.") {
		if strings.HasPrefix(hostname2, "*.") {
			subdomains1 := strings.Split(hostname1, ".")
			subdomains2 := strings.Split(hostname2, ".")

			// Compare number of subdomains
			if len(subdomains1) > len(subdomains2) {
				return hostname1
			}

			return hostname2
		}

		return hostname2
	}
	if strings.HasPrefix(hostname2, "*.") {
		return hostname1
	}

	return ""
}

func routeAllowedByListener(
	listener *Listener,
	routeNS,
	gwNS string,
	namespaces map[types.NamespacedName]*apiv1.Namespace,
) bool {
	if listener.Source.AllowedRoutes != nil {
		switch *listener.Source.AllowedRoutes.Namespaces.From {
		case v1beta1.NamespacesFromAll:
			return true
		case v1beta1.NamespacesFromSame:
			return routeNS == gwNS
		case v1beta1.NamespacesFromSelector:
			if listener.AllowedRouteLabelSelector == nil {
				return false
			}

			ns, exists := namespaces[types.NamespacedName{Name: routeNS}]
			if !exists {
				panic(fmt.Errorf("route namespace %q not found in map", routeNS))
			}
			return listener.AllowedRouteLabelSelector.Matches(labels.Set(ns.Labels))
		}
	}
	return true
}

func getHostname(h *v1beta1.Hostname) string {
	if h == nil {
		return ""
	}
	return string(*h)
}

func getSectionName(s *v1beta1.SectionName) string {
	if s == nil {
		return ""
	}
	return string(*s)
}

func validateHostnames(hostnames []v1beta1.Hostname, path *field.Path) error {
	var allErrs field.ErrorList

	for i := range hostnames {
		if err := validateHostname(string(hostnames[i])); err != nil {
			allErrs = append(allErrs, field.Invalid(path.Index(i), hostnames[i], err.Error()))
			continue
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
	method *v1beta1.HTTPMethod,
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

	if path.Type == nil {
		panicForBrokenWebhookAssumption(errors.New("path type cannot be nil"))
	}
	if path.Value == nil {
		panicForBrokenWebhookAssumption(errors.New("path value cannot be nil"))
	}

	if *path.Type != v1beta1.PathMatchPathPrefix && *path.Type != v1beta1.PathMatchExact {
		valErr := field.NotSupported(fieldPath.Child("type"), *path.Type,
			[]string{string(v1beta1.PathMatchExact), string(v1beta1.PathMatchPathPrefix)})
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
	filter v1beta1.HTTPRouteFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	switch filter.Type {
	case v1beta1.HTTPRouteFilterRequestRedirect:
		return validateFilterRedirect(validator, filter, filterPath)
	case v1beta1.HTTPRouteFilterRequestHeaderModifier:
		return validateFilterHeaderModifier(validator, filter, filterPath)
	default:
		valErr := field.NotSupported(
			filterPath.Child("type"),
			filter.Type,
			[]string{
				string(v1beta1.HTTPRouteFilterRequestRedirect),
				string(v1beta1.HTTPRouteFilterRequestHeaderModifier),
			},
		)
		allErrs = append(allErrs, valErr)
		return allErrs
	}
}

func validateFilterRedirect(
	validator validation.HTTPFieldsValidator,
	filter v1beta1.HTTPRouteFilter,
	filterPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	if filter.RequestRedirect == nil {
		panicForBrokenWebhookAssumption(errors.New("requestRedirect cannot be nil"))
	}

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

func validateFilterHeaderModifier(
	validator validation.HTTPFieldsValidator,
	filter v1beta1.HTTPRouteFilter,
	filterPath *field.Path,
) field.ErrorList {
	headerModifier := filter.RequestHeaderModifier

	headerModifierPath := filterPath.Child("requestHeaderModifier")

	if headerModifier == nil {
		panicForBrokenWebhookAssumption(errors.New("requestHeaderModifier cannot be nil"))
	}

	return validateFilterHeaderModifierFields(validator, headerModifier, headerModifierPath)
}

func validateFilterHeaderModifierFields(
	validator validation.HTTPFieldsValidator,
	headerModifier *v1beta1.HTTPHeaderFilter,
	headerModifierPath *field.Path,
) field.ErrorList {
	var allErrs field.ErrorList

	for _, h := range headerModifier.Add {
		if err := validator.ValidateRequestHeaderName(string(h.Name)); err != nil {
			valErr := field.Invalid(headerModifierPath.Child("add"), h, err.Error())
			allErrs = append(allErrs, valErr)
		}
		if err := validator.ValidateRequestHeaderValue(h.Value); err != nil {
			valErr := field.Invalid(headerModifierPath.Child("add"), h, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}
	for _, h := range headerModifier.Set {
		if err := validator.ValidateRequestHeaderName(string(h.Name)); err != nil {
			valErr := field.Invalid(headerModifierPath.Child("set"), h, err.Error())
			allErrs = append(allErrs, valErr)
		}
		if err := validator.ValidateRequestHeaderValue(h.Value); err != nil {
			valErr := field.Invalid(headerModifierPath.Child("set"), h, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}
	for _, h := range headerModifier.Remove {
		if err := validator.ValidateRequestHeaderName(h); err != nil {
			valErr := field.Invalid(headerModifierPath.Child("remove"), h, err.Error())
			allErrs = append(allErrs, valErr)
		}
	}

	return allErrs
}
