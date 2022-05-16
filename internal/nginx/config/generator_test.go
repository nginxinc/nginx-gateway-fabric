package config

import (
	"encoding/json"
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
							Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodPost),
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
							Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodGet),
							Headers: []v1alpha2.HTTPHeaderMatch{
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
									Name:  "vErsIon", // header names and values should be normalized to lowercase
									Value: "V1",
								},
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
									Name:  "test",
									Value: "foo",
								},
								{
									Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
									Name:  "my-header",
									Value: "my-value",
								},
							},
							QueryParams: []v1alpha2.HTTPQueryParamMatch{
								{
									Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
									Name:  "GrEat", // query names and values should not be normalized to lowercase
									Value: "EXAMPLE",
								},
								{
									Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
									Name:  "test",
									Value: "foo=bar",
								},
							},
						},
					},
					BackendRefs: nil, // no backend refs will cause warnings
				},
				{
					Matches: []v1alpha2.HTTPRouteMatch{
						{
							Path: &v1alpha2.HTTPPathMatch{
								Value: helpers.GetStringPointer("/path-only"),
							},
							// matches that only have path specified will not generate an internal location block
						},
					},
					BackendRefs: []v1alpha2.HTTPBackendRef{
						{
							BackendRef: v1alpha2.BackendRef{
								BackendObjectReference: v1alpha2.BackendObjectReference{
									Name:      "service2",
									Namespace: (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
									Port:      (*v1alpha2.PortNumber)(helpers.GetInt32Pointer(80)),
								},
							},
						},
					},
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
			{
				Path: "/path-only",
				Routes: []state.Route{
					{
						MatchIdx: 0,
						RuleIdx:  2,
						Source:   hr,
					},
				},
			},
		},
	}

	fakeServiceStore := &statefakes.FakeServiceStore{}
	fakeServiceStore.ResolveReturns("10.0.0.1", nil)

	expectedMatchString := func(m httpMatch) string {
		b, err := json.Marshal(m)
		if err != nil {
			t.Errorf("error marshaling test match: %v", err)
		}
		return string(b)
	}

	slashMatches := httpMatch{Method: v1alpha2.HTTPMethodPost, RedirectPath: "/_route0"}
	testMatches := httpMatch{
		Method:       v1alpha2.HTTPMethodGet,
		Headers:      []string{"version:v1", "test:foo", "my-header:my-value"},
		QueryParams:  []string{"GrEat=EXAMPLE", "test=foo=bar"},
		RedirectPath: "/test_route0",
	}

	expected := server{
		ServerName: "example.com",
		Locations: []location{
			{
				Path:         "/",
				HTTPMatchVar: expectedMatchString(slashMatches),
			},
			{
				Path:      "/_route0",
				Internal:  true,
				ProxyPass: "http://10.0.0.1:80",
			},
			{
				Path:         "/test",
				HTTPMatchVar: expectedMatchString(testMatches),
			},
			{
				Path:      "/test_route0",
				Internal:  true,
				ProxyPass: "http://" + nginx502Server,
			},
			{
				Path:      "/path-only",
				ProxyPass: "http://10.0.0.1:80",
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

func TestGenerateMatchLocation(t *testing.T) {
	expected := location{
		Path:      "/path",
		Internal:  true,
		ProxyPass: "http://10.0.0.1:80",
	}

	result := generateMatchLocation("/path", "10.0.0.1:80")
	if result != expected {
		t.Errorf("generateMatchLocation() returned %v but expected %v", result, expected)
	}
}

func TestCreatePathForMatch(t *testing.T) {
	expected := "/path_route1"

	result := createPathForMatch("/path", 1)
	if result != expected {
		t.Errorf("createPathForMatch() returned %q but expected %q", result, expected)
	}
}

func TestCreateArgKeyValString(t *testing.T) {
	expected := "key=value"

	result := createQueryParamKeyValString(v1alpha2.HTTPQueryParamMatch{
		Name:  "key",
		Value: "value",
	})
	if result != expected {
		t.Errorf("createQueryParamKeyValString() returned %q but expected %q", result, expected)
	}

	expected = "KeY=vaLUe=="

	result = createQueryParamKeyValString(v1alpha2.HTTPQueryParamMatch{
		Name:  "KeY",
		Value: "vaLUe==",
	})
	if result != expected {
		t.Errorf("createQueryParamKeyValString() returned %q but expected %q", result, expected)
	}
}

func TestCreateHeaderKeyValString(t *testing.T) {
	expected := "key:value"

	result := createHeaderKeyValString(v1alpha2.HTTPHeaderMatch{
		Name:  "kEy",
		Value: "vALUe",
	})

	if result != expected {
		t.Errorf("createHeaderKeyValString() returned %q but expected %q", result, expected)
	}
}

func TestMatchLocationNeeded(t *testing.T) {
	tests := []struct {
		match    v1alpha2.HTTPRouteMatch
		expected bool
		msg      string
	}{
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
			},
			expected: false,
			msg:      "path only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				Method: helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodGet),
			},
			expected: true,
			msg:      "method defined in match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				Headers: []v1alpha2.HTTPHeaderMatch{
					{
						Name:  "header",
						Value: "val",
					},
				},
			},
			expected: true,
			msg:      "headers defined in match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{

				Path: &v1alpha2.HTTPPathMatch{
					Value: helpers.GetStringPointer("/path"),
				},
				QueryParams: []v1alpha2.HTTPQueryParamMatch{
					{
						Name:  "arg",
						Value: "val",
					},
				},
			},
			expected: true,
			msg:      "query params defined in match",
		},
	}

	for _, tc := range tests {
		result := matchLocationNeeded(tc.match)

		if result != tc.expected {
			t.Errorf("matchLocationNeeded() returned %t but expected %t for test case %q", result, tc.expected, tc.msg)
		}
	}
}

