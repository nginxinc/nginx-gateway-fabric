//nolint:gosec
package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	v1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
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

	fooSvc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "service"}}

	hr1Refs := []BackendRef{
		{
			Svc:    fooSvc,
			Port:   80,
			Valid:  true,
			Weight: 1,
		},
	}

	hr3Refs := []BackendRef{
		{
			Svc:    fooSvc,
			Port:   80,
			Valid:  true,
			Weight: 1,
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

	svc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "service", Name: "foo"}}

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
			ReferenceGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				client.ObjectKeyFromObject(rgSecret):  rgSecret,
				client.ObjectKeyFromObject(rgService): rgService,
			},
			Secrets: map[types.NamespacedName]*v1.Secret{
				client.ObjectKeyFromObject(secret): secret,
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
				Source: gc,
				Valid:  true,
			},
			Gateway: &Gateway{
				Source: gw1,
				Listeners: map[string]*Listener{
					"listener-80-1": {
						Source:     gw1.Spec.Listeners[0],
						Valid:      true,
						Attachable: true,
						Routes: map[types.NamespacedName]*Route{
							{Namespace: "test", Name: "hr-1"}: routeHR1,
						},
						SupportedKinds: []gatewayv1.RouteGroupKind{{Kind: "HTTPRoute"}},
					},
					"listener-443-1": {
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
		}
	}

	normalGC := &gatewayv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: gatewayv1.GatewayClassSpec{
			ControllerName: controllerName,
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
				validation.Validators{HTTPFieldsValidator: &validationfakes.FakeHTTPFieldsValidator{}},
				protectedPorts,
			)

			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
