package dataplane

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
)

func TestConvertMatch(t *testing.T) {
	path := v1beta1.HTTPPathMatch{
		Type:  helpers.GetPointer(v1beta1.PathMatchPathPrefix),
		Value: helpers.GetPointer("/"),
	}

	tests := []struct {
		match    v1beta1.HTTPRouteMatch
		name     string
		expected Match
	}{
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &path,
			},
			expected: Match{},
			name:     "path only",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path:   &path,
				Method: helpers.GetPointer(v1beta1.HTTPMethodGet),
			},
			expected: Match{
				Method: helpers.GetPointer("GET"),
			},
			name: "path and method",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &path,
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Name:  "Test-Header",
						Value: "test-header-value",
					},
				},
			},
			expected: Match{
				Headers: []HTTPHeaderMatch{
					{
						Name:  "Test-Header",
						Value: "test-header-value",
					},
				},
			},
			name: "path and header",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path: &path,
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Name:  "Test-Param",
						Value: "test-param-value",
					},
				},
			},
			expected: Match{
				QueryParams: []HTTPQueryParamMatch{
					{
						Name:  "Test-Param",
						Value: "test-param-value",
					},
				},
			},
			name: "path and query param",
		},
		{
			match: v1beta1.HTTPRouteMatch{
				Path:   &path,
				Method: helpers.GetPointer(v1beta1.HTTPMethodGet),
				Headers: []v1beta1.HTTPHeaderMatch{
					{
						Name:  "Test-Header",
						Value: "test-header-value",
					},
				},
				QueryParams: []v1beta1.HTTPQueryParamMatch{
					{
						Name:  "Test-Param",
						Value: "test-param-value",
					},
				},
			},
			expected: Match{
				Method: helpers.GetPointer("GET"),
				Headers: []HTTPHeaderMatch{
					{
						Name:  "Test-Header",
						Value: "test-header-value",
					},
				},
				QueryParams: []HTTPQueryParamMatch{
					{
						Name:  "Test-Param",
						Value: "test-param-value",
					},
				},
			},
			name: "path, method, header, and query param",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result := convertMatch(test.match)
			g.Expect(helpers.Diff(result, test.expected)).To(BeEmpty())
		})
	}
}

func TestConvertHTTPRequestRedirectFilter(t *testing.T) {
	tests := []struct {
		filter   *v1beta1.HTTPRequestRedirectFilter
		expected *HTTPRequestRedirectFilter
		name     string
	}{
		{
			filter:   &v1beta1.HTTPRequestRedirectFilter{},
			expected: &HTTPRequestRedirectFilter{},
			name:     "empty",
		},
		{
			filter: &v1beta1.HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer[v1beta1.PreciseHostname]("example.com"),
				Port:       helpers.GetPointer[v1beta1.PortNumber](8443),
				StatusCode: helpers.GetPointer(302),
			},
			expected: &HTTPRequestRedirectFilter{
				Scheme:     helpers.GetPointer("https"),
				Hostname:   helpers.GetPointer("example.com"),
				Port:       helpers.GetPointer[int32](8443),
				StatusCode: helpers.GetPointer(302),
			},
			name: "full",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result := convertHTTPRequestRedirectFilter(test.filter)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestConvertHTTPHeaderFilter(t *testing.T) {
	tests := []struct {
		filter   *v1beta1.HTTPHeaderFilter
		expected *HTTPHeaderFilter
		name     string
	}{
		{
			filter:   &v1beta1.HTTPHeaderFilter{},
			expected: &HTTPHeaderFilter{},
			name:     "empty",
		},
		{
			filter: &v1beta1.HTTPHeaderFilter{
				Set: []v1beta1.HTTPHeader{{
					Name:  "My-Set-Header",
					Value: "my-value",
				}},
				Add: []v1beta1.HTTPHeader{{
					Name:  "My-Add-Header",
					Value: "my-value",
				}},
				Remove: []string{"My-remove-header"},
			},
			expected: &HTTPHeaderFilter{
				Set: []HTTPHeader{{
					Name:  "My-Set-Header",
					Value: "my-value",
				}},
				Add: []HTTPHeader{{
					Name:  "My-Add-Header",
					Value: "my-value",
				}},
				Remove: []string{"My-remove-header"},
			},
			name: "full",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			result := convertHTTPHeaderFilter(test.filter)
			g.Expect(result).To(Equal(test.expected))
		})
	}
}

func TestConvertPathType(t *testing.T) {
	g := NewWithT(t)

	tests := []struct {
		pathType v1beta1.PathMatchType
		expected PathType
		panic    bool
	}{
		{
			expected: PathTypePrefix,
			pathType: v1beta1.PathMatchPathPrefix,
		},
		{
			expected: PathTypeExact,
			pathType: v1beta1.PathMatchExact,
		},
		{
			pathType: v1beta1.PathMatchRegularExpression,
			panic:    true,
		},
	}

	for _, tc := range tests {
		if tc.panic {
			g.Expect(func() { convertPathType(tc.pathType) }).To(Panic())
		} else {
			result := convertPathType(tc.pathType)
			g.Expect(result).To(Equal(tc.expected))
		}
	}
}
