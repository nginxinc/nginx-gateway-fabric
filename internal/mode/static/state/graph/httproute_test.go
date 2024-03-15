package graph

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation/validationfakes"
)

const (
	sectionNameOfCreateHTTPRoute = "test-section"
	emptyPathType                = "/empty-type"
	emptyPathValue               = "/empty-value"
)

func createHTTPRoute(
	name string,
	refName string,
	hostname gatewayv1.Hostname,
	paths ...string,
) *gatewayv1.HTTPRoute {
	rules := make([]gatewayv1.HTTPRouteRule, 0, len(paths))
	pathType := helpers.GetPointer(gatewayv1.PathMatchPathPrefix)

	for _, path := range paths {
		if path == emptyPathType {
			pathType = nil
		}
		pathValue := helpers.GetPointer(path)
		if path == emptyPathValue {
			pathValue = nil
		}
		rules = append(rules, gatewayv1.HTTPRouteRule{
			Matches: []gatewayv1.HTTPRouteMatch{
				{
					Path: &gatewayv1.HTTPPathMatch{
						Type:  pathType,
						Value: pathValue,
					},
				},
			},
		})
	}

	return &gatewayv1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      name,
		},
		Spec: gatewayv1.HTTPRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: []gatewayv1.ParentReference{
					{
						Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
						Name:        gatewayv1.ObjectName(refName),
						SectionName: helpers.GetPointer[gatewayv1.SectionName](sectionNameOfCreateHTTPRoute),
					},
				},
			},
			Hostnames: []gatewayv1.Hostname{hostname},
			Rules:     rules,
		},
	}
}

