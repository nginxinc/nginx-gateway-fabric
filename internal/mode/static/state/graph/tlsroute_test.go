package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

func createTLSRoute(
	hostname gatewayv1.Hostname,
	rules []v1alpha2.TLSRouteRule,
	parentRefs []gatewayv1.ParentReference,
) *v1alpha2.TLSRoute {
	return &v1alpha2.TLSRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "tr",
		},
		Spec: v1alpha2.TLSRouteSpec{
			CommonRouteSpec: gatewayv1.CommonRouteSpec{
				ParentRefs: parentRefs,
			},
			Hostnames: []gatewayv1.Hostname{hostname},
			Rules:     rules,
		},
	}
}

func TestBuildTLSRoute(t *testing.T) {
	parentRef := gatewayv1.ParentReference{
		Namespace:   helpers.GetPointer[gatewayv1.Namespace]("test"),
		Name:        "gateway",
		SectionName: helpers.GetPointer[gatewayv1.SectionName]("l1"),
	}
	gatewayNsName := types.NamespacedName{
		Namespace: "test",
		Name:      "gateway",
	}
	parentRefGraph := ParentRef{
		SectionName: helpers.GetPointer[gatewayv1.SectionName]("l1"),
		Gateway:     gatewayNsName,
	}
	duplicateParentRefsGtr := createTLSRoute(
		"hi.example.com",
		nil,
		[]gatewayv1.ParentReference{
			parentRef,
			parentRef,
		},
	)
	noParentRefsGtr := createTLSRoute(
		"hi.example.com",
		nil,
		[]gatewayv1.ParentReference{},
	)
	invalidHostnameGtr := createTLSRoute(
		"hi....com",
		nil,
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)
	noRulesGtr := createTLSRoute(
		"app.example.com",
		nil,
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)
	backedRefDNEGtr := createTLSRoute(
		"app.example.com",
		[]v1alpha2.TLSRouteRule{
			{
				BackendRefs: []gatewayv1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: "hi",
							Port: helpers.GetPointer[gatewayv1.PortNumber](80),
						},
					},
				},
			},
		},
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)

	wrongBackendRefGroupGtr := createTLSRoute(
		"app.example.com",
		[]v1alpha2.TLSRouteRule{
			{
				BackendRefs: []gatewayv1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name:  "hi",
							Port:  helpers.GetPointer[gatewayv1.PortNumber](80),
							Group: helpers.GetPointer[gatewayv1.Group]("wrong"),
						},
					},
				},
			},
		},
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)

	wrongBackendRefKindGtr := createTLSRoute(
		"app.example.com",
		[]v1alpha2.TLSRouteRule{
			{
				BackendRefs: []gatewayv1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: "hi",
							Port: helpers.GetPointer[gatewayv1.PortNumber](80),
							Kind: helpers.GetPointer[gatewayv1.Kind]("not service"),
						},
					},
				},
			},
		},
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)

	wrongBackendRefNamespaceGtr8 := createTLSRoute("app.example.com",
		[]v1alpha2.TLSRouteRule{
			{
				BackendRefs: []gatewayv1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name:      "hi",
							Port:      helpers.GetPointer[gatewayv1.PortNumber](80),
							Namespace: helpers.GetPointer[gatewayv1.Namespace]("wrong"),
						},
					},
				},
			},
		},
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)

	portNilBackendRefGtr := createTLSRoute("app.example.com",
		[]v1alpha2.TLSRouteRule{
			{
				BackendRefs: []gatewayv1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: "hi",
						},
					},
				},
			},
		},
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)

	ipFamilyMismatchGtr := createTLSRoute(
		"app.example.com",
		[]v1alpha2.TLSRouteRule{
			{
				BackendRefs: []gatewayv1.BackendRef{
					{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: "hi",
							Port: helpers.GetPointer[gatewayv1.PortNumber](80),
						},
					},
				},
			},
		},
		[]gatewayv1.ParentReference{
			parentRef,
		},
	)
	svcNsName := types.NamespacedName{
		Namespace: "test",
		Name:      "hi",
	}

	svcNsNameWrong := types.NamespacedName{
		Namespace: "wrong",
		Name:      "hi",
	}

	createSvc := func(name string, port int32) *apiv1.Service {
		return &apiv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: apiv1.ServiceSpec{
				Ports: []apiv1.ServicePort{
					{Port: port},
				},
			},
		}
	}

	badNsSvc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "wrong",
			Name:      "hi",
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{Port: 80},
			},
		},
	}

	ipv4Svc := &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "hi",
		},
		Spec: apiv1.ServiceSpec{
			IPFamilies: []apiv1.IPFamily{
				apiv1.IPv4Protocol,
			},
			Ports: []apiv1.ServicePort{
				{Port: 80},
			},
		},
	}

	tests := []struct {
		expected       *L4Route
		gtr            *v1alpha2.TLSRoute
		services       map[types.NamespacedName]*apiv1.Service
		name           string
		gatewayNsNames []types.NamespacedName
		npCfg          NginxProxy
	}{
		{
			gtr:            duplicateParentRefsGtr,
			expected:       &L4Route{Source: duplicateParentRefsGtr},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services:       map[types.NamespacedName]*apiv1.Service{},
			name:           "duplicate parent refs",
		},
		{
			gtr:            noParentRefsGtr,
			expected:       nil,
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services:       map[types.NamespacedName]*apiv1.Service{},
			name:           "no parent refs",
		},
		{
			gtr: invalidHostnameGtr,
			expected: &L4Route{
				Source:     invalidHostnameGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Conditions: []conditions.Condition{staticConds.NewRouteUnsupportedValue(
					"spec.hostnames[0]: Invalid value: \"hi....com\": a lowercase RFC 1" +
						"123 subdomain must consist of lower case alphanumeric characters" +
						", '-' or '.', and must start and end with an alphanumeric charac" +
						"ter (e.g. 'example.com', regex used for validation is '[a-z0-9](" +
						"[-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*')",
				)},
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services:       map[types.NamespacedName]*apiv1.Service{},
			name:           "invalid hostname",
		},
		{
			gtr: noRulesGtr,
			expected: &L4Route{
				Source:     noRulesGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteBackendRefUnsupportedValue(
					"Must have exactly one Rule and BackendRef",
				)},
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services:       map[types.NamespacedName]*apiv1.Service{},
			name:           "invalid rule",
		},
		{
			gtr: backedRefDNEGtr,
			expected: &L4Route{
				Source:     backedRefDNEGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
					BackendRef: BackendRef{
						SvcNsName: types.NamespacedName{
							Namespace: "test",
							Name:      "hi",
						},
					},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteBackendRefRefBackendNotFound(
					"spec.rules[0].backendRefs[0].name: Not found: \"hi\"",
				)},
				Attachable: true,
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services:       map[types.NamespacedName]*apiv1.Service{},
			name:           "BackendRef not found",
		},
		{
			gtr: wrongBackendRefGroupGtr,
			expected: &L4Route{
				Source:     wrongBackendRefGroupGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
					BackendRef: BackendRef{
						SvcNsName:   svcNsName,
						ServicePort: apiv1.ServicePort{Port: 80},
					},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteBackendRefInvalidKind(
					"spec.rules[0].backendRefs[0].group:" +
						" Unsupported value: \"wrong\": supported values: \"core\", \"\"",
				)},
				Attachable: true,
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services: map[types.NamespacedName]*apiv1.Service{
				svcNsName: createSvc("hi", 80),
			},
			name: "BackendRef group wrong",
		},
		{
			gtr: wrongBackendRefKindGtr,
			expected: &L4Route{
				Source:     wrongBackendRefKindGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
					BackendRef: BackendRef{
						SvcNsName:   svcNsName,
						ServicePort: apiv1.ServicePort{Port: 80},
					},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteBackendRefInvalidKind(
					"spec.rules[0].backendRefs[0].kind:" +
						" Unsupported value: \"not service\": supported values: \"Service\"",
				)},
				Attachable: true,
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services: map[types.NamespacedName]*apiv1.Service{
				svcNsName: createSvc("hi", 80),
			},
			name: "BackendRef kind wrong",
		},
		{
			gtr: wrongBackendRefNamespaceGtr8,
			expected: &L4Route{
				Source:     wrongBackendRefNamespaceGtr8,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
					BackendRef: BackendRef{
						SvcNsName:   svcNsNameWrong,
						ServicePort: apiv1.ServicePort{Port: 80},
					},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteBackendRefUnsupportedValue(
					"Cross-namespace routing is not supported",
				)},
				Attachable: true,
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services: map[types.NamespacedName]*apiv1.Service{
				svcNsNameWrong: badNsSvc,
			},
			name: "BackendRef namespace wrong",
		},
		{
			gtr: portNilBackendRefGtr,
			expected: &L4Route{
				Source:     portNilBackendRefGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
					BackendRef: BackendRef{},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteBackendRefUnsupportedValue(
					"spec.rules[0].backendRefs[0].port: Required value: port cannot be nil",
				)},
				Attachable: true,
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services: map[types.NamespacedName]*apiv1.Service{
				svcNsNameWrong: createSvc("hi", 80),
			},
			name: "BackendRef port nil",
		},
		{
			gtr: ipFamilyMismatchGtr,
			expected: &L4Route{
				Source:     ipFamilyMismatchGtr,
				ParentRefs: []ParentRef{parentRefGraph},
				Spec: L4RouteSpec{
					Hostnames: []gatewayv1.Hostname{
						"app.example.com",
					},
					BackendRef: BackendRef{
						SvcNsName:   svcNsName,
						ServicePort: apiv1.ServicePort{Port: 80},
					},
				},
				Conditions: []conditions.Condition{staticConds.NewRouteInvalidIPFamily(
					"Service configured with IPv4 family but NginxProxy is configured with IPv6",
				)},
				Attachable: true,
			},
			gatewayNsNames: []types.NamespacedName{gatewayNsName},
			services: map[types.NamespacedName]*apiv1.Service{
				svcNsName: ipv4Svc,
			},
			name: "service and npcfg ip family mismatch",
			npCfg: NginxProxy{
				Source: &ngfAPI.NginxProxy{Spec: ngfAPI.NginxProxySpec{IPFamily: helpers.GetPointer(ngfAPI.IPv6)}},
				Valid:  true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			r := buildTLSRoute(
				test.gtr,
				test.gatewayNsNames,
				test.services,
				&test.npCfg,
			)
			g.Expect(helpers.Diff(test.expected, r)).To(BeEmpty())
		})
	}
}
