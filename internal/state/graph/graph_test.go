//nolint:gosec
package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets/secretsfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation/validationfakes"
)

func TestBuildGraph(t *testing.T) {
	const (
		gcName         = "my-class"
		controllerName = "my.controller"
		secretPath     = "/etc/nginx/secrets/test_secret"
	)

	createValidRuleWithBackendGroup := func(group BackendGroup) Rule {
		return Rule{
			ValidMatches: true,
			ValidFilters: true,
			BackendGroup: group,
		}
	}

	createRoute := func(name string, gatewayName string, listenerName string) *v1beta1.HTTPRoute {
		return &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1beta1.HTTPRouteSpec{
				CommonRouteSpec: v1beta1.CommonRouteSpec{
					ParentRefs: []v1beta1.ParentReference{
						{
							Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
							Name:        v1beta1.ObjectName(gatewayName),
							SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer(listenerName)),
						},
					},
				},
				Hostnames: []v1beta1.Hostname{
					"foo.example.com",
				},
				Rules: []v1beta1.HTTPRouteRule{
					{
						Matches: []v1beta1.HTTPRouteMatch{
							{
								Path: &v1beta1.HTTPPathMatch{
									Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
									Value: helpers.GetStringPointer("/"),
								},
							},
						},
						BackendRefs: []v1beta1.HTTPBackendRef{
							{
								BackendRef: v1beta1.BackendRef{
									BackendObjectReference: v1beta1.BackendObjectReference{
										Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Service")),
										Name:      "foo",
										Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
										Port:      (*v1beta1.PortNumber)(helpers.GetInt32Pointer(80)),
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

	fooSvc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: "test"}}

	hr1Group := BackendGroup{
		Source:  types.NamespacedName{Namespace: hr1.Namespace, Name: hr1.Name},
		RuleIdx: 0,
		Backends: []BackendRef{
			{
				Name:   "test_foo_80",
				Svc:    fooSvc,
				Port:   80,
				Valid:  true,
				Weight: 1,
			},
		},
	}

	hr3Group := BackendGroup{
		Source:  types.NamespacedName{Namespace: hr3.Namespace, Name: hr3.Name},
		RuleIdx: 0,
		Backends: []BackendRef{
			{
				Name:   "test_foo_80",
				Svc:    fooSvc,
				Port:   80,
				Valid:  true,
				Weight: 1,
			},
		},
	}

	createGateway := func(name string) *v1beta1.Gateway {
		return &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
			Spec: v1beta1.GatewaySpec{
				GatewayClassName: gcName,
				Listeners: []v1beta1.Listener{
					{
						Name:     "listener-80-1",
						Hostname: nil,
						Port:     80,
						Protocol: v1beta1.HTTPProtocolType,
					},

					{
						Name:     "listener-443-1",
						Hostname: nil,
						Port:     443,
						TLS: &v1beta1.GatewayTLSConfig{
							Mode: helpers.GetTLSModePointer(v1beta1.TLSModeTerminate),
							CertificateRefs: []v1beta1.SecretObjectReference{
								{
									Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Secret")),
									Name:      "secret",
									Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
								},
							},
						},
						Protocol: v1beta1.HTTPSProtocolType,
					},
				},
			},
		}
	}

	gw1 := createGateway("gateway-1")
	gw2 := createGateway("gateway-2")

	svc := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "foo"}}

	createStateWithGatewayClass := func(gc *v1beta1.GatewayClass) ClusterState {
		return ClusterState{
			GatewayClasses: map[types.NamespacedName]*v1beta1.GatewayClass{
				client.ObjectKeyFromObject(gc): gc,
			},
			Gateways: map[types.NamespacedName]*v1beta1.Gateway{
				client.ObjectKeyFromObject(gw1): gw1,
				client.ObjectKeyFromObject(gw2): gw2,
			},
			HTTPRoutes: map[types.NamespacedName]*v1beta1.HTTPRoute{
				client.ObjectKeyFromObject(hr1): hr1,
				client.ObjectKeyFromObject(hr2): hr2,
				client.ObjectKeyFromObject(hr3): hr3,
			},
			Services: map[types.NamespacedName]*v1.Service{
				client.ObjectKeyFromObject(svc): svc,
			},
		}
	}

	routeHR1 := &Route{
		Valid:  true,
		Source: hr1,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw1),
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Rules: []Rule{createValidRuleWithBackendGroup(hr1Group)},
	}

	routeHR3 := &Route{
		Valid:  true,
		Source: hr3,
		ParentRefs: []ParentRef{
			{
				Idx:     0,
				Gateway: client.ObjectKeyFromObject(gw1),
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Rules: []Rule{createValidRuleWithBackendGroup(hr3Group)},
	}

	secretMemoryMgr := &secretsfakes.FakeSecretDiskMemoryManager{}
	secretMemoryMgr.RequestCalls(func(nsname types.NamespacedName) (string, error) {
		if (nsname == types.NamespacedName{Namespace: "test", Name: "secret"}) {
			return secretPath, nil
		}
		panic("unexpected secret request")
	})

	createExpectedGraphWithGatewayClass := func(gc *v1beta1.GatewayClass) *Graph {
		return &Graph{
			GatewayClass: &GatewayClass{
				Source: gc,
				Valid:  true,
			},
			Gateway: &Gateway{
				Source: gw1,
				Listeners: map[string]*Listener{
					"listener-80-1": {
						Source: gw1.Spec.Listeners[0],
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{
							{Namespace: "test", Name: "hr-1"}: routeHR1,
						},
						AcceptedHostnames: map[string]struct{}{
							"foo.example.com": {},
						},
					},
					"listener-443-1": {
						Source: gw1.Spec.Listeners[1],
						Valid:  true,
						Routes: map[types.NamespacedName]*Route{
							{Namespace: "test", Name: "hr-3"}: routeHR3,
						},
						AcceptedHostnames: map[string]struct{}{
							"foo.example.com": {},
						},
						SecretPath: secretPath,
					},
				},
			},
			IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{
				{Namespace: "test", Name: "gateway-2"}: gw2,
			},
			Routes: map[types.NamespacedName]*Route{
				{Namespace: "test", Name: "hr-1"}: routeHR1,
				{Namespace: "test", Name: "hr-3"}: routeHR3,
			},
		}
	}

	normalGC := &v1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: controllerName,
		},
	}
	differentControllerGC := &v1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: gcName,
		},
		Spec: v1beta1.GatewayClassSpec{
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
			g := NewGomegaWithT(t)

			result := BuildGraph(
				test.store,
				controllerName,
				gcName,
				secretMemoryMgr,
				validation.Validators{HTTPFieldsValidator: &validationfakes.FakeHTTPFieldsValidator{}},
			)

			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
