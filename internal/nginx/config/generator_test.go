package config

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/statefakes"
)

func TestGenerateForHost(t *testing.T) {
	generator := NewGeneratorImpl(&statefakes.FakeServiceStore{})

	host := state.Host{Value: "example.com"}

	cfg, warnings := generator.GenerateForHost(host)

	if len(cfg) == 0 {
		t.Errorf("GenerateForHost() generated empty config")
	}
	if len(warnings) > 0 {
		t.Errorf("GenerateForHost() returned unexpected warnings: %v", warnings)
	}
}

func TestGenerate(t *testing.T) {
	hr := &v1alpha2.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "route1",
		},
		Spec: v1alpha2.HTTPRouteSpec{
			Hostnames: []v1alpha2.Hostname{
				"cafe.example.com",
			},
			Rules: []v1alpha2.HTTPRouteRule{
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/"),
							},
						},
					},
					BackendRefs: []v1alpha2.HTTPBackendRef{
						{
							BackendRef: v1alpha2.BackendRef{
								BackendObjectReference: v1alpha2.BackendObjectReference{
									Name:      "service1",
									Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
									Port:      (*v1alpha2.PortNumber)(helpers.GetInt32Pointer(80)),
								},
							},
						},
					},
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/test"),
							},
						},
					},
					BackendRefs: nil, // no backend refs will cause warnings
				},
			},
		},
	}

	host := state.Host{
		Value: "example.com",
		PathRouteGroups: []state.PathRouteGroup{
			{
				Path: "/",
				Routes: []state.Route{
					{
						MatchIdx: 0,
						RuleIdx:  0,
						Source:   hr,
					},
				},
			},
			{
				Path: "/test",
				Routes: []state.Route{
					{
						MatchIdx: 0,
						RuleIdx:  1,
						Source:   hr,
					},
				},
			},
		},
	}

	fakeServiceStore := &statefakes.FakeServiceStore{}
	fakeServiceStore.ResolveReturns("10.0.0.1", nil)

	expected := server{
		ServerName: "example.com",
		Locations: []location{
			{
				Path:      "/",
				ProxyPass: "http://10.0.0.1:80",
			},
			{
				Path:      "/test",
				ProxyPass: "http://" + nginx502Server,
			},
		},
	}
	expectedWarnings := Warnings{
		hr: []string{"empty backend refs"},
	}

	result, warnings := generate(host, fakeServiceStore)

	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("generate() mismatch (-want +got):\n%s", diff)
	}
	if diff := cmp.Diff(expectedWarnings, warnings); diff != "" {
		t.Errorf("generate() mismatch on warnings (-want +got):\n%s", diff)
	}
}

func TestGenerateProxyPass(t *testing.T) {
	expected := "http://10.0.0.1:80"

	result := generateProxyPass("10.0.0.1:80")
	if result != expected {
		t.Errorf("generateProxyPass() returned %s but expected %s", result, expected)
	}

	expected = "http://" + nginx502Server

	result = generateProxyPass("")
	if result != expected {
		t.Errorf("generateProxyPass() returned %s but expected %s", result, expected)
	}
}

func TestGetBackendAddress(t *testing.T) {
	getNormalRefs := func() []v1alpha2.HTTPBackendRef {
		return []v1alpha2.HTTPBackendRef{
			{
				BackendRef: v1alpha2.BackendRef{
					BackendObjectReference: v1alpha2.BackendObjectReference{
						Group:     (*v1alpha2.Group)(helpers.GetStringPointer("networking.k8s.io")),
						Kind:      (*v1alpha2.Kind)(helpers.GetStringPointer("Service")),
						Name:      "service1",
						Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
						Port:      (*v1alpha2.PortNumber)(helpers.GetInt32Pointer(80)),
					},
				},
			},
		}
	}

	getModifiedRefs := func(mod func([]v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
		return mod(getNormalRefs())
	}

	tests := []struct {
		refs                      []v1alpha2.HTTPBackendRef
		parentNS                  string
		storeAddress              string
		storeErr                  error
		expectedResolverCallCount int
		expectedNsName            types.NamespacedName
		expectedAddress           string
		expectErr                 bool
		msg                       string
	}{
		{
			refs:                      getNormalRefs(),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "normal case",
		},
		{
			refs: getModifiedRefs(func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
				refs[0].BackendRef.Namespace = nil
				return refs
			}),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "normal case with implicit namespace",
		},
		{
			refs: getModifiedRefs(func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
				refs[0].BackendRef.Group = nil
				refs[0].BackendRef.Kind = nil
				return refs
			}),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "normal case with implicit service",
		},
		{
			refs: getModifiedRefs(func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
				secondRef := refs[0].DeepCopy()
				secondRef.Name = "service2"
				return append(refs, *secondRef)
			}),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "10.0.0.1:80",
			expectErr:                 false,
			msg:                       "first backend ref is used",
		},
		{
			refs: getModifiedRefs(func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
				refs[0].BackendRef.Kind = (*v1alpha2.Kind)(helpers.GetStringPointer("NotService"))
				return refs
			}),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 0,
			expectedNsName:            types.NamespacedName{},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "not a service Kind",
		},
		{
			refs:                      nil,
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 0,
			expectedNsName:            types.NamespacedName{},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "no refs",
		},
		{
			refs: getModifiedRefs(func(refs []v1alpha2.HTTPBackendRef) []v1alpha2.HTTPBackendRef {
				refs[0].BackendRef.Port = nil
				return refs
			}),
			parentNS:                  "test",
			storeAddress:              "10.0.0.1",
			storeErr:                  nil,
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "no port",
		},
		{
			refs:                      getNormalRefs(),
			parentNS:                  "test",
			storeAddress:              "",
			storeErr:                  errors.New(""),
			expectedResolverCallCount: 1,
			expectedNsName:            types.NamespacedName{Namespace: "test", Name: "service1"},
			expectedAddress:           "",
			expectErr:                 true,
			msg:                       "service doesn't exist",
		},
	}

	for _, test := range tests {
		fakeServiceStore := &statefakes.FakeServiceStore{}
		fakeServiceStore.ResolveReturns(test.storeAddress, test.storeErr)

		result, err := getBackendAddress(test.refs, test.parentNS, fakeServiceStore)
		if result != test.expectedAddress {
			t.Errorf("getBackendAddress() returned %s but expected %s for case %q", result, test.expectedAddress, test.msg)
		}

		if test.expectErr {
			if err == nil {
				t.Errorf("getBackendAddress() didn't return any error for case %q", test.msg)
			}
		} else {
			if err != nil {
				t.Errorf("getBackendAddress() returned unexpected error %v for case %q", err, test.msg)
			}
		}

		callCount := fakeServiceStore.ResolveCallCount()
		if callCount != test.expectedResolverCallCount {
			t.Errorf("getBackendAddress() called fakeServiceStore.Resolve %d times but expected %d for case %q", callCount, test.expectedResolverCallCount, test.msg)
		}

		if test.expectedResolverCallCount == 0 {
			continue
		}

		nsname := fakeServiceStore.ResolveArgsForCall(0)
		if nsname != test.expectedNsName {
			t.Errorf("getBackendAddress() called fakeServiceStore.Resolve with %v but expected %v for case %q", nsname, test.expectedNsName, test.msg)
		}
	}
}
