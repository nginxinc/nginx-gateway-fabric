package graph

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func getNormalRef() v1beta1.BackendRef {
	return v1beta1.BackendRef{
		BackendObjectReference: v1beta1.BackendObjectReference{
			Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Service")),
			Name:      "service1",
			Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
			Port:      (*v1beta1.PortNumber)(helpers.GetInt32Pointer(80)),
		},
		Weight: helpers.GetInt32Pointer(1),
	}
}

func getModifiedRef(mod func(ref v1beta1.BackendRef) v1beta1.BackendRef) v1beta1.BackendRef {
	return mod(getNormalRef())
}

func TestValidateBackendRef(t *testing.T) {
	tests := []struct {
		ref    v1beta1.BackendRef
		msg    string
		expErr bool
	}{
		{
			msg:    "normal case",
			ref:    getNormalRef(),
			expErr: false,
		},
		{
			msg: "normal case with implicit namespace",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Namespace = nil
				return backend
			}),
			expErr: false,
		},
		{
			msg: "normal case with implicit kind Service",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Kind = nil
				return backend
			}),
			expErr: false,
		},
		{
			msg: "not a service kind",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Kind = (*v1beta1.Kind)(helpers.GetStringPointer("NotService"))
				return backend
			}),
			expErr: true,
		},
		{
			msg: "invalid namespace",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Namespace = (*v1beta1.Namespace)(helpers.GetStringPointer("invalid"))
				return backend
			}),
			expErr: true,
		},
		{
			msg: "missing port",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Port = nil
				return backend
			}),
			expErr: true,
		},
	}

	for _, test := range tests {
		err := validateBackendRef(test.ref, "test")
		errOccurred := err != nil
		if errOccurred != test.expErr {
			t.Errorf("validateBackendRef() returned incorrect error for %q; error: %v", test.msg, err)
		}
	}
}

func TestGetServiceAndPortFromRef(t *testing.T) {
	svc1 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service1",
			Namespace: "test",
		},
	}

	svc2 := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service2",
			Namespace: "test",
		},
	}

	tests := []struct {
		ref        v1beta1.BackendRef
		expService *v1.Service
		msg        string
		expPort    int32
		expErr     bool
	}{
		{
			msg:        "normal case",
			ref:        getNormalRef(),
			expService: svc1,
			expPort:    80,
		},
		{
			msg: "invalid backend ref",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Port = nil
				return backend
			}),
			expErr: true,
		},
		{
			msg: "service does not exist",
			ref: getModifiedRef(func(backend v1beta1.BackendRef) v1beta1.BackendRef {
				backend.Name = "dne"
				return backend
			}),
			expErr: true,
		},
	}

	services := map[types.NamespacedName]*v1.Service{
		{Namespace: "test", Name: "service1"}: svc1,
		{Namespace: "test", Name: "service2"}: svc2,
	}

	for _, test := range tests {
		svc, port, err := getServiceAndPortFromRef(test.ref, "test", services)

		errOccurred := err != nil
		if errOccurred != test.expErr {
			t.Errorf("getServiceAndPortFromRef() returned incorrect error for %q; error: %v", test.msg, err)
		}

		if svc != test.expService {
			t.Errorf("getServiceAndPortFromRef() returned incorrect service for %q; expected: %v, got: %v",
				test.msg, test.expService, svc)
		}

		if port != test.expPort {
			t.Errorf("getServiceAndPortFromRef() returned incorrect port for %q; expected: %d, got: %d",
				test.msg, test.expPort, port)
		}
	}
}

