//nolint:gosec
package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/controller/index"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation/validationfakes"
)

func TestBuildGraph(t *testing.T) {
	const (
		gcName         = "my-class"
		controllerName = "my.controller"
	)

	protectedPorts := ProtectedPorts{
		9113: "MetricsPort",
		8081: "HealthPort",
	}

	createValidRuleWithBackendRefs := func(refs []BackendRef) Rule {
		return Rule{
			ValidMatches: true,
			ValidFilters: true,
			BackendRefs:  refs,
		}
	}

	createRoute := func(name string, gatewayName string, listenerName string) *gatewayv1.HTTPRoute {
		return &gatewayv1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: gatewayv1.HTTPRouteSpec{
				CommonRouteSpec: gatewayv1.CommonRouteSpec{
					ParentRefs: []gatewayv1.ParentReference{
						{
							Namespace:   (*gatewayv1.Namespace)(helpers.GetPointer("test")),
							Name:        gatewayv1.ObjectName(gatewayName),
							SectionName: (*gatewayv1.SectionName)(helpers.GetPointer(listenerName)),
						},
					},
				},
				Hostnames: []gatewayv1.Hostname{
					"foo.example.com",
				},
				Rules: []gatewayv1.HTTPRouteRule{
					{
						Matches: []gatewayv1.HTTPRouteMatch{
							{
								Path: &gatewayv1.HTTPPathMatch{
									Type:  helpers.GetPointer(gatewayv1.PathMatchPathPrefix),
									Value: helpers.GetPointer("/"),
								},
							},
						},
						BackendRefs: []gatewayv1.HTTPBackendRef{
							{
								BackendRef: gatewayv1.BackendRef{
									BackendObjectReference: gatewayv1.BackendObjectReference{
										Kind:      (*gatewayv1.Kind)(helpers.GetPointer("Service")),
										Name:      "foo",
										Namespace: (*gatewayv1.Namespace)(helpers.GetPointer("service")),
										Port:      (*gatewayv1.PortNumber)(helpers.GetPointer[int32](80)),
									},
								},
							},
						},
					},
				},
			},
		}
	}

	hr1 := createRoute("hr-1", "gateway-1", "listener-80-1")
	hr2 := createRoute("hr-2", "wrong-gateway", "listener-80-1")
	hr3 := createRoute("hr-3", "gateway-1", "listener-443-1") // https listener; should not conflict with hr1

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "configmap",
			Namespace: "service",
		},
		Data: map[string]string{
			"ca.crt": caBlock,
		},
	}

	btpAcceptedConds := []conditions.Condition{
		staticConds.NewBackendTLSPolicyAccepted(),
		staticConds.NewBackendTLSPolicyAccepted(),
	}

	btp := BackendTLSPolicy{
		Source: &v1alpha2.BackendTLSPolicy{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "btp",
				Namespace: "service",
			},
			Spec: v1alpha2.BackendTLSPolicySpec{
				TargetRef: v1alpha2.PolicyTargetReferenceWithSectionName{
					PolicyTargetReference: v1alpha2.PolicyTargetReference{
						Group:     "",
						Kind:      "Service",
						Name:      "foo",
						Namespace: (*gatewayv1.Namespace)(helpers.GetPointer("service")),
					},
				},
				TLS: v1alpha2.BackendTLSPolicyConfig{
					Hostname: "foo.example.com",
					CACertRefs: []v1alpha2.LocalObjectReference{
						{
							Kind:  "ConfigMap",
							Name:  "configmap",
							Group: "",
						},
					},
				},
			},
		},
		Valid:        true,
		IsReferenced: true,
		Gateway:      types.NamespacedName{Namespace: "test", Name: "gateway-1"},
		Conditions:   btpAcceptedConds,
		CaCertRef:    types.NamespacedName{Namespace: "service", Name: "configmap"},
	}

	hr1Refs := []BackendRef{
		{
			SvcNsName:        types.NamespacedName{Namespace: "service", Name: "foo"},
			ServicePort:      v1.ServicePort{Port: 80},
			Valid:            true,
			Weight:           1,
			BackendTLSPolicy: &btp,
		},
	}

	hr3Refs := []BackendRef{
		{
			SvcNsName:        types.NamespacedName{Namespace: "service", Name: "foo"},
			ServicePort:      v1.ServicePort{Port: 80},
			Valid:            true,
			Weight:           1,
			BackendTLSPolicy: &btp,
		},
	}

	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "secret",
		},
		Data: map[string][]byte{
			v1.TLSCertKey:       cert,
			v1.TLSPrivateKeyKey: key,
		},
		Type: v1.SecretTypeTLS,
	}

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Labels: map[string]string{
				"app": "allowed",
			},
		},
	}

	createGateway := func(name string) *gatewayv1.Gateway {
		return &gatewayv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: gatewayv1.GatewaySpec{
				GatewayClassName: gcName,
				Listeners: []gatewayv1.Listener{
					{
						Name:     "listener-80-1",
						Hostname: nil,
						Port:     80,
						Protocol: gatewayv1.HTTPProtocolType,
						AllowedRoutes: &gatewayv1.AllowedRoutes{
							Namespaces: &gatewayv1.RouteNamespaces{
								From: helpers.GetPointer(gatewayv1.NamespacesFromSelector),
								Selector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"app": "allowed",
									},
								},
							},
						},
					},

					{
						Name:     "listener-443-1",
						Hostname: nil,
						Port:     443,
						TLS: &gatewayv1.GatewayTLSConfig{
							Mode: helpers.GetPointer(gatewayv1.TLSModeTerminate),
							CertificateRefs: []gatewayv1.SecretObjectReference{
								{
									Kind:      helpers.GetPointer[gatewayv1.Kind]("Secret"),
									Name:      gatewayv1.ObjectName(secret.Name),
									Namespace: helpers.GetPointer(gatewayv1.Namespace(secret.Namespace)),
								},
							},
						},
						Protocol: gatewayv1.HTTPSProtocolType,
					},
				},
			},
		}
	}

	gw1 := createGateway("gateway-1")
	gw2 := createGateway("gateway-2")

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "service", Name: "foo",
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Port: 80,
				},
			},
		},
	}

	rgSecret := &v1beta1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rg-secret",
			Namespace: "certificate",
		},
		Spec: v1beta1.ReferenceGrantSpec{
			From: []v1beta1.ReferenceGrantFrom{
				{
					Group:     gatewayv1.GroupName,
					Kind:      "Gateway",
					Namespace: "test",
				},
			},
			To: []v1beta1.ReferenceGrantTo{
				{
					Kind: "Secret",
				},
			},
		},
	}

	rgService := &v1beta1.ReferenceGrant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rg-service",
			Namespace: "service",
		},
		Spec: v1beta1.ReferenceGrantSpec{
			From: []v1beta1.ReferenceGrantFrom{
				{
					Group:     gatewayv1.GroupName,
					Kind:      "HTTPRoute",
					Namespace: "test",
				},
			},
			To: []v1beta1.ReferenceGrantTo{
				{
					Kind: "Service",
				},
			},
		},
	}

	proxy := &ngfAPI.NginxProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-proxy",
		},
		Spec: ngfAPI.NginxProxySpec{
			Telemetry: &ngfAPI.Telemetry{
				Exporter: &ngfAPI.TelemetryExporter{
					Endpoint:   "1.2.3.4:123",
					Interval:   helpers.GetPointer(ngfAPI.Duration("5s")),
					BatchSize:  helpers.GetPointer(int32(512)),
					BatchCount: helpers.GetPointer(int32(4)),
				},
				ServiceName: helpers.GetPointer("my-svc"),
			},
		},
	}

	createStateWithGatewayClass := func(gc *gatewayv1.GatewayClass) ClusterState {
		return ClusterState{
			GatewayClasses: map[types.NamespacedName]*gatewayv1.GatewayClass{
				client.ObjectKeyFromObject(gc): gc,
			},
			Gateways: map[types.NamespacedName]*gatewayv1.Gateway{
				client.ObjectKeyFromObject(gw1): gw1,
				client.ObjectKeyFromObject(gw2): gw2,
			},
			HTTPRoutes: map[types.NamespacedName]*gatewayv1.HTTPRoute{
				client.ObjectKeyFromObject(hr1): hr1,
				client.ObjectKeyFromObject(hr2): hr2,
				client.ObjectKeyFromObject(hr3): hr3,
			},
			Services: map[types.NamespacedName]*v1.Service{
				client.ObjectKeyFromObject(svc): svc,
			},
			Namespaces: map[types.NamespacedName]*v1.Namespace{
				client.ObjectKeyFromObject(ns): ns,
			},
			ReferenceGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				client.ObjectKeyFromObject(rgSecret):  rgSecret,
				client.ObjectKeyFromObject(rgService): rgService,
			},
			Secrets: map[types.NamespacedName]*v1.Secret{
				client.ObjectKeyFromObject(secret): secret,
			},
			BackendTLSPolicies: map[types.NamespacedName]*v1alpha2.BackendTLSPolicy{
				client.ObjectKeyFromObject(btp.Source): btp.Source,
			},
			ConfigMaps: map[types.NamespacedName]*v1.ConfigMap{
				client.ObjectKeyFromObject(cm): cm,
			},
			NginxProxies: map[types.NamespacedName]*ngfAPI.NginxProxy{
				client.ObjectKeyFromObject(proxy): proxy,
			},
		}
	}

	routeHR1 := &Route{
		Valid:      true,
		Attachable: true,
		Source:     hr1,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw1),
				Attachment: &ParentRefAttachmentStatus{
					Attached:          true,
					AcceptedHostnames: map[string][]string{"listener-80-1": {"foo.example.com"}},
				},
			},
		},
		Rules: []Rule{createValidRuleWithBackendRefs(hr1Refs)},
	}

	routeHR3 := &Route{
		Valid:      true,
		Attachable: true,
		Source:     hr3,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw1),
				Attachment: &ParentRefAttachmentStatus{
					Attached:          true,
					AcceptedHostnames: map[string][]string{"listener-443-1": {"foo.example.com"}},
				},
			},
		},
		Rules: []Rule{createValidRuleWithBackendRefs(hr3Refs)},
	}

	createExpectedGraphWithGatewayClass := func(gc *gatewayv1.GatewayClass) *Graph {
		return &Graph{
			GatewayClass: &GatewayClass{
				Source:     gc,
				Valid:      true,
				Conditions: []conditions.Condition{staticConds.NewGatewayClassResolvedRefs()},
			},
			Gateway: &Gateway{
				Source: gw1,
				Listeners: []*Listener{
					{
						Name:       "listener-80-1",
						Source:     gw1.Spec.Listeners[0],
						Valid:      true,
						Attachable: true,
						Routes: map[types.NamespacedName]*Route{
							{Namespace: "test", Name: "hr-1"}: routeHR1,
						},
						SupportedKinds:            []gatewayv1.RouteGroupKind{{Kind: "HTTPRoute"}},
						AllowedRouteLabelSelector: labels.SelectorFromSet(map[string]string{"app": "allowed"}),
					},
					{
						Name:       "listener-443-1",
						Source:     gw1.Spec.Listeners[1],
						Valid:      true,
						Attachable: true,
						Routes: map[types.NamespacedName]*Route{
							{Namespace: "test", Name: "hr-3"}: routeHR3,
						},
						ResolvedSecret: helpers.GetPointer(client.ObjectKeyFromObject(secret)),
						SupportedKinds: []gatewayv1.RouteGroupKind{{Kind: "HTTPRoute"}},
					},
				},
				Valid: true,
			},
			IgnoredGateways: map[types.NamespacedName]*gatewayv1.Gateway{
				{Namespace: "test", Name: "gateway-2"}: gw2,
			},
			Routes: map[types.NamespacedName]*Route{
				{Namespace: "test", Name: "hr-1"}: routeHR1,
				{Namespace: "test", Name: "hr-3"}: routeHR3,
			},
			ReferencedSecrets: map[types.NamespacedName]*Secret{
				client.ObjectKeyFromObject(secret): {
					Source: secret,
				},
			},
			ReferencedNamespaces: map[types.NamespacedName]*v1.Namespace{
				client.ObjectKeyFromObject(ns): ns,
			},
			ReferencedServices: map[types.NamespacedName]struct{}{
				client.ObjectKeyFromObject(svc): {},
			},
			ReferencedCaCertConfigMaps: map[types.NamespacedName]*CaCertConfigMap{
				client.ObjectKeyFromObject(cm): {
					Source: cm,
					CACert: []byte(caBlock),
				},
			},
			BackendTLSPolicies: map[types.NamespacedName]*BackendTLSPolicy{
				client.ObjectKeyFromObject(btp.Source): &btp,
			},
			NginxProxy: proxy,
		}
	}

	normalGC := &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: controllerName,
			ParametersRef: &gatewayv1.ParametersReference{
				Group: gatewayv1.Group("gateway.nginx.org"),
				Kind:  gatewayv1.Kind("NginxProxy"),
				Name:  "nginx-proxy",
			},
		},
	}
	differentControllerGC := &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: "different-controller",
		},
	}

	tests := []struct {
		store    ClusterState
		expected *Graph
		name     string
	}{
		{
			store:    createStateWithGatewayClass(normalGC),
			expected: createExpectedGraphWithGatewayClass(normalGC),
			name:     "normal case",
		},
		{
			store:    createStateWithGatewayClass(differentControllerGC),
			expected: &Graph{},
			name:     "gatewayclass belongs to a different controller",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result := BuildGraph(
				test.store,
				controllerName,
				gcName,
				validation.Validators{
					HTTPFieldsValidator: &validationfakes.FakeHTTPFieldsValidator{},
					GenericValidator:    &validationfakes.FakeGenericValidator{},
				},
				protectedPorts,
			)

			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}

func TestIsReferenced(t *testing.T) {
	baseSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "secret",
		},
	}
	sameNamespaceDifferentNameSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "secret-different-name",
		},
	}
	differentNamespaceSameNameSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-different-namespace",
			Name:      "secret",
		},
	}

	nsInGraph := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
			Labels: map[string]string{
				"app": "allowed",
			},
		},
	}
	nsNotInGraph := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "different-name",
			Labels: map[string]string{
				"app": "allowed",
			},
		},
	}

	serviceInGraph := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "serviceInGraph",
		},
	}
	serviceNotInGraph := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default",
			Name:      "serviceNotInGraph",
		},
	}
	serviceNotInGraphSameNameDifferentNS := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "not-default",
			Name:      "serviceInGraph",
		},
	}
	emptyService := &v1.Service{}

	createEndpointSlice := func(name string, svcName string) *discoveryV1.EndpointSlice {
		return &discoveryV1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "default",
				Name:      name,
				Labels:    map[string]string{index.KubernetesServiceNameLabel: svcName},
			},
		}
	}
	endpointSliceInGraph := createEndpointSlice("endpointSliceInGraph", "serviceInGraph")
	endpointSliceNotInGraph := createEndpointSlice("endpointSliceNotInGraph", "serviceNotInGraph")
	emptyEndpointSlice := &discoveryV1.EndpointSlice{}

	gw := &Gateway{
		Listeners: []*Listener{
			{
				Name:                      "listener-1",
				Valid:                     true,
				AllowedRouteLabelSelector: labels.SelectorFromSet(map[string]string{"apples": "oranges"}),
			},
		},
		Valid: true,
	}

	nsNotInGraphButInGateway := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "notInGraphButInGateway",
			Labels: map[string]string{
				"apples": "oranges",
			},
		},
	}

	baseConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "configmap",
		},
	}
	sameNamespaceDifferentNameConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "configmap-different-name",
		},
	}
	differentNamespaceSameNameConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test-different-namespace",
			Name:      "configmap",
		},
	}

	gcWithNginxProxy := &GatewayClass{
		Source: &gatewayv1.GatewayClass{
			Spec: gatewayv1.GatewayClassSpec{
				ParametersRef: &gatewayv1.ParametersReference{
					Group: ngfAPI.GroupName,
					Kind:  gatewayv1.Kind("NginxProxy"),
					Name:  "nginx-proxy-in-gc",
				},
			},
		},
	}

	npInGraph := &ngfAPI.NginxProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-proxy",
		},
	}

	npNotInGraphButInGatewayClass := &ngfAPI.NginxProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-proxy-in-gc",
		},
	}

	npNotInGraph := &ngfAPI.NginxProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name: "nginx-proxy-not-referenced",
		},
	}

	graph := &Graph{
		Gateway: gw,
		ReferencedSecrets: map[types.NamespacedName]*Secret{
			client.ObjectKeyFromObject(baseSecret): {
				Source: baseSecret,
			},
		},
		ReferencedNamespaces: map[types.NamespacedName]*v1.Namespace{
			client.ObjectKeyFromObject(nsInGraph): nsInGraph,
		},
		ReferencedServices: map[types.NamespacedName]struct{}{
			client.ObjectKeyFromObject(serviceInGraph): {},
		},
		ReferencedCaCertConfigMaps: map[types.NamespacedName]*CaCertConfigMap{
			client.ObjectKeyFromObject(baseConfigMap): {
				Source: baseConfigMap,
				CACert: []byte(caBlock),
			},
		},
		NginxProxy: npInGraph,
	}

	tests := []struct {
		graph    *Graph
		gc       *GatewayClass
		resource client.Object
		name     string
		expected bool
	}{
		// Namespace tests
		{
			name:     "Namespace in graph's ReferencedNamespaces is referenced",
			resource: nsInGraph,
			graph:    graph,
			expected: true,
		},
		{
			name:     "Namespace with a different name but same labels is not referenced",
			resource: nsNotInGraph,
			graph:    graph,
			expected: false,
		},
		{
			name: "Namespace not in ReferencedNamespaces but in Gateway Listener's AllowedRouteLabelSelector" +
				" is referenced",
			resource: nsNotInGraphButInGateway,
			graph:    graph,
			expected: true,
		},

		// Secret tests
		{
			name:     "Secret in graph's ReferencedSecrets is referenced",
			resource: baseSecret,
			graph:    graph,
			expected: true,
		},
		{
			name:     "Secret not in ReferencedSecrets with same Namespace and different Name is not referenced",
			resource: sameNamespaceDifferentNameSecret,
			graph:    graph,
			expected: false,
		},
		{
			name:     "Secret not in ReferencedSecrets with different Namespace and same Name is not referenced",
			resource: differentNamespaceSameNameSecret,
			graph:    graph,
			expected: false,
		},

		// Service tests
		{
			name:     "Service is referenced",
			resource: serviceInGraph,
			graph:    graph,
			expected: true,
		},
		{
			name:     "Service is not referenced",
			resource: serviceNotInGraph,
			graph:    graph,
			expected: false,
		},
		{
			name:     "Service with same name but different namespace is not referenced",
			resource: serviceNotInGraphSameNameDifferentNS,
			graph:    graph,
			expected: false,
		},
		{
			name:     "Empty Service",
			resource: emptyService,
			graph:    graph,
			expected: false,
		},

		// EndpointSlice tests
		{
			name:     "EndpointSlice with Service owner in graph's ReferencedServices is referenced",
			resource: endpointSliceInGraph,
			graph:    graph,
			expected: true,
		},
		{
			name:     "EndpointSlice with Service owner not in graph's ReferencedServices is not referenced",
			resource: endpointSliceNotInGraph,
			graph:    graph,
			expected: false,
		},
		{
			name:     "Empty EndpointSlice",
			resource: emptyEndpointSlice,
			graph:    graph,
			expected: false,
		},

		// ConfigMap tests
		{
			name:     "ConfigMap in graph's ReferencedConfigMaps is referenced",
			resource: baseConfigMap,
			graph:    graph,
			expected: true,
		},
		{
			name:     "ConfigMap not in ReferencedConfigMaps with same Namespace and different Name is not referenced",
			resource: sameNamespaceDifferentNameConfigMap,
			graph:    graph,
			expected: false,
		},
		{
			name:     "ConfigMap not in ReferencedConfigMaps with different Namespace and same Name is not referenced",
			resource: differentNamespaceSameNameConfigMap,
			graph:    graph,
			expected: false,
		},

		// NginxProxy tests
		{
			name:     "NginxProxy in the Graph is referenced",
			resource: npInGraph,
			graph:    graph,
			expected: true,
		},
		{
			name:     "NginxProxy is not yet in Graph but is referenced in GatewayClass",
			resource: npNotInGraphButInGatewayClass,
			gc:       gcWithNginxProxy,
			graph:    graph,
			expected: true,
		},
		{
			name:     "NginxProxy not in Graph or referenced in GatewayClass",
			resource: npNotInGraph,
			gc:       gcWithNginxProxy,
			graph:    graph,
			expected: false,
		},

		// Edge cases
		{
			name:     "Resource is not supported by IsReferenced",
			resource: &gatewayv1.HTTPRoute{},
			graph:    graph,
			expected: false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			test.graph.GatewayClass = test.gc
			result := test.graph.IsReferenced(test.resource, client.ObjectKeyFromObject(test.resource))
			g.Expect(result).To(Equal(test.expected))
		})
	}
}