func TestCreateHTTPMatch(t *testing.T) {
	testPath := "/internal_loc"
	testMethodMatch := helpers.GetHTTPMethodPointer(v1alpha2.HTTPMethodPut)
	testHeaderMatches := []v1alpha2.HTTPHeaderMatch{
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
			Name:  "header-1",
			Value: "val-1",
		},
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
			Name:  "header-2",
			Value: "val-2",
		},
		{
			// regex type is not supported. This should not be added to the httpMatch headers.
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchRegularExpression),
			Name:  "ignore-this-header",
			Value: "val",
		},
		{
			Type:  helpers.GetHeaderMatchTypePointer(v1alpha2.HeaderMatchExact),
			Name:  "header-3",
			Value: "val-3",
		},
	}
	testQueryParamMatches := []v1alpha2.HTTPQueryParamMatch{
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
			Name:  "arg1",
			Value: "val1",
		},
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
			Name:  "arg2",
			Value: "val2=another-val",
		},
		{
			// regex type is not supported. This should not be added to the httpMatch args
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchRegularExpression),
			Name:  "ignore-this-arg",
			Value: "val",
		},
		{
			Type:  helpers.GetQueryParamMatchTypePointer(v1alpha2.QueryParamMatchExact),
			Name:  "arg3",
			Value: "==val3",
		},
	}

	expectedHeaders := []string{"header-1:val-1", "header-2:val-2", "header-3:val-3"}
	expectedArgs := []string{"arg1=val1", "arg2=val2=another-val", "arg3===val3"}
	tests := []struct {
		match    v1alpha2.HTTPRouteMatch
		expected httpMatch
		msg      string
	}{
		{
			match: v1alpha2.HTTPRouteMatch{
				Method: testMethodMatch,
			},
			expected: httpMatch{
				Method:       "PUT",
				RedirectPath: testPath,
			},
			msg: "method only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				RedirectPath: testPath,
				Headers:      expectedHeaders,
			},
			msg: "headers only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "query params only match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Method:      testMethodMatch,
				QueryParams: testQueryParamMatches,
			},
			expected: httpMatch{
				Method:       "PUT",
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "method and query params match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Method:  testMethodMatch,
				Headers: testHeaderMatches,
			},
			expected: httpMatch{
				Method:       "PUT",
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "method and headers match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				QueryParams: testQueryParamMatches,
				Headers:     testHeaderMatches,
			},
			expected: httpMatch{
				QueryParams:  expectedArgs,
				Headers:      expectedHeaders,
				RedirectPath: testPath,
			},
			msg: "query params and headers match",
		},
		{
			match: v1alpha2.HTTPRouteMatch{
				Headers:     testHeaderMatches,
				QueryParams: testQueryParamMatches,
				Method:      testMethodMatch,
			},
			expected: httpMatch{
				Method:       "PUT",
				Headers:      expectedHeaders,
				QueryParams:  expectedArgs,
				RedirectPath: testPath,
			},
			msg: "method, headers, and query params match",
		},
	}
	for _, tc := range tests {
		result := createHTTPMatch(tc.match, testPath)
		if diff := helpers.Diff(result, tc.expected); diff != "" {
			t.Errorf("createHTTPMatch() returned incorrect httpMatch for test case: %q, diff: %+v", tc.msg, diff)
		}
	}
}
