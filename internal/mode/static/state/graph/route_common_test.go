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
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

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
			HTTPRoutes: map[types.NamespacedName]*HTTPRoute{},
			GRPCRoutes: map[types.NamespacedName]*GRPCRoute{},
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

	var normalRoute *HTTPRoute
	createNormalRoute := func(gateway *gatewayv1.Gateway) *HTTPRoute {
		normalRoute = &HTTPRoute{
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
	getLastNormalRoute := func() *HTTPRoute {
		return normalRoute
	}

	invalidAttachableRoute1 := &HTTPRoute{
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
	invalidAttachableRoute2 := &HTTPRoute{
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

	routeWithMissingSectionName := &HTTPRoute{
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
	routeWithEmptySectionName := &HTTPRoute{
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
	routeWithNonExistingListener := &HTTPRoute{
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
	routeWithPort := &HTTPRoute{
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
	routeWithIgnoredGateway := &HTTPRoute{
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
	invalidRoute := &HTTPRoute{
		Valid:  false,
		Source: hr,
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

	gr := createGRPCRoute("gr", "listener-80-1", gw.Name, "foo.example.com", []v1alpha2.GRPCRouteRule{})

	var normalGRPCRoute *GRPCRoute
	createNormalGRPCRoute := func(gateway *gatewayv1.Gateway) *GRPCRoute {
		normalGRPCRoute = &GRPCRoute{
			Source:     gr,
			Valid:      true,
			Attachable: true,
			ParentRefs: []ParentRef{
				{
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gateway),
				},
			},
		}
		return normalGRPCRoute
	}

	getLastNormalGRPCRoute := func() *GRPCRoute {
		return normalGRPCRoute
	}

	tests := []struct {
		route                    Route
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
						client.ObjectKeyFromObject(hr): getLastNormalRoute(),
					}
				}),
			},
			name: "normal case",
		},
		{
			route: createNormalGRPCRoute(gw),
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
					l.GRPCRoutes = map[types.NamespacedName]*GRPCRoute{
						client.ObjectKeyFromObject(gr): getLastNormalGRPCRoute(),
					}
				}),
			},
			name: "normal case with GRPCRoute",
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
						client.ObjectKeyFromObject(hr): routeWithEmptySectionName,
					}
				}),
				createModifiedListener("listener-8080", func(l *Listener) {
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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
					l.HTTPRoutes = map[types.NamespacedName]*HTTPRoute{
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

			var attachable bool
			var ns string
			var hostnames []gatewayv1.Hostname
			var parentRefs []ParentRef

			switch v := test.route.(type) {
			case *HTTPRoute:
				attachable = v.Attachable
				ns = v.Source.Namespace
				hostnames = v.Source.Spec.Hostnames
				parentRefs = v.ParentRefs
			case *GRPCRoute:
				attachable = v.Attachable
				ns = v.Source.Namespace
				hostnames = v.Source.Spec.Hostnames
				parentRefs = v.ParentRefs
			}

			bindRouteToListeners(
				test.route,
				attachable,
				ns,
				hostnames,
				test.gateway,
				namespaces,
			)

			g.Expect(parentRefs).To(Equal(test.expectedSectionNameRefs))
			g.Expect(helpers.Diff(test.gateway.Listeners, test.expectedGatewayListeners)).To(BeEmpty())
			switch v := test.route.(type) {
			case *HTTPRoute:
				g.Expect(helpers.Diff(v.Conditions, test.expectedConditions)).To(BeEmpty())
			case *GRPCRoute:
				g.Expect(helpers.Diff(v.Conditions, test.expectedConditions)).To(BeEmpty())
			}
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
