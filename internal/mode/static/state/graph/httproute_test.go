package graph

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/validation/validationfakes"
)

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
					ParentRefs: []ParentRef{
						{
							Idx:     0,
							Gateway: gwNsName,
						},
					},
					Valid: true,
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
			g := NewGomegaWithT(t)
			routes := buildRoutesForGateways(validator, hrRoutes, test.gwNsNames)
			g.Expect(helpers.Diff(test.expected, routes)).To(BeEmpty())
		})
	}
}

func TestBuildSectionNameRefs(t *testing.T) {
	const routeNamespace = "test"

	gwNsName1 := types.NamespacedName{Namespace: routeNamespace, Name: "gateway-1"}
	gwNsName2 := types.NamespacedName{Namespace: routeNamespace, Name: "gateway-2"}

	parentRefs := []v1beta1.ParentReference{
		{
			Name:        v1beta1.ObjectName(gwNsName1.Name),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("one"),
		},
		{
			Name:        v1beta1.ObjectName("some-other-gateway"),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("two"),
		},
		{
			Name:        v1beta1.ObjectName(gwNsName2.Name),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("three"),
		},
		{
			Name:        v1beta1.ObjectName(gwNsName1.Name),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("same-name"),
		},
		{
			Name:        v1beta1.ObjectName(gwNsName2.Name),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("same-name"),
		},
		{
			Name:        v1beta1.ObjectName("some-other-gateway"),
			SectionName: helpers.GetPointer[v1beta1.SectionName]("same-name"),
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

	g := NewGomegaWithT(t)

	result := buildSectionNameRefs(parentRefs, routeNamespace, gwNsNames)
	g.Expect(result).To(Equal(expected))
}

func TestBuildSectionNameRefsPanicsForDuplicateParentRefs(t *testing.T) {
	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	tests := []struct {
		name       string
		parentRefs []v1beta1.ParentReference
	}{
		{
			parentRefs: []v1beta1.ParentReference{
				{
					Name:        v1beta1.ObjectName(gwNsName.Name),
					SectionName: helpers.GetPointer[v1beta1.SectionName]("http"),
				},
				{
					Name:        v1beta1.ObjectName(gwNsName.Name),
					SectionName: helpers.GetPointer[v1beta1.SectionName]("http"),
				},
			},
			name: "with sectionNames",
		},
		{
			parentRefs: []v1beta1.ParentReference{
				{
					Name:        v1beta1.ObjectName(gwNsName.Name),
					SectionName: nil,
				},
				{
					Name:        v1beta1.ObjectName(gwNsName.Name),
					SectionName: nil,
				},
			},
			name: "nil sectionNames",
		},
	}

	gwNsNames := []types.NamespacedName{gwNsName}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			run := func() { buildSectionNameRefs(test.parentRefs, gwNsName.Namespace, gwNsNames) }
			g.Expect(run).To(Panic())
		})
	}
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
				Namespace: helpers.GetPointer(v1beta1.Namespace(gwNsName1.Namespace)),
				Name:      v1beta1.ObjectName(gwNsName1.Name),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName1,
			name:             "found",
		},
		{
			ref: v1beta1.ParentReference{
				Group:     helpers.GetPointer[v1beta1.Group](v1beta1.GroupName),
				Kind:      helpers.GetPointer[v1beta1.Kind]("Gateway"),
				Namespace: helpers.GetPointer(v1beta1.Namespace(gwNsName1.Namespace)),
				Name:      v1beta1.ObjectName(gwNsName1.Name),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName1,
			name:             "found with explicit group and kind",
		},
		{
			ref: v1beta1.ParentReference{
				Name: v1beta1.ObjectName(gwNsName2.Name),
			},
			expectedFound:    true,
			expectedGwNsName: gwNsName2,
			name:             "found with implicit namespace",
		},
		{
			ref: v1beta1.ParentReference{
				Kind: helpers.GetPointer[v1beta1.Kind]("NotGateway"),
				Name: v1beta1.ObjectName(gwNsName2.Name),
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "wrong kind",
		},
		{
			ref: v1beta1.ParentReference{
				Group: helpers.GetPointer[v1beta1.Group]("wrong-group"),
				Name:  v1beta1.ObjectName(gwNsName2.Name),
			},
			expectedFound:    false,
			expectedGwNsName: types.NamespacedName{},
			name:             "wrong group",
		},
		{
			ref: v1beta1.ParentReference{
				Namespace: helpers.GetPointer(v1beta1.Namespace(gwNsName1.Namespace)),
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
			g := NewGomegaWithT(t)

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
		ValidatePathInMatchStub: func(path string) error {
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
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Valid: true,
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
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(`spec.hostnames[0]: Invalid value: "": cannot be empty string`),
				},
			},
			name: "invalid hostname",
		},
		{
			validator: validatorInvalidFieldsInRule,
			hr:        hrInvalidMatches,
			expected: &Route{
				Source: hrInvalidMatches,
				Valid:  false,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(
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
				Source: hrInvalidFilters,
				Valid:  false,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					conditions.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].filters[0].requestRedirect.hostname: ` +
							`Invalid value: "invalid.example.com": invalid hostname`),
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
			hr:        hrInvalidValidRules,
			expected: &Route{
				Source: hrInvalidValidRules,
				Valid:  true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					conditions.NewTODO(
						`Some rules are invalid: ` +
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
	createListener := func(name string) *Listener {
		return &Listener{
			Source: v1beta1.Listener{
				Name:     v1beta1.SectionName(name),
				Hostname: (*v1beta1.Hostname)(helpers.GetStringPointer("foo.example.com")),
			},
			Valid:  true,
			Routes: map[types.NamespacedName]*Route{},
		}
	}
	createModifiedListener := func(name string, m func(*Listener)) *Listener {
		l := createListener(name)
		m(l)
		return l
	}

	gw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway",
		},
	}
	gwDiffNamespace := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "diff-namespace",
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
	hrWithNilSectionName := createHTTPRouteWithSectionNameAndPort(nil, nil)
	hrWithEmptySectionName := createHTTPRouteWithSectionNameAndPort(helpers.GetPointer[v1beta1.SectionName](""), nil)
	hrWithPort := createHTTPRouteWithSectionNameAndPort(
		helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
		helpers.GetPointer[v1beta1.PortNumber](80),
	)
	hrWithNonExistingListener := createHTTPRouteWithSectionNameAndPort(
		helpers.GetPointer[v1beta1.SectionName]("listener-80-2"),
		nil,
	)

	var normalRoute *Route
	createNormalRoute := func(gateway *v1beta1.Gateway) *Route {
		normalRoute = &Route{
			Source: hr,
			Valid:  true,
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

	routeWithMissingSectionName := &Route{
		Source: hrWithNilSectionName,
		Valid:  true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	routeWithEmptySectionName := &Route{
		Source: hrWithEmptySectionName,
		Valid:  true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	routeWithNonExistingListener := &Route{
		Source: hrWithNonExistingListener,
		Valid:  true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	routeWithPort := &Route{
		Source: hrWithPort,
		Valid:  true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}
	ignoredGwNsName := types.NamespacedName{Namespace: "test", Name: "ignored-gateway"}
	routeWithIgnoredGateway := &Route{
		Source: hr,
		Valid:  true,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: ignoredGwNsName,
			},
		},
	}
	notValidRoute := &Route{
		Valid: false,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw),
			},
		},
	}

	notValidListener := createModifiedListener("", func(l *Listener) {
		l.Valid = false
	})
	nonMatchingHostnameListener := createModifiedListener("", func(l *Listener) {
		l.Source.Hostname = helpers.GetPointer[v1beta1.Hostname]("bar.example.com")
	})

	tests := []struct {
		route                    *Route
		gateway                  *Gateway
		expectedGatewayListeners map[string]*Listener
		name                     string
		expectedSectionNameRefs  []ParentRef
	}{
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
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
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
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
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
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
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
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
				Listeners: map[string]*Listener{
					"listener-80":   createListener("listener-80"),
					"listener-8080": createListener("listener-8080"),
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
			expectedGatewayListeners: map[string]*Listener{
				"listener-80": createModifiedListener("listener-80", func(l *Listener) {
					l.Routes = map[types.NamespacedName]*Route{
						client.ObjectKeyFromObject(hr): routeWithEmptySectionName,
					}
				}),
				"listener-8080": createModifiedListener("listener-8080", func(l *Listener) {
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
				Listeners: map[string]*Listener{
					"listener-80-1": notValidListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   conditions.NewRouteInvalidListener(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": notValidListener,
			},
			name: "empty section name with no valid listeners",
		},
		{
			route: routeWithPort,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached: false,
						FailedCondition: conditions.NewRouteUnsupportedValue(
							`spec.parentRefs[0].port: Forbidden: cannot be set`,
						),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener("listener-80-1"),
			},
			name: "port is configured",
		},
		{
			route: routeWithNonExistingListener,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   conditions.NewRouteNoMatchingParent(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener("listener-80-1"),
			},
			name: "listener doesn't exist",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": notValidListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   conditions.NewRouteInvalidListener(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": notValidListener,
			},
			name: "listener isn't valid",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": nonMatchingHostnameListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   conditions.NewRouteNoMatchingListenerHostname(),
						AcceptedHostnames: map[string][]string{},
					},
				},
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
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: ignoredGwNsName,
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   conditions.NewTODO("Gateway is ignored"),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener("listener-80-1"),
			},
			name: "gateway is ignored",
		},
		{
			route: notValidRoute,
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:        0,
					Gateway:    client.ObjectKeyFromObject(gw),
					Attachment: nil,
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener("listener-80-1"),
			},
			name: "route isn't valid",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  false,
				Listeners: map[string]*Listener{
					"listener-80-1": createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   conditions.NewRouteInvalidGateway(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createListener("listener-80-1"),
			},
			name: "invalid gateway",
		},
		{
			route: createNormalRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: map[string]*Listener{
					"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
							Namespaces: &v1beta1.RouteNamespaces{
								From: helpers.GetPointer(v1beta1.NamespacesFromSelector),
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
						FailedCondition:   conditions.NewRouteNotAllowedByListeners(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
						Namespaces: &v1beta1.RouteNamespaces{
							From: helpers.GetPointer(v1beta1.NamespacesFromSelector),
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
				Listeners: map[string]*Listener{
					"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
							Namespaces: &v1beta1.RouteNamespaces{
								From: helpers.GetPointer(v1beta1.NamespacesFromSelector),
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
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
					allowedLabels := map[string]string{"app": "allowed"}
					l.AllowedRouteLabelSelector = labels.SelectorFromSet(allowedLabels)
					l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
						Namespaces: &v1beta1.RouteNamespaces{
							From: helpers.GetPointer(v1beta1.NamespacesFromSelector),
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
				Listeners: map[string]*Listener{
					"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
							Namespaces: &v1beta1.RouteNamespaces{
								From: helpers.GetPointer(v1beta1.NamespacesFromSame),
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
						FailedCondition:   conditions.NewRouteNotAllowedByListeners(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
						Namespaces: &v1beta1.RouteNamespaces{
							From: helpers.GetPointer(v1beta1.NamespacesFromSame),
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
				Listeners: map[string]*Listener{
					"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
							Namespaces: &v1beta1.RouteNamespaces{
								From: helpers.GetPointer(v1beta1.NamespacesFromSame),
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
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
						Namespaces: &v1beta1.RouteNamespaces{
							From: helpers.GetPointer(v1beta1.NamespacesFromSame),
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
				Listeners: map[string]*Listener{
					"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
							Namespaces: &v1beta1.RouteNamespaces{
								From: helpers.GetPointer(v1beta1.NamespacesFromAll),
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
			expectedGatewayListeners: map[string]*Listener{
				"listener-80-1": createModifiedListener("listener-80-1", func(l *Listener) {
					l.Source.AllowedRoutes = &v1beta1.AllowedRoutes{
						Namespaces: &v1beta1.RouteNamespaces{
							From: helpers.GetPointer(v1beta1.NamespacesFromAll),
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
			g := NewGomegaWithT(t)

			bindRouteToListeners(test.route, test.gateway, namespaces)

			g.Expect(test.route.ParentRefs).To(Equal(test.expectedSectionNameRefs))
			g.Expect(helpers.Diff(test.gateway.Listeners, test.expectedGatewayListeners)).To(BeEmpty())
		})
	}
}

func TestFindAcceptedHostnames(t *testing.T) {
	var listenerHostnameFoo v1beta1.Hostname = "foo.example.com"
	var listenerHostnameCafe v1beta1.Hostname = "cafe.example.com"
	var listenerHostnameWildcard v1beta1.Hostname = "*.example.com"
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
			routeHostnames:   []v1beta1.Hostname{"*.example.com"},
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
			routeHostnames:   []v1beta1.Hostname{"*.example.com"},
			expected:         []string{"*.example.com"},
			msg:              "route wildcard hostname; nil listener hostname",
		},
		{
			listenerHostname: &listenerHostnameWildcard,
			routeHostnames:   []v1beta1.Hostname{"*.bar.example.com"},
			expected:         []string{"*.bar.example.com"},
			msg:              "route and listener wildcard hostnames",
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
		name      string
		hostnames []v1beta1.Hostname
		expectErr bool
	}{
		{
			hostnames: []v1beta1.Hostname{
				validHostname,
				"example.org",
				"foo.example.net",
			},
			expectErr: false,
			name:      "multiple valid",
		},
		{
			hostnames: []v1beta1.Hostname{
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
			g := NewGomegaWithT(t)

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
					Type:  helpers.GetPointer(v1beta1.PathMatchExact),
					Value: helpers.GetPointer("/"),
				},
			},
			expectErrCount: 0,
			name:           "valid exact match",
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
				validator.ValidatePathInMatchReturns(errors.New("invalid path value"))
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
	tests := []struct {
		filter         v1beta1.HTTPRouteFilter
		name           string
		expectErrCount int
	}{
		{
			filter: v1beta1.HTTPRouteFilter{
				Type:            v1beta1.HTTPRouteFilterRequestRedirect,
				RequestRedirect: &v1beta1.HTTPRequestRedirectFilter{},
			},
			expectErrCount: 0,
			name:           "valid redirect filter",
		},
		{
			filter: v1beta1.HTTPRouteFilter{
				Type:                  v1beta1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{},
			},
			expectErrCount: 0,
			name:           "valid request header modifiers filter",
		},
		{
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterURLRewrite,
			},
			expectErrCount: 1,
			name:           "unsupported filter",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
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
					Hostname: helpers.GetPointer[v1beta1.PreciseHostname](
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
					Hostname: helpers.GetPointer[v1beta1.PreciseHostname](
						"example.com",
					), // any value is invalid by the validator
					Port: helpers.GetPointer[v1beta1.PortNumber](
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
			g := NewGomegaWithT(t)
			allErrs := validateFilterRedirect(test.validator, test.filter, filterPath)
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
		filter         v1beta1.HTTPRouteFilter
		validator      *validationfakes.FakeHTTPFieldsValidator
		name           string
		expectErrCount int
	}{
		{
			validator: createAllValidValidator(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
					Set: []v1beta1.HTTPHeader{
						{Name: "Connection", Value: "close"},
					},
					Add: []v1beta1.HTTPHeader{
						{Name: "Accept-Encoding", Value: "gzip"},
					},
					Remove: []string{"Cache-Control"},
				},
			},
			expectErrCount: 0,
			name:           "valid request header modifier filter",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateRequestHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
					Add: []v1beta1.HTTPHeader{
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
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
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
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
					Add: []v1beta1.HTTPHeader{
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
			filter: v1beta1.HTTPRouteFilter{
				Type: v1beta1.HTTPRouteFilterRequestHeaderModifier,
				RequestHeaderModifier: &v1beta1.HTTPHeaderFilter{
					Set: []v1beta1.HTTPHeader{
						{Name: "Host", Value: "my_host"},
					},
					Add: []v1beta1.HTTPHeader{
						{Name: "}90yh&$", Value: "gzip$"},
						{Name: "}67yh&$", Value: "compress$"},
					},
					Remove: []string{"Cache-Control$}"},
				},
			},
			expectErrCount: 7,
			name:           "request header modifier filter all fields invalid",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)
			allErrs := validateFilterHeaderModifier(test.validator, test.filter, filterPath)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}