func addFilterToPath(hr *gatewayv1.HTTPRoute, path string, filter gatewayv1.HTTPRouteFilter) {
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

func TestBuildRoutes(t *testing.T) {
	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	hr := createHTTPRoute("hr-1", gwNsName.Name, "example.com", "/")
	hrWrongGateway := createHTTPRoute("hr-2", "some-gateway", "example.com", "/")

	hrRoutes := map[types.NamespacedName]*gatewayv1.HTTPRoute{
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
					ParentRefs: []ParentRef{
						{
							Idx:     0,
							Gateway: gwNsName,
						},
					},
					Valid:      true,
					Attachable: true,
					Rules: []Rule{
						{
							ValidMatches: true,
							ValidFilters: true,
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
			g := NewWithT(t)
			routes := buildRoutesForGateways(validator, hrRoutes, test.gwNsNames)
			g.Expect(helpers.Diff(test.expected, routes)).To(BeEmpty())
		})
	}
}

func TestBuildSectionNameRefs(t *testing.T) {
	const routeNamespace = "test"

	gwNsName1 := types.NamespacedName{Namespace: routeNamespace, Name: "gateway-1"}
	gwNsName2 := types.NamespacedName{Namespace: routeNamespace, Name: "gateway-2"}

	parentRefs := []gatewayv1.ParentReference{
		{
			Name:        gatewayv1.ObjectName(gwNsName1.Name),
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("one"),
		},
		{
			Name:        gatewayv1.ObjectName("some-other-gateway"),
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("two"),
		},
		{
			Name:        gatewayv1.ObjectName(gwNsName2.Name),
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("three"),
		},
		{
			Name:        gatewayv1.ObjectName(gwNsName1.Name),
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("same-name"),
		},
		{
			Name:        gatewayv1.ObjectName(gwNsName2.Name),
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("same-name"),
		},
		{
			Name:        gatewayv1.ObjectName("some-other-gateway"),
			SectionName: helpers.GetPointer[gatewayv1.SectionName]("same-name"),
		},
	}

	gwNsNames := []types.NamespacedName{gwNsName1, gwNsName2}

	expected := []ParentRef{
		{
			Idx:     0,
			Gateway: gwNsName1,
		},
		{
			Idx:     2,
			Gateway: gwNsName2,
		},
		{
			Idx:     3,
			Gateway: gwNsName1,
		},
		{
			Idx:     4,
			Gateway: gwNsName2,
		},
	}

	tests := []struct {
		expectedError error
		name          string
		parentRefs    []gatewayv1.ParentReference
		expectedRefs  []ParentRef
	}{
		{
			name:          "normal case",
			parentRefs:    parentRefs,
			expectedRefs:  expected,
			expectedError: nil,
		},
		{
			parentRefs: []gatewayv1.ParentReference{
				{
					Name:        gatewayv1.ObjectName(gwNsName1.Name),
					SectionName: helpers.GetPointer[gatewayv1.SectionName]("http"),
				},
				{
					Name:        gatewayv1.ObjectName(gwNsName1.Name),
					SectionName: helpers.GetPointer[gatewayv1.SectionName]("http"),
				},
			},
			name:          "duplicate sectionNames",
			expectedError: errors.New("duplicate section name \"http\" for Gateway test/gateway-1"),
		},
		{
			parentRefs: []gatewayv1.ParentReference{
				{
					Name:        gatewayv1.ObjectName(gwNsName1.Name),
					SectionName: nil,
				},
				{
					Name:        gatewayv1.ObjectName(gwNsName1.Name),
					SectionName: nil,
				},
			},
			name:          "nil sectionNames",
			expectedError: errors.New("duplicate section name \"\" for Gateway test/gateway-1"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result, err := buildSectionNameRefs(test.parentRefs, routeNamespace, gwNsNames)
			g.Expect(result).To(Equal(test.expectedRefs))
			if test.expectedError != nil {
				g.Expect(err).To(Equal(test.expectedError))
			} else {
				g.Expect(err).ToNot(HaveOccurred())
			}
		})
	}
}

func TestFindGatewayForParentRef(t *testing.T) {
	gwNsName1 := types.NamespacedName{Namespace: "test-1", Name: "gateway-1"}
	gwNsName2 := types.NamespacedName{Namespace: "test-2", Name: "gateway-2"}

	tests := []struct {
		ref              gatewayv1.ParentReference
		expectedGwNsName types.NamespacedName
		name             string
		expectedFound    bool
	}{
		{
			ref: gatewayv1.ParentReference{
				Namespace: helpers.GetPointer(gatewayv1.Namespace(gwNsName1.Namespace)),
				Name:      gatewayv1.ObjectName(gwNsName1.Name),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName1,
			name:             "found",
		},
		{
			ref: gatewayv1.ParentReference{
				Group:     helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName),
				Kind:      helpers.GetPointer[gatewayv1.Kind]("Gateway"),
				Namespace: helpers.GetPointer(gatewayv1.Namespace(gwNsName1.Namespace)),
				Name:      gatewayv1.ObjectName(gwNsName1.Name),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName1,
			name:             "found with explicit group and kind",
		},
		{
			ref: gatewayv1.ParentReference{
				Name: gatewayv1.ObjectName(gwNsName2.Name),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName2,
			name:             "found with implicit namespace",
		},
		{
			ref: gatewayv1.ParentReference{
				Kind: helpers.GetPointer[gatewayv1.Kind]("NotGateway"),
				Name: gatewayv1.ObjectName(gwNsName2.Name),
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "wrong kind",
		},
		{
			ref: gatewayv1.ParentReference{
				Group: helpers.GetPointer[gatewayv1.Group]("wrong-group"),
				Name:  gatewayv1.ObjectName(gwNsName2.Name),
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "wrong group",
		},
		{
			ref: gatewayv1.ParentReference{
				Namespace: helpers.GetPointer(gatewayv1.Namespace(gwNsName1.Namespace)),
				Name:      "some-gateway",
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "not found",
		},
	}

	routeNamespace := "test-2"

	gwNsNames := []types.NamespacedName{
		gwNsName1,
		gwNsName2,
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			gw, found := findGatewayForParentRef(test.ref, routeNamespace, gwNsNames)
			g.Expect(found).To(Equal(test.expectedFound))
			g.Expect(gw).To(Equal(test.expectedGwNsName))
		})
	}
}

func TestBuildRoute(t *testing.T) {
	const (
		invalidPath             = "/invalid"
		invalidRedirectHostname = "invalid.example.com"
	)

	gatewayNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	validFilter := gatewayv1.HTTPRouteFilter{
		Type:            gatewayv1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{},
	}
	invalidFilter := gatewayv1.HTTPRouteFilter{
		Type: gatewayv1.HTTPRouteFilterRequestRedirect,
		RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
			Hostname: helpers.GetPointer[gatewayv1.PreciseHostname](invalidRedirectHostname),
		},
	}

	hr := createHTTPRoute("hr", gatewayNsName.Name, "example.com", "/", "/filter")
	addFilterToPath(hr, "/filter", validFilter)

	hrInvalidHostname := createHTTPRoute("hr", gatewayNsName.Name, "", "/")
	hrNotNGF := createHTTPRoute("hr", "some-gateway", "example.com", "/")
	hrInvalidMatches := createHTTPRoute("hr", gatewayNsName.Name, "example.com", invalidPath)

	hrInvalidMatchesEmptyPathType := createHTTPRoute("hr", gatewayNsName.Name, "example.com", emptyPathType)
	hrInvalidMatchesEmptyPathValue := createHTTPRoute("hr", gatewayNsName.Name, "example.com", emptyPathValue)

	hrInvalidFilters := createHTTPRoute("hr", gatewayNsName.Name, "example.com", "/filter")
	addFilterToPath(hrInvalidFilters, "/filter", invalidFilter)

	hrDroppedInvalidMatches := createHTTPRoute("hr", gatewayNsName.Name, "example.com", invalidPath, "/")

	hrDroppedInvalidMatchesAndInvalidFilters := createHTTPRoute(
		"hr",
		gatewayNsName.Name,
		"example.com",
		invalidPath, "/filter", "/")
	addFilterToPath(hrDroppedInvalidMatchesAndInvalidFilters, "/filter", invalidFilter)

	hrDroppedInvalidFilters := createHTTPRoute("hr", gatewayNsName.Name, "example.com", "/filter", "/")
	addFilterToPath(hrDroppedInvalidFilters, "/filter", validFilter)
	addFilterToPath(hrDroppedInvalidFilters, "/", invalidFilter)

	hrDuplicateSectionName := createHTTPRoute("hr", gatewayNsName.Name, "example.com", "/")
	hrDuplicateSectionName.Spec.ParentRefs = append(
		hrDuplicateSectionName.Spec.ParentRefs,
		hrDuplicateSectionName.Spec.ParentRefs[0],
	)

	validatorInvalidFieldsInRule := &validationfakes.FakeHTTPFieldsValidator{
		ValidatePathInMatchStub: func(path string) error {
			if path == invalidPath {
				return errors.New("invalid path")
			}
			return nil
		},
		ValidateHostnameStub: func(h string) error {
			if h == invalidRedirectHostname {
				return errors.New("invalid hostname")
			}
			return nil
		},
	}

	tests := []struct {
		validator *validationfakes.FakeHTTPFieldsValidator
		hr        *gatewayv1.HTTPRoute
		expected  *Route
		name      string
	}{
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hr,
			expected: &Route{
				Source: hr,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Valid:      true,
				Attachable: true,
				Rules: []Rule{
					{
						ValidMatches: true,
						ValidFilters: true,
					},
					{
						ValidMatches: true,
						ValidFilters: true,
					},
				},
			},
			name: "normal case",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrInvalidMatchesEmptyPathType,
			expected: &Route{
				Source:     hrInvalidMatchesEmptyPathType,
				Valid:      false,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].matches[0].path.type: Required value: path type cannot be nil`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
					},
				},
			},
			name: "invalid matches with empty path type",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrDuplicateSectionName,
			expected: &Route{
				Source: hrDuplicateSectionName,
			},
			name: "invalid route with duplicate sectionName",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrInvalidMatchesEmptyPathValue,
			expected: &Route{
				Source:     hrInvalidMatchesEmptyPathValue,
				Valid:      false,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].matches[0].path.value: Required value: path value cannot be nil`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
					},
				},
			},
			name: "invalid matches with empty path value",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrNotNGF,
			expected:  nil,
			name:      "not NGF route",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			hr:        hrInvalidHostname,
			expected: &Route{
				Source:     hrInvalidHostname,
				Valid:      false,
				Attachable: false,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`spec.hostnames[0]: Invalid value: "": cannot be empty string`,
					),
				},
			},
			name: "invalid hostname",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrInvalidMatches,
			expected: &Route{
				Source:     hrInvalidMatches,
				Valid:      false,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].matches[0].path.value: Invalid value: "/invalid": invalid path`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
					},
				},
			},
			name: "all rules invalid, with invalid matches",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrInvalidFilters,
			expected: &Route{
				Source:     hrInvalidFilters,
				Valid:      false,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].filters[0].requestRedirect.hostname: ` +
							`Invalid value: "invalid.example.com": invalid hostname`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: true,
						ValidFilters: false,
					},
				},
			},
			name: "all rules invalid, with invalid filters",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrDroppedInvalidMatches,
			expected: &Route{
				Source:     hrDroppedInvalidMatches,
				Valid:      true,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRoutePartiallyInvalid(
						`spec.rules[0].matches[0].path.value: Invalid value: "/invalid": invalid path`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
					},
					{
						ValidMatches: true,
						ValidFilters: true,
					},
				},
			},
			name: "dropped invalid rule with invalid matches",
		},

		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrDroppedInvalidMatchesAndInvalidFilters,
			expected: &Route{
				Source:     hrDroppedInvalidMatchesAndInvalidFilters,
				Valid:      true,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRoutePartiallyInvalid(
						`[spec.rules[0].matches[0].path.value: Invalid value: "/invalid": invalid path, ` +
							`spec.rules[1].filters[0].requestRedirect.hostname: Invalid value: ` +
							`"invalid.example.com": invalid hostname]`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: false,
						ValidFilters: true,
					},
					{
						ValidMatches: true,
						ValidFilters: false,
					},
					{
						ValidMatches: true,
						ValidFilters: true,
					},
				},
			},
			name: "dropped invalid rule with invalid filters and invalid rule with invalid matches",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrDroppedInvalidFilters,
			expected: &Route{
				Source:     hrDroppedInvalidFilters,
				Valid:      true,
				Attachable: true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRoutePartiallyInvalid(
						`spec.rules[1].filters[0].requestRedirect.hostname: Invalid value: ` +
							`"invalid.example.com": invalid hostname`,
					),
				},
				Rules: []Rule{
					{
						ValidMatches: true,
						ValidFilters: true,
					},
					{
						ValidMatches: true,
						ValidFilters: false,
					},
				},
			},
			name: "dropped invalid rule with invalid filters",
		},
	}

	gatewayNsNames := []types.NamespacedName{gatewayNsName}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			route := buildRoute(test.validator, test.hr, gatewayNsNames)
			g.Expect(helpers.Diff(test.expected, route)).To(BeEmpty())
		})
	}
}

