package graph

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation/validationfakes"
)

func TestRouteGetAllBackendGroups(t *testing.T) {
	group0 := BackendGroup{
		RuleIdx: 0,
	}
	group1 := BackendGroup{
		RuleIdx: 1,
	}
	group2 := BackendGroup{
		RuleIdx: 2,
	}
	group3 := BackendGroup{
		RuleIdx: 3,
	}

	tests := []struct {
		route    *Route
		name     string
		expected []BackendGroup
	}{
		{
			route:    &Route{},
			expected: nil,
			name:     "no rules",
		},
		{
			route: &Route{
				Rules: []Rule{
					{
						ValidMatches: true,
						ValidFilters: true,
						BackendGroup: group0,
					},
					{
						ValidMatches: false,
						ValidFilters: true,
						BackendGroup: group1,
					},
					{
						ValidMatches: true,
						ValidFilters: false,
						BackendGroup: group2,
					},
					{
						ValidMatches: false,
						ValidFilters: false,
						BackendGroup: group3,
					},
				},
			},
			expected: []BackendGroup{group0},
			name:     "mix of valid and invalid rules",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			result := test.route.GetAllBackendGroups()
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestGetAllConditionsForSectionName(t *testing.T) {
	const (
		sectionName = "foo"
	)

	sectionNameRefs := map[string]ParentRef{
		sectionName: {
			Idx:     0,
			Gateway: types.NamespacedName{Namespace: "test", Name: "gateway"},
		},
	}

	tests := []struct {
		route    *Route
		name     string
		expected []conditions.Condition
	}{
		{
			route: &Route{
				SectionNameRefs: sectionNameRefs,
				Conditions:      nil,
			},
			expected: nil,
			name:     "no conditions",
		},
		{
			route: &Route{
				SectionNameRefs: sectionNameRefs,
				UnattachedSectionNameRefs: map[string]conditions.Condition{
					sectionName: conditions.NewTODO("unattached"),
				},
			},
			expected: []conditions.Condition{
				conditions.NewTODO("unattached"),
			},
			name: "unattached section",
		},
		{
			route: &Route{
				SectionNameRefs: sectionNameRefs,
				UnattachedSectionNameRefs: map[string]conditions.Condition{
					sectionName: conditions.NewTODO("unattached"),
				},
				Conditions: []conditions.Condition{conditions.NewTODO("route")},
			},
			expected: []conditions.Condition{
				conditions.NewTODO("unattached"),
				conditions.NewTODO("route"),
			},
			name: "unattached section and route",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			result := test.route.GetAllConditionsForSectionName(sectionName)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestGetAllConditionsForSectionNamePanics(t *testing.T) {
	route := &Route{
		SectionNameRefs: map[string]ParentRef{
			"foo": {
				Idx:     0,
				Gateway: types.NamespacedName{Namespace: "test", Name: "gateway"},
			},
		},
	}

	invoke := func() { _ = route.GetAllConditionsForSectionName("bar") }

	g := NewGomegaWithT(t)
	g.Expect(invoke).To(Panic())
}

func TestBuildRoutes(t *testing.T) {
	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	hr := createHTTPRoute("hr-1", gwNsName.Name, "example.com", "/")
	hrWrongGateway := createHTTPRoute("hr-2", "some-gateway", "example.com", "/")

	hrRoutes := map[types.NamespacedName]*v1beta1.HTTPRoute{
		client.ObjectKeyFromObject(hr):             hr,
		client.ObjectKeyFromObject(hrWrongGateway): hrWrongGateway,
	}

	tests := []struct {
		expected  map[types.NamespacedName]*Route
		name      string
		gwNsNames []types.NamespacedName
	}{
		{
			gwNsNames: []types.NamespacedName{gwNsName},
			expected: map[types.NamespacedName]*Route{
				client.ObjectKeyFromObject(hr): {
					Source: hr,
					SectionNameRefs: map[string]ParentRef{
						sectionNameOfCreateHTTPRoute: {
							Idx:     0,
							Gateway: gwNsName,
						},
					},
					UnattachedSectionNameRefs: map[string]conditions.Condition{},
					Valid:                     true,
					Rules: []Rule{
						{
							ValidMatches: true,
							ValidFilters: true,
							BackendGroup: BackendGroup{
								Source:  client.ObjectKeyFromObject(hr),
								RuleIdx: 0,
							},
						},
					},
				},
			},
			name: "normal case",
		},
		{
			gwNsNames: []types.NamespacedName{},
			expected:  nil,
			name:      "no gateways",
		},
	}

	validator := &validationfakes.FakeHTTPFieldsValidator{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			routes := buildRoutesForGateways(validator, hrRoutes, test.gwNsNames)
			g.Expect(helpers.Diff(test.expected, routes)).To(BeEmpty())
		})
	}
}

func TestBuildSectionNameRefs(t *testing.T) {
	gwNsName1 := types.NamespacedName{Namespace: "test", Name: "gateway-1"}
	gwNsName2 := types.NamespacedName{Namespace: "test", Name: "gateway-2"}

	parentRefs := []v1beta1.ParentReference{
		{
			Name:        v1beta1.ObjectName(gwNsName1.Name),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("one"),
		},
		{
			Name:        v1beta1.ObjectName("some-gateway"),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("other"),
		},
		{
			Name:        v1beta1.ObjectName(gwNsName2.Name),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("two"),
		},
	}

	gwNsNames := []types.NamespacedName{gwNsName1, gwNsName2}
	routeNamespace := "test"

	expected := map[string]ParentRef{
		"one": {
			Idx:     0,
			Gateway: gwNsName1,
		},
		"two": {
			Idx:     2,
			Gateway: gwNsName2,
		},
	}

	g := NewGomegaWithT(t)

	result := buildSectionNameRefs(parentRefs, routeNamespace, gwNsNames)
	g.Expect(result).To(Equal(expected))
}

func TestFindGatewayForParentRef(t *testing.T) {
	gwNsName1 := types.NamespacedName{Namespace: "test-1", Name: "gateway-1"}
	gwNsName2 := types.NamespacedName{Namespace: "test-2", Name: "gateway-2"}

	tests := []struct {
		ref              v1beta1.ParentReference
		expectedGwNsName types.NamespacedName
		name             string
		expectedFound    bool
	}{
		{
			ref: v1beta1.ParentReference{
				Namespace:   helpers.GetPointer(v1beta1.Namespace(gwNsName1.Namespace)),
				Name:        v1beta1.ObjectName(gwNsName1.Name),
				SectionName: helpers.GetPointer[v1beta1.SectionName]("one"),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName1,
			name:             "found",
		},
		{
			ref: v1beta1.ParentReference{
				Name:        v1beta1.ObjectName(gwNsName2.Name),
				SectionName: helpers.GetPointer[v1beta1.SectionName]("one"),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName2,
			name:             "found with implicit namespace",
		},
		{
			ref: v1beta1.ParentReference{
				Kind:        helpers.GetPointer[v1beta1.Kind]("NotGateway"),
				Name:        v1beta1.ObjectName(gwNsName2.Name),
				SectionName: helpers.GetPointer[v1beta1.SectionName]("one"),
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "wrong kind",
		},
		{
			ref: v1beta1.ParentReference{
				Group:       helpers.GetPointer[v1beta1.Group]("wrong-group"),
				Name:        v1beta1.ObjectName(gwNsName2.Name),
				SectionName: helpers.GetPointer[v1beta1.SectionName]("one"),
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "wrong group",
		},
	}

	routeNamespace := "test-2"

	gwNsNames := []types.NamespacedName{
		gwNsName1,
		gwNsName2,
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			gw, found := findGatewayForParentRef(test.ref, routeNamespace, gwNsNames)
			g.Expect(found).To(Equal(test.expectedFound))
			g.Expect(gw).To(Equal(test.expectedGwNsName))
		})
	}
}

const (
	sectionNameOfCreateHTTPRoute = "test-section"
)

func createHTTPRoute(
	name string,
	refName string,
	hostname v1beta1.Hostname,
	paths ...string,
) *v1beta1.HTTPRoute {
	rules := make([]v1beta1.HTTPRouteRule, 0, len(paths))

	for _, path := range paths {
		rules = append(rules, v1beta1.HTTPRouteRule{
			Matches: []v1beta1.HTTPRouteMatch{
				{
					Path: &v1beta1.HTTPPathMatch{
						Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
						Value: helpers.GetPointer(path),
					},
				},
			},
		})
	}

	return &v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      name,
		},
		Spec: v1beta1.HTTPRouteSpec{
			CommonRouteSpec: v1beta1.CommonRouteSpec{
				ParentRefs: []v1beta1.ParentReference{
					{
						Namespace:   helpers.GetPointer[v1beta1.Namespace]("test"),
						Name:        v1beta1.ObjectName(refName),
						SectionName: helpers.GetPointer[v1beta1.SectionName](sectionNameOfCreateHTTPRoute),
					},
				},
			},
			Hostnames: []v1beta1.Hostname{hostname},
			Rules:     rules,
		},
	}
}

func addFilterToPath(hr *v1beta1.HTTPRoute, path string, filter v1beta1.HTTPRouteFilter) {
	for i := range hr.Spec.Rules {
		for _, match := range hr.Spec.Rules[i].Matches {
			if match.Path == nil {
				panic("unexpected nil path")
			}
			if *match.Path.Value == path {
				hr.Spec.Rules[i].Filters = append(hr.Spec.Rules[i].Filters, filter)
			}
		}
	}
}

func TestBuildRoute(t *testing.T) {
	const (
		invalidPath             = "/invalid"
		invalidRedirectHostname = "invalid.example.com"
	)

	gatewayNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	validFilter := v1beta1.HTTPRouteFilter{
		Type:            v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{},
	}
	invalidFilter := v1beta1.HTTPRouteFilter{
		Type: v1beta1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
			Hostname: helpers.GetPointer[v1beta1.PreciseHostname](invalidRedirectHostname),
		},
	}

	hr := createHTTPRoute("hr", gatewayNsName.Name, "example.com", "/", "/filter")
	addFilterToPath(hr, "/filter", validFilter)

	hrInvalidHostname := createHTTPRoute("hr", gatewayNsName.Name, "", "/")
	hrNotNKG := createHTTPRoute("hr", "some-gateway", "example.com", "/")
	hrInvalidMatches := createHTTPRoute("hr", gatewayNsName.Name, "example.com", invalidPath)

	hrInvalidFilters := createHTTPRoute("hr", gatewayNsName.Name, "example.com", "/filter")
	addFilterToPath(hrInvalidFilters, "/filter", invalidFilter)

	hrInvalidValidRules := createHTTPRoute("hr", gatewayNsName.Name, "example.com", invalidPath, "/filter", "/")
	addFilterToPath(hrInvalidValidRules, "/filter", invalidFilter)

	validatorInvalidFieldsInRule := &validationfakes.FakeHTTPFieldsValidator{
		ValidatePathInPrefixMatchStub: func(path string) error {
			if path == invalidPath {
				return errors.New("invalid path")
			}
			return nil
		},
		ValidateRedirectHostnameStub: func(h string) error {
			if h == invalidRedirectHostname {
				return errors.New("invalid hostname")
			}
			return nil
		},
	}

	tests := []struct {
		validator *validationfakes.FakeHTTPFieldsValidator
		hr        *v1beta1.HTTPRoute
		expected  *Route
		name      string
	}{
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hr,
			expected: &Route{
				Source: hr,
				SectionNameRefs: map[string]ParentRef{
					sectionNameOfCreateHTTPRoute: {
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				UnattachedSectionNameRefs: map[string]conditions.Condition{},
				Valid:                     true,
				Rules: []Rule{
					{
						ValidMatches: true,
						ValidFilters: true,
						BackendGroup: BackendGroup{
							Source: types.NamespacedName{
								Namespace: hr.Namespace,
								Name:      hr.Name,
							},
							RuleIdx: 0,
						},
					},
					{
						ValidMatches: true,
						ValidFilters: true,
						BackendGroup: BackendGroup{
							Source: types.NamespacedName{
								Namespace: hr.Namespace,
								Name:      hr.Name,
							},
							RuleIdx: 1,
						},
					},
				},
			},
			name: "normal case",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrNotNKG,
			expected:  nil,
			name:      "not NKG route",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrInvalidHostname,
			expected: &Route{
				Source: hrInvalidHostname,
				Valid:  false,
				SectionNameRefs: map[string]ParentRef{
					sectionNameOfCreateHTTPRoute: {
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				UnattachedSectionNameRefs: map[string]conditions.Condition{},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(`spec.hostnames[0]: Invalid value: "": cannot be empty string`),
				},
			},
			name: "invalid hostname",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{
				ValidateHostnameInServerStub: func(string) error {
					return errors.New("invalid hostname")
				},
			},
			hr: hr,
			expected: &Route{
				Source: hr,
				Valid:  false,
				SectionNameRefs: map[string]ParentRef{
					sectionNameOfCreateHTTPRoute: {
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				UnattachedSectionNameRefs: map[string]conditions.Condition{},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(`spec.hostnames[0]: Invalid value: "example.com": invalid hostname`),
				},
			},
			name: "invalid hostname by the data-plane",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrInvalidMatches,
			expected: &Route{
				Source: hrInvalidMatches,
				Valid:  false,
				SectionNameRefs: map[string]ParentRef{
					sectionNameOfCreateHTTPRoute: {
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				UnattachedSectionNameRefs: map[string]conditions.Condition{},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].matches[0].path: Invalid value: "/invalid": invalid path`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
						BackendGroup: BackendGroup{
							Source:  client.ObjectKeyFromObject(hr),
							RuleIdx: 0,
						},
					},
				},
			},
			name: "all rules invalid, with invalid matches",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrInvalidFilters,
			expected: &Route{
				Source: hrInvalidFilters,
				Valid:  false,
				SectionNameRefs: map[string]ParentRef{
					sectionNameOfCreateHTTPRoute: {
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				UnattachedSectionNameRefs: map[string]conditions.Condition{},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].filters[0].requestRedirect.hostname: ` +
							`Invalid value: "invalid.example.com": invalid hostname`),
				},
				Rules: []Rule{
					{
						ValidMatches: true,
						ValidFilters: false,
						BackendGroup: BackendGroup{
							Source:  client.ObjectKeyFromObject(hr),
							RuleIdx: 0,
						},
					},
				},
			},
			name: "all rules invalid, with invalid filters",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrInvalidValidRules,
			expected: &Route{
				Source: hrInvalidValidRules,
				Valid:  true,
				SectionNameRefs: map[string]ParentRef{
					sectionNameOfCreateHTTPRoute: {
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				UnattachedSectionNameRefs: map[string]conditions.Condition{},
				Conditions: []conditions.Condition{
					conditions.NewTODO(
						`Some rules are invalid: ` +
							`[spec.rules[0].matches[0].path: Invalid value: "/invalid": invalid path, ` +
							`spec.rules[1].filters[0].requestRedirect.hostname: Invalid value: ` +
							`"invalid.example.com": invalid hostname]`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
						BackendGroup: BackendGroup{
							Source:  client.ObjectKeyFromObject(hrInvalidValidRules),
							RuleIdx: 0,
						},
					},
					{
						ValidMatches: true,
						ValidFilters: false,
						BackendGroup: BackendGroup{
							Source:  client.ObjectKeyFromObject(hrInvalidValidRules),
							RuleIdx: 1,
						},
					},
					{
						ValidMatches: true,
						ValidFilters: true,
						BackendGroup: BackendGroup{
							Source:  client.ObjectKeyFromObject(hrInvalidValidRules),
							RuleIdx: 2,
						},
					},
				},
			},
			name: "invalid with invalid and valid rules",
		},
	}

	gatewayNsNames := []types.NamespacedName{gatewayNsName}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			route := buildRoute(test.validator, test.hr, gatewayNsNames)
			g.Expect(helpers.Diff(test.expected, route)).To(BeEmpty())
		})
	}
}

func TestBindRouteToListeners(t *testing.T) {
	// we create a new listener each time because the function under test can modify it
	createListener := func() *Listener {
		return &Listener{
			Source: v1beta1.Listener{
				Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
			},
			Valid:             true,
			Routes:            map[types.NamespacedName]*Route{},
			AcceptedHostnames: map[string]struct{}{},
		}
	}
	createModifiedListener := func(m func(*Listener)) *Listener {
		l := createListener()
		m(l)
		return l
	}

	gw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway",
		},
	}

	createHTTPRouteWithSectionNameAndPort := func(
		sectionName *v1beta1.SectionName,
		port *v1beta1.PortNumber,
	) *v1beta1.HTTPRoute {
		return &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "hr",
			},
			Spec: v1beta1.HTTPRouteSpec{
				CommonRouteSpec: v1beta1.CommonRouteSpec{
					ParentRefs: []v1beta1.ParentReference{
						{
							Name:        v1beta1.ObjectName(gw.Name),
							SectionName: sectionName,
							Port:        port,
						},
					},
				},
				Hostnames: []v1beta1.Hostname{
					"foo.example.com",
				},
			},
		}
	}

	hr := createHTTPRouteWithSectionNameAndPort(helpers.GetPointer[v1beta1.SectionName]("listener-80-1"), nil)
	hrWithMissingSectionName := createHTTPRouteWithSectionNameAndPort(nil, nil)
	hrWithEmptySectionName := createHTTPRouteWithSectionNameAndPort(helpers.GetPointer[v1beta1.SectionName](""), nil)
	hrWithPort := createHTTPRouteWithSectionNameAndPort(
		helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
		helpers.GetPointer[v1beta1.PortNumber](80),
	)
	hrWithNonExistingListener := createHTTPRouteWithSectionNameAndPort(
		helpers.GetPointer[v1beta1.SectionName]("listener-80-2"),
		nil,
	)

	normalRoute := &Route{
		Source: hr,
		Valid:  true,
		SectionNameRefs: map[string]ParentRef{
			"listener-80-1": {
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}
	routeWithMissingSectionName := &Route{
		Source: hrWithMissingSectionName,
		Valid:  true,
		SectionNameRefs: map[string]ParentRef{
			"": {
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}
	routeWithEmptySectionName := &Route{
		Source: hrWithEmptySectionName,
		Valid:  true,
		SectionNameRefs: map[string]ParentRef{
			"": {
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}
	routeWithNonExistingListener := &Route{
		Source: hrWithNonExistingListener,
		Valid:  true,
		SectionNameRefs: map[string]ParentRef{
			"listener-80-2": {
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}
	routeWithPort := &Route{
		Source: hrWithPort,
		Valid:  true,
		SectionNameRefs: map[string]ParentRef{
			"listener-80-1": {
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}
	routeWithIgnoredGateway := &Route{
		Source: hr,
		Valid:  true,
		SectionNameRefs: map[string]ParentRef{
			"listener-80-1": {
				Idx:     0,
				Gateway: types.NamespacedName{Namespace: "test", Name: "ignored-gateway"},
			},
		},
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}
	notValidRoute := &Route{
		Valid:                     false,
		UnattachedSectionNameRefs: map[string]conditions.Condition{},
	}

	notValidListener := createModifiedListener(func(l *Listener) {
		l.Valid = false
	})
	nonMatchingHostnameListener := createModifiedListener(func(l *Listener) {
		l.Source.Hostname = helpers.GetPointer[v1beta1.Hostname]("bar.example.com")
	})

	tests := []struct {
		route                                  *Route
		gateway                                *Gateway
		expectedRouteUnattachedSectionNameRefs map[string]conditions.Condition
		expectedGatewayListeners               map[string]*Listener
		name                                   string
	}{
		{
			route: normalRoute,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener(func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): normalRoute,
					}
					l.AcceptedHostnames = map[string]struct{}{
						"foo.example.com": {},
					}
				}),
			},
			name: "normal case",
		},
		{
			route: routeWithMissingSectionName,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"": conditions.NewRouteUnsupportedValue(`spec.parentRefs[0].sectionName: Required value: cannot be empty`),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			name: "section name is missing",
		},
		{
			route: routeWithEmptySectionName,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"": conditions.NewRouteUnsupportedValue(`spec.parentRefs[0].sectionName: Required value: cannot be empty`),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			name: "section name is empty",
		},
		{
			route: routeWithPort,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"listener-80-1": conditions.NewRouteUnsupportedValue(`spec.parentRefs[0].port: Forbidden: cannot be set`),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			name: "port is configured",
		},
		{
			route: routeWithNonExistingListener,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"listener-80-2": conditions.NewTODO("listener is not found"),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			name: "listener doesn't exist",
		},
		{
			route: normalRoute,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": notValidListener,
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"listener-80-1": conditions.NewRouteInvalidListener(),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": notValidListener,
			},
			name: "listener isn't valid",
		},
		{
			route: normalRoute,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": nonMatchingHostnameListener,
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"listener-80-1": conditions.NewRouteNoMatchingListenerHostname(),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": nonMatchingHostnameListener,
			},
			name: "no matching listener hostname",
		},
		{
			route: routeWithIgnoredGateway,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{
				"listener-80-1": conditions.NewTODO("Gateway is ignored"),
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			name: "gateway is ignored",
		},
		{
			route: notValidRoute,
			gateway: &Gateway{
				Source: gw,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener(),
				},
			},
			expectedRouteUnattachedSectionNameRefs: map[string]conditions.Condition{},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener(),
			},
			name: "route isn't valid",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			bindRouteToListeners(test.route, test.gateway)

			g.Expect(test.route.UnattachedSectionNameRefs).To(Equal(test.expectedRouteUnattachedSectionNameRefs))
			g.Expect(helpers.Diff(test.gateway.Listeners, test.expectedGatewayListeners)).To(BeEmpty())
		})
	}
}

func TestFindAcceptedHostnames(t *testing.T) {
	var listenerHostnameFoo v1beta1.Hostname = "foo.example.com"
	var listenerHostnameCafe v1beta1.Hostname = "cafe.example.com"
	routeHostnames := []v1beta1.Hostname{"foo.example.com", "bar.example.com"}

	tests := []struct {
		listenerHostname *v1beta1.Hostname
		msg              string
		routeHostnames   []v1beta1.Hostname
		expected         []string
	}{
		{
			listenerHostname: &listenerHostnameFoo,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com"},
			msg:              "one match",
		},
		{
			listenerHostname: &listenerHostnameCafe,
			routeHostnames:   routeHostnames,
			expected:         nil,
			msg:              "no match",
		},
		{
			listenerHostname: nil,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com", "bar.example.com"},
			msg:              "nil listener hostname",
		},
	}

	for _, test := range tests {
		result := findAcceptedHostnames(test.listenerHostname, test.routeHostnames)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("findAcceptedHostnames() %q  mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}

func TestGetHostname(t *testing.T) {
	var emptyHostname v1beta1.Hostname
	var hostname v1beta1.Hostname = "example.com"

	tests := []struct {
		h        *v1beta1.Hostname
		expected string
		msg      string
	}{
		{
			h:        nil,
			expected: "",
			msg:      "nil hostname",
		},
		{
			h:        &emptyHostname,
			expected: "",
			msg:      "empty hostname",
		},
		{
			h:        &hostname,
			expected: string(hostname),
			msg:      "normal hostname",
		},
	}

	for _, test := range tests {
		result := getHostname(test.h)
		if result != test.expected {
			t.Errorf("getHostname() returned %q but expected %q for the case of %q", result, test.expected, test.msg)
		}
	}
}

func TestValidateHostnames(t *testing.T) {
	const validHostname = "example.com"

	tests := []struct {
		validator *validationfakes.FakeHTTPFieldsValidator
		name      string
		hostnames []v1beta1.Hostname
		expectErr bool
	}{
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hostnames: []v1beta1.Hostname{validHostname},
			expectErr: false,
			name:      "valid",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hostnames: []v1beta1.Hostname{
				validHostname,
				"",
			},
			expectErr: true,
			name:      "valid and invalid",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{
				ValidateHostnameInServerStub: func(h string) error {
					if h == validHostname {
						return nil
					}
					return errors.New("invalid hostname")
				},
			},
			hostnames: []v1beta1.Hostname{
				validHostname,
				"value", // invalid by the validator
			},
			expectErr: true,
			name:      "valid and invalid by the validator",
		},
	}

	path := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			err := validateHostnames(test.validator, test.hostnames, path)

			if test.expectErr {
				g.Expect(err).To(HaveOccurred())
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestValidateMatch(t *testing.T) {
	createAllValidValidator := func() *validationfakes.FakeHTTPFieldsValidator {
		v := &validationfakes.FakeHTTPFieldsValidator{}
		v.ValidateMethodInMatchReturns(true, nil)
		return v
	}

	tests := []struct {
		match          v1beta1.HTTPRouteMatch
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
					Value: helpers.GetPointer("/"),
				},
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(v1beta1.HeaderMatchExact),
						Name:  "header",
						Value: "x",
					},
				},
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(v1beta1.QueryParamMatchExact),
						Name:  "param",
						Value: "y",
					},
				},
				Method: helpers.GetPointer(v1beta1.HTTPMethodGet),
			},
			expectErrCount: 0,
			name:           "valid",
		},
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Type:  helpers.GetPointer(v1beta1.PathMatchRegularExpression),
					Value: helpers.GetPointer("/"),
				},
			},
			expectErrCount: 1,
			name:           "wrong path type",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidatePathInPrefixMatchReturns(errors.New("invalid path value"))
				return validator
			}(),
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
					Value: helpers.GetPointer("/"),
				},
			},
			expectErrCount: 1,
			name:           "wrong path value",
		},
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Type:  nil,
						Name:  "header",
						Value: "x",
					},
				},
			},
			expectErrCount: 1,
			name:           "header match type is nil",
		},
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(v1beta1.HeaderMatchRegularExpression),
						Name:  "header",
						Value: "x",
					},
				},
			},
			expectErrCount: 1,
			name:           "header match type is invalid",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateHeaderNameInMatchReturns(errors.New("invalid header name"))
				return validator
			}(),
			match: v1beta1.HTTPRouteMatch{
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(v1beta1.HeaderMatchExact),
						Name:  "header", // any value is invalid by the validator
						Value: "x",
					},
				},
			},
			expectErrCount: 1,
			name:           "header name is invalid",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateHeaderValueInMatchReturns(errors.New("invalid header value"))
				return validator
			}(),
			match: v1beta1.HTTPRouteMatch{
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(v1beta1.HeaderMatchExact),
						Name:  "header",
						Value: "x", // any value is invalid by the validator
					},
				},
			},
			expectErrCount: 1,
			name:           "header value is invalid",
		},
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Type:  nil,
						Name:  "param",
						Value: "y",
					},
				},
			},
			expectErrCount: 1,
			name:           "query param match type is nil",
		},
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(v1beta1.QueryParamMatchRegularExpression),
						Name:  "param",
						Value: "y",
					},
				},
			},
			expectErrCount: 1,
			name:           "query param match type is invalid",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateQueryParamNameInMatchReturns(errors.New("invalid query param name"))
				return validator
			}(),
			match: v1beta1.HTTPRouteMatch{
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(v1beta1.QueryParamMatchExact),
						Name:  "param", // any value is invalid by the validator
						Value: "y",
					},
				},
			},
			expectErrCount: 1,
			name:           "query param name is invalid",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateQueryParamValueInMatchReturns(errors.New("invalid query param value"))
				return validator
			}(),
			match: v1beta1.HTTPRouteMatch{
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(v1beta1.QueryParamMatchExact),
						Name:  "param",
						Value: "y", // any value is invalid by the validator
					},
				},
			},
			expectErrCount: 1,
			name:           "query param value is invalid",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateMethodInMatchReturns(false, []string{"VALID_METHOD"})
				return validator
			}(),
			match: v1beta1.HTTPRouteMatch{
				Method: helpers.GetPointer(v1beta1.HTTPMethodGet), // any value is invalid by the validator
			},
			expectErrCount: 1,
			name:           "method is invalid",
		},
		{
			validator: createAllValidValidator(),
			match: v1beta1.HTTPRouteMatch{
				Path: &v1beta1.HTTPPathMatch{
					Type:  helpers.GetPointer(v1beta1.PathMatchRegularExpression), // invalid
					Value: helpers.GetPointer("/"),
				},
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(v1beta1.HeaderMatchRegularExpression), // invalid
						Name:  "header",
						Value: "x",
					},
				},
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(v1beta1.QueryParamMatchRegularExpression), // invalid
						Name:  "param",
						Value: "y",
					},
				},
			},
			expectErrCount: 3,
			name:           "multiple errors",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			allErrs := validateMatch(test.validator, test.match, field.NewPath("test"))
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestValidateFilter(t *testing.T) {
	createAllValidValidator := func() *validationfakes.FakeHTTPFieldsValidator {
		v := &validationfakes.FakeHTTPFieldsValidator{}

		v.ValidateRedirectSchemeReturns(true, nil)
		v.ValidateRedirectStatusCodeReturns(true, nil)

		return v
	}

	tests := []struct {
		filter         v1beta1.HTTPRouteFilter
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: createAllValidValidator(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					Scheme:     helpers.GetPointer("http"),
					Hostname:   helpers.GetPointer[v1beta1.PreciseHostname]("example.com"),
					Port:       helpers.GetPointer[v1beta1.PortNumber](80),
					StatusCode: helpers.GetPointer(301),
				},
			},
			expectErrCount: 0,
			name:           "valid redirect filter",
		},
		{
			validator: createAllValidValidator(),
			filter: v1beta1.HTTPRouteFilter{
				Type:            v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{},
			},
			expectErrCount: 0,
			name:           "valid redirect filter with no fields set",
		},
		{
			validator: createAllValidValidator(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterURLRewrite,
			},
			expectErrCount: 1,
			name:           "unsupported filter",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateRedirectSchemeReturns(false, []string{"valid-scheme"})
				return validator
			}(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					Scheme: helpers.GetPointer("http"), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid scheme",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateRedirectHostnameReturns(errors.New("invalid hostname"))
				return validator
			}(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					Hostname: helpers.GetPointer[v1beta1.PreciseHostname]("example.com"), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid hostname",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateRedirectPortReturns(errors.New("invalid port"))
				return validator
			}(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					Port: helpers.GetPointer[v1beta1.PortNumber](80), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid port",
		},
		{
			validator: createAllValidValidator(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					Path: &v1beta1.HTTPPathModifier{},
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with unsupported path modifier",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateRedirectStatusCodeReturns(false, []string{"200"})
				return validator
			}(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					StatusCode: helpers.GetPointer(301), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid status code",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateRedirectHostnameReturns(errors.New("invalid hostname"))
				validator.ValidateRedirectPortReturns(errors.New("invalid port"))
				return validator
			}(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{
					Hostname: helpers.GetPointer[v1beta1.PreciseHostname]("example.com"), // any value is invalid by the validator
					Port:     helpers.GetPointer[v1beta1.PortNumber](80),                 // any value is invalid by the validator
				},
			},
			expectErrCount: 2,
			name:           "redirect filter with multiple errors",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			allErrs := validateFilter(test.validator, test.filter, filterPath)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}
