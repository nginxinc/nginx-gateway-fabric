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
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation/validationfakes"
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
			Idx:         0,
			Gateway:     gwNsName1,
			SectionName: parentRefs[0].SectionName,
		},
		{
			Idx:         2,
			Gateway:     gwNsName2,
			SectionName: parentRefs[2].SectionName,
		},
		{
			Idx:         3,
			Gateway:     gwNsName1,
			SectionName: parentRefs[3].SectionName,
		},
		{
			Idx:         4,
			Gateway:     gwNsName2,
			SectionName: parentRefs[4].SectionName,
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
				Kind:      helpers.GetPointer[gatewayv1.Kind](kinds.Gateway),
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
				Protocol: gatewayv1.HTTPProtocolType,
			},
			Valid:      true,
			Attachable: true,
			Routes:     map[RouteKey]*L7Route{},
			SupportedKinds: []gatewayv1.RouteGroupKind{
				{Kind: gatewayv1.Kind(kinds.HTTPRoute), Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName)},
				{Kind: gatewayv1.Kind(kinds.GRPCRoute), Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName)},
			},
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
			TypeMeta: metav1.TypeMeta{
				Kind: "HTTPRoute",
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

	var normalHTTPRoute *L7Route
	createNormalHTTPRoute := func(gateway *gatewayv1.Gateway) *L7Route {
		normalHTTPRoute = &L7Route{
			RouteType: RouteTypeHTTP,
			Source:    hr,
			Spec: L7RouteSpec{
				Hostnames: hr.Spec.Hostnames,
			},
			Valid:      true,
			Attachable: true,
			ParentRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gateway),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
				},
			},
		}
		return normalHTTPRoute
	}

	getLastNormalHTTPRoute := func() *L7Route {
		return normalHTTPRoute
	}

	invalidAttachableRoute1 := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hr,
		Valid:      false,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hr.Spec.ParentRefs[0].SectionName,
			},
		},
	}
	invalidAttachableRoute2 := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hr,
		Valid:      false,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hr.Spec.ParentRefs[0].SectionName,
			},
		},
	}

	routeWithMissingSectionName := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hrWithNilSectionName,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hrWithNilSectionName.Spec.ParentRefs[0].SectionName,
			},
		},
	}
	routeWithEmptySectionName := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hrWithEmptySectionName,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hrWithEmptySectionName.Spec.ParentRefs[0].SectionName,
			},
		},
	}
	routeWithNonExistingListener := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hrWithNonExistingListener,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hrWithNonExistingListener.Spec.ParentRefs[0].SectionName,
			},
		},
	}
	routeWithPort := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hrWithPort,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hrWithPort.Spec.ParentRefs[0].SectionName,
				Port:        hrWithPort.Spec.ParentRefs[0].Port,
			},
		},
	}
	ignoredGwNsName := types.NamespacedName{Namespace: "test", Name: "ignored-gateway"}
	routeWithIgnoredGateway := &L7Route{
		RouteType:  RouteTypeHTTP,
		Source:     hr,
		Valid:      true,
		Attachable: true,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     ignoredGwNsName,
				SectionName: hr.Spec.ParentRefs[0].SectionName,
			},
		},
	}
	invalidRoute := &L7Route{
		RouteType: RouteTypeHTTP,
		Valid:     false,
		Source:    hr,
		ParentRefs: []ParentRef{
			{
				Idx:         0,
				Gateway:     client.ObjectKeyFromObject(gw),
				SectionName: hr.Spec.ParentRefs[0].SectionName,
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

	createGRPCRouteWithSectionNameAndPort := func(
		sectionName *gatewayv1.SectionName,
		port *gatewayv1.PortNumber,
	) *gatewayv1.GRPCRoute {
		return &gatewayv1.GRPCRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "hr",
			},
			TypeMeta: metav1.TypeMeta{
				Kind: "GRPCRoute",
			},
			Spec: gatewayv1.GRPCRouteSpec{
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

	gr := createGRPCRouteWithSectionNameAndPort(helpers.GetPointer[gatewayv1.SectionName]("listener-80-1"), nil)

	var normalGRPCRoute *L7Route
	createNormalGRPCRoute := func(gateway *gatewayv1.Gateway) *L7Route {
		normalGRPCRoute = &L7Route{
			RouteType: RouteTypeGRPC,
			Source:    gr,
			Spec: L7RouteSpec{
				Hostnames: gr.Spec.Hostnames,
			},
			Valid:      true,
			Attachable: true,
			ParentRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gateway),
					SectionName: gr.Spec.ParentRefs[0].SectionName,
				},
			},
		}
		return normalGRPCRoute
	}

	getLastNormalGRPCRoute := func() *L7Route {
		return normalGRPCRoute
	}

	tests := []struct {
		route                    *L7Route
		gateway                  *Gateway
		expectedGatewayListeners []*Listener
		name                     string
		expectedSectionNameRefs  []ParentRef
		expectedConditions       []conditions.Condition
	}{
		{
			route: createNormalHTTPRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): getLastNormalHTTPRoute(),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hrWithNilSectionName.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): routeWithMissingSectionName,
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hrWithEmptySectionName.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): routeWithEmptySectionName,
					}
				}),
				createModifiedListener("listener-8080", func(l *Listener) {
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): routeWithEmptySectionName,
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hrWithEmptySectionName.Spec.ParentRefs[0].SectionName,
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hrWithPort.Spec.ParentRefs[0].SectionName,
					Attachment: &ParentRefAttachmentStatus{
						Attached: false,
						FailedCondition: staticConds.NewRouteUnsupportedValue(
							`spec.parentRefs[0].port: Forbidden: cannot be set`,
						),
						AcceptedHostnames: map[string][]string{},
					},
					Port: hrWithPort.Spec.ParentRefs[0].Port,
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hrWithNonExistingListener.Spec.ParentRefs[0].SectionName,
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
			route: createNormalHTTPRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					invalidNotAttachableListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
			route: createNormalHTTPRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					nonMatchingHostnameListener,
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					Idx:         0,
					Gateway:     ignoredGwNsName,
					SectionName: hr.Spec.ParentRefs[0].SectionName,
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteNotAcceptedGatewayIgnored(),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					Attachment:  nil,
					SectionName: hr.Spec.ParentRefs[0].SectionName,
				},
			},
			expectedGatewayListeners: []*Listener{
				createListener("listener-80-1"),
			},
			name: "route isn't valid",
		},
		{
			route: createNormalHTTPRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  false,
				Listeners: []*Listener{
					createListener("listener-80-1"),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
			route: createNormalHTTPRoute(gw),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): getLastNormalHTTPRoute(),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): invalidAttachableRoute1,
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): invalidAttachableRoute2,
					}
				}),
			},
			expectedConditions: []conditions.Condition{staticConds.NewRouteInvalidListener()},
			name:               "invalid attachable listener with invalid attachable route",
		},
		{
			route: createNormalHTTPRoute(gw),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
			route: createNormalHTTPRoute(gw),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): getLastNormalHTTPRoute(),
					}
				}),
			},
			name: "route allowed via labels",
		},
		{
			route: createNormalHTTPRoute(gwDiffNamespace),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gwDiffNamespace),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
			route: createNormalHTTPRoute(gw),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): getLastNormalHTTPRoute(),
					}
				}),
			},
			name: "route allowed via same namespace",
		},
		{
			route: createNormalHTTPRoute(gwDiffNamespace),
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
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gwDiffNamespace),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): getLastNormalHTTPRoute(),
					}
				}),
			},
			name: "route allowed via all namespaces",
		},
		{
			route: createNormalGRPCRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.SupportedKinds = []gatewayv1.RouteGroupKind{
							{Kind: gatewayv1.Kind(kinds.HTTPRoute), Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName)},
						}
						l.Routes = map[RouteKey]*L7Route{
							CreateRouteKey(gr): getLastNormalGRPCRoute(),
						}
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: gr.Spec.ParentRefs[0].SectionName,
					Attachment: &ParentRefAttachmentStatus{
						Attached:          false,
						FailedCondition:   staticConds.NewRouteNotAllowedByListeners(),
						AcceptedHostnames: map[string][]string{},
					},
				},
			},
			expectedGatewayListeners: []*Listener{
				createModifiedListener("listener-80-1", func(l *Listener) {
					l.SupportedKinds = []gatewayv1.RouteGroupKind{
						{Kind: gatewayv1.Kind(kinds.HTTPRoute), Group: helpers.GetPointer[gatewayv1.Group](gatewayv1.GroupName)},
					}
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(gr): getLastNormalGRPCRoute(),
					}
				}),
			},
			name: "grpc route not allowed when listener kind is HTTPRoute",
		},
		{
			route: createNormalHTTPRoute(gw),
			gateway: &Gateway{
				Source: gw,
				Valid:  true,
				Listeners: []*Listener{
					createModifiedListener("listener-80-1", func(l *Listener) {
						l.Source.AllowedRoutes = &gatewayv1.AllowedRoutes{
							Kinds: []gatewayv1.RouteGroupKind{
								{Kind: "HTTPRoute"},
							},
						}
						l.Routes = map[RouteKey]*L7Route{
							CreateRouteKey(hr): getLastNormalHTTPRoute(),
						}
					}),
				},
			},
			expectedSectionNameRefs: []ParentRef{
				{
					Idx:         0,
					Gateway:     client.ObjectKeyFromObject(gw),
					SectionName: hr.Spec.ParentRefs[0].SectionName,
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
						Kinds: []gatewayv1.RouteGroupKind{
							{Kind: "HTTPRoute"},
						},
					}
					l.Routes = map[RouteKey]*L7Route{
						CreateRouteKey(hr): getLastNormalHTTPRoute(),
					}
				}),
			},
			name: "http route allowed when listener kind is HTTPRoute",
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

			bindRouteToListeners(
				test.route,
				test.gateway,
				namespaces,
			)

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
				v.ValidateFilterHeaderNameReturns(errors.New("Invalid header"))
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
				v.ValidateFilterHeaderNameReturns(errors.New("Invalid header"))
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
				v.ValidateFilterHeaderValueReturns(errors.New("Invalid header value"))
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
				v.ValidateFilterHeaderValueReturns(errors.New("Invalid header value"))
				v.ValidateFilterHeaderNameReturns(errors.New("Invalid header"))
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
			allErrs := validateFilterHeaderModifier(
				test.validator, test.filter.RequestHeaderModifier, filterPath,
			)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestValidateFilterResponseHeaderModifier(t *testing.T) {
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
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
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
			name:           "valid response header modifier filter",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type:                   gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: nil,
			},
			expectErrCount: 1,
			name:           "nil response header modifier filter",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateFilterHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{Name: "$var_name", Value: "gzip"},
					},
				},
			},
			expectErrCount: 1,
			name:           "response header modifier filter with invalid add",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateFilterHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Remove: []string{"$var-name"},
				},
			},
			expectErrCount: 1,
			name:           "response header modifier filter with invalid remove",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateFilterHeaderValueReturns(errors.New("Invalid header value"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Add: []gatewayv1.HTTPHeader{
						{Name: "Accept-Encoding", Value: "yhu$"},
					},
				},
			},
			expectErrCount: 1,
			name:           "response header modifier filter with invalid header value",
		},
		{
			validator: func() *validationfakes.FakeHTTPFieldsValidator {
				v := createAllValidValidator()
				v.ValidateFilterHeaderValueReturns(errors.New("Invalid header value"))
				v.ValidateFilterHeaderNameReturns(errors.New("Invalid header"))
				return v
			}(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
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
			name:           "response header modifier filter all fields invalid",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
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
			name:           "response header modifier filter not unique names",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set: []gatewayv1.HTTPHeader{
						{Name: "Content-Length", Value: "163"},
					},
					Add: []gatewayv1.HTTPHeader{
						{Name: "Content-Type", Value: "text/plain"},
					},
					Remove: []string{"X-Pad"},
				},
			},
			expectErrCount: 3,
			name:           "response header modifier filter with disallowed header name",
		},
		{
			validator: createAllValidValidator(),
			filter: gatewayv1.HTTPRouteFilter{
				Type: gatewayv1.HTTPRouteFilterResponseHeaderModifier,
				ResponseHeaderModifier: &gatewayv1.HTTPHeaderFilter{
					Set: []gatewayv1.HTTPHeader{
						{Name: "X-Accel-Redirect", Value: "/protected/iso.img"},
					},
					Add: []gatewayv1.HTTPHeader{
						{Name: "X-Accel-Limit-Rate", Value: "1024"},
					},
					Remove: []string{"X-Accel-Charset"},
				},
			},
			expectErrCount: 3,
			name:           "response header modifier filter with disallowed header name prefix",
		},
	}

	filterPath := field.NewPath("test")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			allErrs := validateFilterResponseHeaderModifier(
				test.validator, test.filter.ResponseHeaderModifier, filterPath,
			)
			g.Expect(allErrs).To(HaveLen(test.expectErrCount))
		})
	}
}