func TestBindRouteToListeners(t *testing.T) {
	// we create a new listener each time because the function under test can modify it
	createListener := func(name string) *Listener {
		return &Listener{
			Name: name,
			Source: gatewayv1.Listener{
				Name:     gatewayv1.SectionName(name),
				Hostname: (*gatewayv1.Hostname)(helpers.GetPointer("foo.example.com")),
			},
			Valid:      true,
			Attachable: true,
			Routes:     map[types.NamespacedName]*Route{},
		}
	}
	createModifiedListener := func(name string, m func(*Listener)) *Listener {
		l := createListener(name)
		m(l)
		return l
	}

	gw := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway",
		},
	}
	gwDiffNamespace := &gatewayv1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "diff-namespace",
			Name:      "gateway",
		},
	}

	createHTTPRouteWithSectionNameAndPort := func(
		sectionName *gatewayv1.SectionName,
		port *gatewayv1.PortNumber,
	) *gatewayv1.HTTPRoute {
		return &gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "hr",
			},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Name:        gatewayv1.ObjectName(gw.Name),
							SectionName: sectionName,
							Port:        port,
						},
					},
				},
				Hostnames: []gatewayv1.Hostname{
					"foo.example.com",
				},
			},
		}
	}

	hr := createHTTPRouteWithSectionNameAndPort(helpers.GetPointer[gatewayv1.SectionName]("listener-80-1"), nil)
	hrWithNilSectionName := createHTTPRouteWithSectionNameAndPort(nil, nil)
	hrWithEmptySectionName := createHTTPRouteWithSectionNameAndPort(helpers.GetPointer[gatewayv1.SectionName](""), nil)
	hrWithPort := createHTTPRouteWithSectionNameAndPort(
		helpers.GetPointer[gatewayv1.SectionName]("listener-80-1"),
		helpers.GetPointer[gatewayv1.PortNumber](80),
	)
	hrWithNonExistingListener := createHTTPRouteWithSectionNameAndPort(
		helpers.GetPointer[gatewayv1.SectionName]("listener-80-2"),
		nil,
	)

	var normalRoute *Route
	createNormalRoute := func(gateway *gatewayv1.Gateway) *Route {
		normalRoute = &Route{
			Source:     hr,
			Valid:      true,
			Attachable: true,
			ParentRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gateway),
				},
			},
		}
		return normalRoute
	}
	getLastNormalRoute := func() *Route {
		return normalRoute
	}

	invalidAttachableRoute1 := &Route{
		Source:     hr,
		Valid:      false,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	invalidAttachableRoute2 := &Route{
		Source:     hr,
		Valid:      false,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}

	routeWithMissingSectionName := &Route{
		Source:     hrWithNilSectionName,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	routeWithEmptySectionName := &Route{
		Source:     hrWithEmptySectionName,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	routeWithNonExistingListener := &Route{
		Source:     hrWithNonExistingListener,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	routeWithPort := &Route{
		Source:     hrWithPort,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	ignoredGwNsName := types.NamespacedName{Namespace: "test", Name: "ignored-gateway"}
	routeWithIgnoredGateway := &Route{
		Source:     hr,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: ignoredGwNsName,
			},
		},
	}
	invalidRoute := &Route{
		Valid: false,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}

	invalidNotAttachableListener := createModifiedListener("listener-80-1", func(l *Listener) {
		l.Valid = false
		l.Attachable = false
	})
	nonMatchingHostnameListener := createModifiedListener("listener-80-1", func(l *Listener) {
		l.Source.Hostname = helpers.GetPointer[gatewayv1.Hostname]("bar.example.com")
	})

	tests := []struct {
		route                    *Route
		gateway                  *Gateway
		expectedGatewayListeners []*Listener
		name                     string
		expectedSectionNameRefs  []ParentRef
		expectedConditions       []conditions.Condition
	}{
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): getLastNormalRoute(),
					}
				}),
			},
			name: "normal case",
		},
		{
			route: routeWithMissingSectionName,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): routeWithMissingSectionName,
					}
				}),
			},
			name: "section name is nil",
		},
		{
			route: routeWithEmptySectionName,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80"),
					createListener("listener-8080"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80":   {"foo.example.com"},
							"listener-8080": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80", func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): routeWithEmptySectionName,
					}
				}),
				createModifiedListener("listener-8080", func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): routeWithEmptySectionName,
					}
				}),
			},
			name: "section name is empty; bind to multiple listeners",
		},
		{
			route: routeWithEmptySectionName,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					invalidNotAttachableListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteInvalidListener(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				invalidNotAttachableListener,
			},
			name: "empty section name with no valid and attachable listeners",
		},
		{
			route: routeWithPort,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: false,
						FailedCondition: staticConds.NewRouteUnsupportedValue(
							`spec.parentRefs[0].port: Forbidden: cannot be set`,
						),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createListener("listener-80-1"),
			},
			name: "port is configured",
		},
		{
			route: routeWithNonExistingListener,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteNoMatchingParent(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createListener("listener-80-1"),
			},
			name: "listener doesn't exist",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					invalidNotAttachableListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteInvalidListener(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				invalidNotAttachableListener,
			},
			name: "listener isn't valid and attachable",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					nonMatchingHostnameListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteNoMatchingListenerHostname(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				nonMatchingHostnameListener,
			},
			name: "no matching listener hostname",
		},
		{
			route: routeWithIgnoredGateway,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: ignoredGwNsName,
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewTODO("Gateway is ignored"),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createListener("listener-80-1"),
			},
			name: "gateway is ignored",
		},
		{
			route: invalidRoute,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:        0,
					Gateway:    client.ObjectKeyFromObject(gw),
					Attachment: nil,
				},
			},
			expectedGatewayListeners: []*Listener{
				createListener("listener-80-1"),
			},
			name: "route isn't valid",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  false,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteInvalidGateway(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createListener("listener-80-1"),
			},
			name: "invalid gateway",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Valid = false
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Valid = false
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): getLastNormalRoute(),
					}
				}),
			},
			expectedConditions: []conditions.Condition{staticConds.NewRouteInvalidListener()},
			name:               "invalid attachable listener",
		},
		{
			route: invalidAttachableRoute1,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): invalidAttachableRoute1,
					}
				}),
			},
			name: "invalid attachable route",
		},
		{
			route: invalidAttachableRoute2,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Valid = false
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Valid = false
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): invalidAttachableRoute2,
					}
				}),
			},
			expectedConditions: []conditions.Condition{staticConds.NewRouteInvalidListener()},
			name:               "invalid attachable listener with invalid attachable route",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
							Namespaces: &gatewayv1.RouteNamespaces{
								From: helpers.GetPointer(gatewayv1.NamespacesFromSelector),
							},
						}
						allowedLabels := map[string]string{"app": "not-allowed"}
						l.AllowedRouteLabelSelector = labels.SelectorFromSet(allowedLabels)
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteNotAllowedByListeners(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: helpers.GetPointer(gatewayv1.NamespacesFromSelector),
						},
					}
					allowedLabels := map[string]string{"app": "not-allowed"}
					l.AllowedRouteLabelSelector = labels.SelectorFromSet(allowedLabels)
				}),
			},
			name: "route not allowed via labels",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
							Namespaces: &gatewayv1.RouteNamespaces{
								From: helpers.GetPointer(gatewayv1.NamespacesFromSelector),
							},
						}
						allowedLabels := map[string]string{"app": "allowed"}
						l.AllowedRouteLabelSelector = labels.SelectorFromSet(allowedLabels)
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					allowedLabels := map[string]string{"app": "allowed"}
					l.AllowedRouteLabelSelector = labels.SelectorFromSet(allowedLabels)
					l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: helpers.GetPointer(gatewayv1.NamespacesFromSelector),
						},
					}
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): getLastNormalRoute(),
					}
				}),
			},
			name: "route allowed via labels",
		},
		{
			route: createNormalRoute(gwDiffNamespace),
			gateway: &Gateway{
				Source: gwDiffNamespace,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
							Namespaces: &gatewayv1.RouteNamespaces{
								From: helpers.GetPointer(gatewayv1.NamespacesFromSame),
							},
						}
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gwDiffNamespace),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteNotAllowedByListeners(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: helpers.GetPointer(gatewayv1.NamespacesFromSame),
						},
					}
				}),
			},
			name: "route not allowed via same namespace",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
							Namespaces: &gatewayv1.RouteNamespaces{
								From: helpers.GetPointer(gatewayv1.NamespacesFromSame),
							},
						}
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: helpers.GetPointer(gatewayv1.NamespacesFromSame),
						},
					}
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): getLastNormalRoute(),
					}
				}),
			},
			name: "route allowed via same namespace",
		},
		{
			route: createNormalRoute(gwDiffNamespace),
			gateway: &Gateway{
				Source: gwDiffNamespace,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
							Namespaces: &gatewayv1.RouteNamespaces{
								From: helpers.GetPointer(gatewayv1.NamespacesFromAll),
							},
						}
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gwDiffNamespace),
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
						AcceptedHostnames: map[string][]string{
							"listener-80-1": {"foo.example.com"},
						},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
						Namespaces: &gatewayv1.RouteNamespaces{
							From: helpers.GetPointer(gatewayv1.NamespacesFromAll),
						},
					}
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): getLastNormalRoute(),
					}
				}),
			},
			name: "route allowed via all namespaces",
		},
	}

	namespaces := map[types.NamespacedName]*v1.Namespace{
		{Name: "test"}: {
			ObjectMeta: metav1.ObjectMeta{
				Name:   "test",
				Labels: map[string]string{"app": "allowed"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			bindRouteToListeners(test.route, test.gateway, namespaces)

			g.Expect(test.route.ParentRefs).To(Equal(test.expectedSectionNameRefs))
			g.Expect(helpers.Diff(test.gateway.Listeners, test.expectedGatewayListeners)).To(BeEmpty())
			g.Expect(helpers.Diff(test.route.Conditions, test.expectedConditions)).To(BeEmpty())
		})
	}
}