func TestResolveBackendRefs(t *testing.T) {
	createRoute := func(name string, kind string, serviceNames ...string) *v1beta1.HTTPRoute {
		hr := &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      name,
			},
		}

		hr.Spec.Rules = make([]v1beta1.HTTPRouteRule, len(serviceNames))

		for idx, svcName := range serviceNames {
			hr.Spec.Rules[idx] = v1beta1.HTTPRouteRule{
				BackendRefs: []v1beta1.HTTPBackendRef{
					{
						BackendRef: v1beta1.BackendRef{
							BackendObjectReference: v1beta1.BackendObjectReference{
								Kind:      (*v1beta1.Kind)(helpers.GetStringPointer(kind)),
								Name:      v1beta1.ObjectName(svcName),
								Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
								Port:      (*v1beta1.PortNumber)(helpers.GetInt32Pointer(80)),
							},
							Weight: nil, // should use default weight of 1
						},
					},
					{
						BackendRef: v1beta1.BackendRef{
							BackendObjectReference: v1beta1.BackendObjectReference{
								Kind:      (*v1beta1.Kind)(helpers.GetStringPointer(kind)),
								Name:      v1beta1.ObjectName(svcName),
								Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
								Port:      (*v1beta1.PortNumber)(helpers.GetInt32Pointer(81)),
							},
							Weight: helpers.GetInt32Pointer(5),
						},
					},
				},
			}
		}
		return hr
	}

	removeRefs := func(route *v1beta1.HTTPRoute) *v1beta1.HTTPRoute {
		route.Spec.Rules[0].BackendRefs = nil
		return route
	}

	hr1 := createRoute("hr1", "Service", "svc1", "svc2", "svc3")
	hr2 := createRoute("hr2", "Service", "svc1", "svc4")
	hr3 := createRoute("hr3", "NotService", "not-svc")
	hr4 := removeRefs(createRoute("hr4", "Service", "no-backend-refs"))

	routes := map[types.NamespacedName]*Route{
		{Namespace: "test", Name: "hr1"}: {
			Source: hr1,
		},
		{Namespace: "test", Name: "hr2"}: {
			Source: hr2,
		},
		{Namespace: "test", Name: "hr3"}: {
			Source: hr3,
		},
		{Namespace: "test", Name: "hr4"}: {
			Source: hr4,
		},
	}

	svc1 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "svc1"}}
	svc2 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "svc2"}}
	svc3 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "svc3"}}
	svc4 := &v1.Service{ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: "svc4"}}

	services := map[types.NamespacedName]*v1.Service{
		{Namespace: "test", Name: "svc1"}: svc1,
		{Namespace: "test", Name: "svc2"}: svc2,
		{Namespace: "test", Name: "svc3"}: svc3,
		{Namespace: "test", Name: "svc4"}: svc4,
	}

	expRoutes := map[types.NamespacedName]*Route{
		{Namespace: "test", Name: "hr1"}: {
			Source: hr1,
			BackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hr1),
					RuleIdx: 0,
					Errors:  []string{},
					Backends: []BackendRef{
						{
							Name:   "test_svc1_80",
							Svc:    svc1,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
						{
							Name:   "test_svc1_81",
							Svc:    svc1,
							Port:   81,
							Valid:  true,
							Weight: 5,
						},
					},
				},
				{
					Source:  client.ObjectKeyFromObject(hr1),
					RuleIdx: 1,
					Errors:  []string{},
					Backends: []BackendRef{
						{
							Name:   "test_svc2_80",
							Svc:    svc2,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
						{
							Name:   "test_svc2_81",
							Svc:    svc2,
							Port:   81,
							Valid:  true,
							Weight: 5,
						},
					},
				},
				{
					Source:  client.ObjectKeyFromObject(hr1),
					RuleIdx: 2,
					Errors:  []string{},
					Backends: []BackendRef{
						{
							Name:   "test_svc3_80",
							Svc:    svc3,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
						{
							Name:   "test_svc3_81",
							Svc:    svc3,
							Port:   81,
							Valid:  true,
							Weight: 5,
						},
					},
				},
			},
		},
		{Namespace: "test", Name: "hr2"}: {
			Source: hr2,
			BackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hr2),
					RuleIdx: 0,
					Errors:  []string{},
					Backends: []BackendRef{
						{
							Name:   "test_svc1_80",
							Svc:    svc1,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
						{
							Name:   "test_svc1_81",
							Svc:    svc1,
							Port:   81,
							Valid:  true,
							Weight: 5,
						},
					},
				},
				{
					Source:  client.ObjectKeyFromObject(hr2),
					RuleIdx: 1,
					Errors:  []string{},
					Backends: []BackendRef{
						{
							Name:   "test_svc4_80",
							Svc:    svc4,
							Port:   80,
							Valid:  true,
							Weight: 1,
						},
						{
							Name:   "test_svc4_81",
							Svc:    svc4,
							Port:   81,
							Valid:  true,
							Weight: 5,
						},
					},
				},
			},
		},
		{Namespace: "test", Name: "hr3"}: {
			Source: hr3,
			BackendGroups: []BackendGroup{
				{
					Errors: []string{
						"the Kind must be Service; got NotService",
						"the Kind must be Service; got NotService",
					},
					Source:  client.ObjectKeyFromObject(hr3),
					RuleIdx: 0,
					Backends: []BackendRef{
						{
							Weight: 1,
						},
						{
							Weight: 5,
						},
					},
				},
			},
		},
		{Namespace: "test", Name: "hr4"}: {
			Source: hr4,
			BackendGroups: []BackendGroup{
				{
					Source:  client.ObjectKeyFromObject(hr4),
					RuleIdx: 0,
				},
			},
		},
	}

	addBackendGroupsToRoutes(routes, services)

	if diff := cmp.Diff(expRoutes, routes); diff != "" {
		t.Errorf("resolveBackendRefs() mismatch on routes (-want +got):\n%s", diff)
	}
}