func TestRouteKeyForKind(t *testing.T) {
	nsname := types.NamespacedName{Namespace: testNs, Name: "route"}

	g := NewWithT(t)

	key := routeKeyForKind(kinds.HTTPRoute, nsname)
	g.Expect(key).To(Equal(RouteKey{RouteType: RouteTypeHTTP, NamespacedName: nsname}))

	key = routeKeyForKind(kinds.GRPCRoute, nsname)
	g.Expect(key).To(Equal(RouteKey{RouteType: RouteTypeGRPC, NamespacedName: nsname}))

	rk := func() {
		_ = routeKeyForKind(kinds.Gateway, nsname)
	}

	g.Expect(rk).To(Panic())
}

func TestAllowedRouteType(t *testing.T) {
	test := []struct {
		listener  *Listener
		name      string
		routeType RouteType
		expResult bool
	}{
		{
			name:      "grpcRoute is allowed when listener supports grpcRoute kind",
			routeType: RouteTypeGRPC,
			listener: &Listener{
				SupportedKinds: []gatewayv1.RouteGroupKind{
					{Kind: kinds.GRPCRoute},
				},
			},
			expResult: true,
		},
		{
			name:      "grpcRoute is allowed when listener supports grpcRoute and httpRoute kind",
			routeType: RouteTypeGRPC,
			listener: &Listener{
				SupportedKinds: []gatewayv1.RouteGroupKind{
					{Kind: kinds.HTTPRoute},
					{Kind: kinds.GRPCRoute},
				},
			},
			expResult: true,
		},
		{
			name:      "grpcRoute is allowed when listener supports httpRoute kind",
			routeType: RouteTypeGRPC,
			listener: &Listener{
				SupportedKinds: []gatewayv1.RouteGroupKind{
					{Kind: kinds.HTTPRoute},
				},
			},
			expResult: false,
		},
		{
			name:      "httpRoute not allowed when listener supports grpcRoute kind",
			routeType: RouteTypeHTTP,
			listener: &Listener{
				SupportedKinds: []gatewayv1.RouteGroupKind{
					{Kind: kinds.GRPCRoute},
				},
			},
			expResult: false,
		},
	}

	for _, test := range test {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(isRouteTypeAllowedByListener(test.listener, test.routeType)).To(Equal(test.expResult))
		})
	}
}