func TestFindAcceptedHostnames(t *testing.T) {
	var listenerHostnameFoo gatewayv1.Hostname = "foo.example.com"
	var listenerHostnameCafe gatewayv1.Hostname = "cafe.example.com"
	var listenerHostnameWildcard gatewayv1.Hostname = "*.example.com"
	routeHostnames := []gatewayv1.Hostname{"foo.example.com", "bar.example.com"}

	tests := []struct {
		listenerHostname *gatewayv1.Hostname
		msg              string
		routeHostnames   []gatewayv1.Hostname
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
		{
			listenerHostname: &listenerHostnameFoo,
			routeHostnames:   nil,
			expected:         []string{"foo.example.com"},
			msg:              "route has empty hostnames",
		},
		{
			listenerHostname: nil,
			routeHostnames:   nil,
			expected:         []string{wildcardHostname},
			msg:              "both listener and route have empty hostnames",
		},
		{
			listenerHostname: &listenerHostnameWildcard,
			routeHostnames:   routeHostnames,
			expected:         []string{"foo.example.com", "bar.example.com"},
			msg:              "listener wildcard hostname",
		},
		{
			listenerHostname: &listenerHostnameFoo,
			routeHostnames:   []gatewayv1.Hostname{"*.example.com"},
			expected:         []string{"foo.example.com"},
			msg:              "route wildcard hostname; specific listener hostname",
		},
		{
			listenerHostname: &listenerHostnameWildcard,
			routeHostnames:   nil,
			expected:         []string{"*.example.com"},
			msg:              "listener wildcard hostname; nil route hostname",
		},
		{
			listenerHostname: nil,
			routeHostnames:   []gatewayv1.Hostname{"*.example.com"},
			expected:         []string{"*.example.com"},
			msg:              "route wildcard hostname; nil listener hostname",
		},
		{
			listenerHostname: &listenerHostnameWildcard,
			routeHostnames:   []gatewayv1.Hostname{"*.bar.example.com"},
			expected:         []string{"*.bar.example.com"},
			msg:              "route and listener wildcard hostnames",
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)
			result := findAcceptedHostnames(test.listenerHostname, test.routeHostnames)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestGetHostname(t *testing.T) {
	var emptyHostname gatewayv1.Hostname
	var hostname gatewayv1.Hostname = "example.com"

	tests := []struct {
		h        *gatewayv1.Hostname
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
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)
			result := getHostname(test.h)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestValidateHostnames(t *testing.T) {
	const validHostname = "example.com"

	tests := []struct {
		name      string
		hostnames []gatewayv1.Hostname
		expectErr bool
	}{
		{
			hostnames: []gatewayv1.Hostname{
				validHostname,
				"example.org",
				"foo.example.net",
			},
			expectErr: false,
			name:      "multiple valid",
		},
		{
			hostnames: []gatewayv1.Hostname{
				validHostname,
				"",
			},
			expectErr: true,
			name:      "valid and invalid",
		},
	}

	path := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			err := validateHostnames(test.hostnames, path)

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
		match          gatewayv1.HTTPRouteMatch
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: createAllValidValidator(),
			match: gatewayv1.HTTPRouteMatch{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  helpers.GetPointer(gatewayv1.PathMatchPathPrefix),
					Value: helpers.GetPointer("/"),
				},
				Headers: []gatewayv1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.HeaderMatchExact),
						Name:  "header",
						Value: "x",
					},
				},
				QueryParams: []gatewayv1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.QueryParamMatchExact),
						Name:  "param",
						Value: "y",
					},
				},
				Method: helpers.GetPointer(gatewayv1.HTTPMethodGet),
			},
			expectErrCount: 0,
			name:           "valid",
		},
		{
			validator: createAllValidValidator(),
			match: gatewayv1.HTTPRouteMatch{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  helpers.GetPointer(gatewayv1.PathMatchExact),
					Value: helpers.GetPointer("/"),
				},
			},
			expectErrCount: 0,
			name:           "valid exact match",
		},
		{
			validator: createAllValidValidator(),
			match: gatewayv1.HTTPRouteMatch{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  helpers.GetPointer(gatewayv1.PathMatchRegularExpression),
					Value: helpers.GetPointer("/"),
				},
			},
			expectErrCount: 1,
			name:           "wrong path type",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidatePathInMatchReturns(errors.New("invalid path value"))
				return validator
			}(),
			match: gatewayv1.HTTPRouteMatch{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  helpers.GetPointer(gatewayv1.PathMatchPathPrefix),
					Value: helpers.GetPointer("/"),
				},
			},
			expectErrCount: 1,
			name:           "wrong path value",
		},
		{
			validator: createAllValidValidator(),
			match: gatewayv1.HTTPRouteMatch{
				Headers: []gatewayv1.HTTPHeaderMatch{
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
			match: gatewayv1.HTTPRouteMatch{
				Headers: []gatewayv1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.HeaderMatchRegularExpression),
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
			match: gatewayv1.HTTPRouteMatch{
				Headers: []gatewayv1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.HeaderMatchExact),
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
			match: gatewayv1.HTTPRouteMatch{
				Headers: []gatewayv1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.HeaderMatchExact),
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
			match: gatewayv1.HTTPRouteMatch{
				QueryParams: []gatewayv1.HTTPQueryParamMatch{
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
			match: gatewayv1.HTTPRouteMatch{
				QueryParams: []gatewayv1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.QueryParamMatchRegularExpression),
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
			match: gatewayv1.HTTPRouteMatch{
				QueryParams: []gatewayv1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.QueryParamMatchExact),
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
			match: gatewayv1.HTTPRouteMatch{
				QueryParams: []gatewayv1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.QueryParamMatchExact),
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
			match: gatewayv1.HTTPRouteMatch{
				Method: helpers.GetPointer(gatewayv1.HTTPMethodGet), // any value is invalid by the validator
			},
			expectErrCount: 1,
			name:           "method is invalid",
		},
		{
			validator: createAllValidValidator(),
			match: gatewayv1.HTTPRouteMatch{
				Path: &gatewayv1.HTTPPathMatch{
					Type:  helpers.GetPointer(gatewayv1.PathMatchRegularExpression), // invalid
					Value: helpers.GetPointer("/"),
				},
				Headers: []gatewayv1.HTTPHeaderMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.HeaderMatchRegularExpression), // invalid
						Name:  "header",
						Value: "x",
					},
				},
				QueryParams: []gatewayv1.HTTPQueryParamMatch{
					{
						Type:  helpers.GetPointer(gatewayv1.QueryParamMatchRegularExpression), // invalid
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
			g := NewWithT(t)
			allErrs := validateMatch(test.validator, test.match, field.NewPath("test"))
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestValidateFilter(t *testing.T) {
	tests := []struct {
		filter         gatewayv1.HTTPRouteFilter
		name           string
		expectErrCount int
	}{
		{
			filter: gatewayv1.HTTPRouteFilter{
				Type:            gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{},
			},
			expectErrCount: 0,
			name:           "valid redirect filter",
		},
		{
			filter: gatewayv1.HTTPRouteFilter{
				Type:       gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{},
			},
			expectErrCount: 0,
			name:           "valid rewrite filter",
		},
		{
			filter: gatewayv1.HTTPRouteFilter{
				Type:                  gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{},
			},
			expectErrCount: 0,
			name:           "valid request header modifiers filter",
		},
		{
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestMirror,
			},
			expectErrCount: 1,
			name:           "unsupported filter",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			allErrs := validateFilter(&validationfakes.FakeHTTPFieldsValidator{}, test.filter, filterPath)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestValidateFilterRedirect(t *testing.T) {
	createAllValidValidator := func() *validationfakes.FakeHTTPFieldsValidator {
		v := &validationfakes.FakeHTTPFieldsValidator{}

		v.ValidateRedirectSchemeReturns(true, nil)
		v.ValidateRedirectStatusCodeReturns(true, nil)

		return v
	}

	tests := []struct {
		filter         gatewayv1.HTTPRouteFilter
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			filter: gatewayv1.HTTPRouteFilter{
				Type:            gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: nil,
			},
			name:           "nil filter",
			expectErrCount: 1,
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Scheme:     helpers.GetPointer("http"),
					Hostname:   helpers.GetPointer[gatewayv1.PreciseHostname]("example.com"),
					Port:       helpers.GetPointer[gatewayv1.PortNumber](80),
					StatusCode: helpers.GetPointer(301),
				},
			},
			expectErrCount: 0,
			name:           "valid redirect filter",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type:            gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{},
			},
			expectErrCount: 0,
			name:           "valid redirect filter with no fields set",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateRedirectSchemeReturns(false, []string{"valid-scheme"})
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Scheme: helpers.GetPointer("http"), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid scheme",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateHostnameReturns(errors.New("invalid hostname"))
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Hostname: helpers.GetPointer[gatewayv1.PreciseHostname](
						"example.com",
					), // any value is invalid by the validator
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
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Port: helpers.GetPointer[gatewayv1.PortNumber](80), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid port",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Path: &gatewayv1.HTTPPathModifier{},
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
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					StatusCode: helpers.GetPointer(301), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "redirect filter with invalid status code",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := createAllValidValidator()
				validator.ValidateHostnameReturns(errors.New("invalid hostname"))
				validator.ValidateRedirectPortReturns(errors.New("invalid port"))
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &gatewayv1.HTTPRequestRedirectFilter{
					Hostname: helpers.GetPointer[gatewayv1.PreciseHostname](
						"example.com",
					), // any value is invalid by the validator
					Port: helpers.GetPointer[gatewayv1.PortNumber](
						80,
					), // any value is invalid by the validator
				},
			},
			expectErrCount: 2,
			name:           "redirect filter with multiple errors",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			allErrs := validateFilterRedirect(test.validator, test.filter, filterPath)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestValidateFilterRewrite(t *testing.T) {
	tests := []struct {
		filter         gatewayv1.HTTPRouteFilter
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			filter: gatewayv1.HTTPRouteFilter{
				Type:       gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: nil,
			},
			name:           "nil filter",
			expectErrCount: 1,
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Hostname: helpers.GetPointer[gatewayv1.PreciseHostname]("example.com"),
					Path: &gatewayv1.HTTPPathModifier{
						Type:            gatewayv1.FullPathHTTPPathModifier,
						ReplaceFullPath: helpers.GetPointer("/path"),
					},
				},
			},
			expectErrCount: 0,
			name:           "valid rewrite filter",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			filter: gatewayv1.HTTPRouteFilter{
				Type:       gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{},
			},
			expectErrCount: 0,
			name:           "valid rewrite filter with no fields set",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := &validationfakes.FakeHTTPFieldsValidator{}
				validator.ValidateHostnameReturns(errors.New("invalid hostname"))
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Hostname: helpers.GetPointer[gatewayv1.PreciseHostname](
						"example.com",
					), // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "rewrite filter with invalid hostname",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Path: &gatewayv1.HTTPPathModifier{
						Type: "bad-type",
					},
				},
			},
			expectErrCount: 1,
			name:           "rewrite filter with invalid path type",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := &validationfakes.FakeHTTPFieldsValidator{}
				validator.ValidateRewritePathReturns(errors.New("invalid path value"))
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Path: &gatewayv1.HTTPPathModifier{
						Type:            gatewayv1.FullPathHTTPPathModifier,
						ReplaceFullPath: helpers.GetPointer("/path"),
					}, // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "rewrite filter with invalid full path",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := &validationfakes.FakeHTTPFieldsValidator{}
				validator.ValidateRewritePathReturns(errors.New("invalid path"))
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Path: &gatewayv1.HTTPPathModifier{
						Type:               gatewayv1.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: helpers.GetPointer("/path"),
					}, // any value is invalid by the validator
				},
			},
			expectErrCount: 1,
			name:           "rewrite filter with invalid prefix path",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				validator := &validationfakes.FakeHTTPFieldsValidator{}
				validator.ValidateHostnameReturns(errors.New("invalid hostname"))
				validator.ValidateRewritePathReturns(errors.New("invalid path"))
				return validator
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterURLRewrite,
				URLRewrite: &gatewayv1.HTTPURLRewriteFilter{
					Hostname: helpers.GetPointer[gatewayv1.PreciseHostname](
						"example.com",
					), // any value is invalid by the validator
					Path: &gatewayv1.HTTPPathModifier{
						Type:               gatewayv1.PrefixMatchHTTPPathModifier,
						ReplacePrefixMatch: helpers.GetPointer("/path"),
					}, // any value is invalid by the validator
				},
			},
			expectErrCount: 2,
			name:           "rewrite filter with multiple errors",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			allErrs := validateFilterRewrite(test.validator, test.filter, filterPath)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestValidateFilterRequestHeaderModifier(t *testing.T) {
	createAllValidValidator := func() *validationfakes.FakeHTTPFieldsValidator {
		v := &validationfakes.FakeHTTPFieldsValidator{}
		return v
	}

	tests := []struct {
		filter         gatewayv1.HTTPRouteFilter
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set: []gatewayv1.HTTPHeader{
						{Name: "MyBespokeHeader", Value: "my-value"},
					},
					Add: []gatewayv1.HTTPHeader{
						{Name: "Accept-Encoding", Value: "gzip"},
					},
					Remove: []string{"Cache-Control"},
				},
			},
			expectErrCount: 0,
			name:           "valid request header modifier filter",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type:                  gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: nil,
			},
			expectErrCount: 1,
			name:           "nil request header modifier filter",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateRequestHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{Name: "$var_name", Value: "gzip"},
					},
				},
			},
			expectErrCount: 1,
			name:           "request header modifier filter with invalid add",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateRequestHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Remove: []string{"$var-name"},
				},
			},
			expectErrCount: 1,
			name:           "request header modifier filter with invalid remove",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateRequestHeaderValueReturns(errors.New("Invalid header value"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{Name: "Accept-Encoding", Value: "yhu$"},
					},
				},
			},
			expectErrCount: 1,
			name:           "request header modifier filter with invalid header value",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateRequestHeaderValueReturns(errors.New("Invalid header value"))
				v.ValidateRequestHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set: []gatewayv1.HTTPHeader{
						{Name: "Host", Value: "my_host"},
					},
					Add: []gatewayv1.HTTPHeader{
						{Name: "}90yh&$", Value: "gzip$"},
						{Name: "}67yh&$", Value: "compress$"},
					},
					Remove: []string{"Cache-Control$}"},
				},
			},
			expectErrCount: 7,
			name:           "request header modifier filter all fields invalid",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set: []gatewayv1.HTTPHeader{
						{Name: "MyBespokeHeader", Value: "my-value"},
						{Name: "mYbespokeHEader", Value: "duplicate"},
					},
					Add: []gatewayv1.HTTPHeader{
						{Name: "Accept-Encoding", Value: "gzip"},
						{Name: "accept-encodING", Value: "gzip"},
					},
					Remove: []string{"Cache-Control", "cache-control"},
				},
			},
			expectErrCount: 3,
			name:           "request header modifier filter not unique names",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			allErrs := validateFilterHeaderModifier(test.validator, test.filter, filterPath)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}
