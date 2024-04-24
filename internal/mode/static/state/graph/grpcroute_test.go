package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation/validationfakes"
)

func createGRPCMethodMatch(serviceName, methodName, methodType string) v1alpha2.GRPCRouteRule {
	var mt *v1alpha2.GRPCMethodMatchType
	if methodType != "nilType" {
		mt = (*v1alpha2.GRPCMethodMatchType)(&methodType)
	}
	return v1alpha2.GRPCRouteRule{
		Matches: []v1alpha2.GRPCRouteMatch{
			{
				Method: &v1alpha2.GRPCMethodMatch{
					Type:    mt,
					Service: &serviceName,
					Method:  &methodName,
				},
			},
		},
	}
}

func createGRPCHeadersMatch(headerType, headerName, headerValue string) v1alpha2.GRPCRouteRule {
	return v1alpha2.GRPCRouteRule{
		Matches: []v1alpha2.GRPCRouteMatch{
			{
				Headers: []v1alpha2.GRPCHeaderMatch{
					{
						Type:  (*v1.HeaderMatchType)(&headerType),
						Name:  v1alpha2.GRPCHeaderName(headerName),
						Value: headerValue,
					},
				},
			},
		},
	}
}

func createGRPCRoute(
	name string,
	refName string,
	hostname v1.Hostname,
	rules []v1alpha2.GRPCRouteRule,
) *v1alpha2.GRPCRoute {
	return &v1alpha2.GRPCRoute{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      name,
		},
		Spec: v1alpha2.GRPCRouteSpec{
			CommonRouteSpec: v1.CommonRouteSpec{
				ParentRefs: []v1.ParentReference{
					{
						Namespace:   helpers.GetPointer[v1.Namespace]("test"),
						Name:        v1.ObjectName(refName),
						SectionName: helpers.GetPointer[v1.SectionName](v1.SectionName(sectionNameOfCreateHTTPRoute)),
					},
				},
			},
			Hostnames: []v1.Hostname{hostname},
			Rules:     rules,
		},
	}
}

func TestBuildGRPCRoutes(t *testing.T) {
	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	gr := createGRPCRoute("gr-1", gwNsName.Name, "example.com", []v1alpha2.GRPCRouteRule{})

	grWrongGateway := createGRPCRoute("gr-2", "some-gateway", "example.com", []v1alpha2.GRPCRouteRule{})

	grRoutes := map[types.NamespacedName]*v1alpha2.GRPCRoute{
		client.ObjectKeyFromObject(gr):             gr,
		client.ObjectKeyFromObject(grWrongGateway): grWrongGateway,
	}

	grRouteKey := RouteKey{
		NamespacedName: client.ObjectKeyFromObject(gr),
		RouteType:      RouteTypeGRPC,
	}

	tests := []struct {
		expected  map[RouteKey]*L7Route
		name      string
		gwNsNames []types.NamespacedName
	}{
		{
			gwNsNames: []types.NamespacedName{gwNsName},
			expected: map[RouteKey]*L7Route{
				grRouteKey: {
					RouteType:     RouteTypeGRPC,
					Source:        gr,
					SrcParentRefs: gr.Spec.ParentRefs,
					ParentRefs: []ParentRef{
						{
							Idx:     0,
							Gateway: gwNsName,
						},
					},
					Valid:      true,
					Attachable: true,
					Spec: L7RouteSpec{
						Hostnames: gr.Spec.Hostnames,
						Rules:     []RouteRule{},
					},
				},
			},
			name: "normal case",
		},
		{
			gwNsNames: []types.NamespacedName{},
			expected:  nil,
			name:      "no gateways",
		},
	}

	validator := &validationfakes.FakeHTTPFieldsValidator{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			routes := buildRoutesForGateways(validator, map[types.NamespacedName]*v1.HTTPRoute{}, grRoutes, test.gwNsNames)
			g.Expect(helpers.Diff(test.expected, routes)).To(BeEmpty())
		})
	}
}

func TestBuildGRPCRoute(t *testing.T) {
	gatewayNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}

	methodMatchRule := createGRPCMethodMatch("myService", "myMethod", "Exact")
	headersMatchRule := createGRPCHeadersMatch("Exact", "MyHeader", "SomeValue")

	methodMatchEmptyFields := createGRPCMethodMatch("", "", "")
	methodMatchNilType := createGRPCMethodMatch("myService", "myMethod", "nilType")
	headersMatchInvalid := createGRPCHeadersMatch("", "MyHeader", "SomeValue")

	grBoth := createGRPCRoute(
		"gr-1",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{methodMatchRule, headersMatchRule},
	)

	grInvalidHostname := createGRPCRoute("gr-1", gatewayNsName.Name, "", []v1alpha2.GRPCRouteRule{methodMatchRule})
	grNotNGF := createGRPCRoute("gr", "some-gateway", "example.com", []v1alpha2.GRPCRouteRule{methodMatchRule})

	grInvalidMatchesEmptyMethodFields := createGRPCRoute(
		"gr-1",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{methodMatchEmptyFields},
	)
	grInvalidMatchesNilMethodType := createGRPCRoute(
		"gr-1",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{methodMatchNilType},
	)
	grInvalidHeadersEmptyType := createGRPCRoute(
		"gr-1",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{headersMatchInvalid},
	)
	grOneInvalid := createGRPCRoute(
		"gr-1",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{methodMatchRule, headersMatchInvalid},
	)

	grDuplicateSectionName := createGRPCRoute(
		"gr",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{methodMatchRule},
	)
	grDuplicateSectionName.Spec.ParentRefs = append(
		grDuplicateSectionName.Spec.ParentRefs,
		grDuplicateSectionName.Spec.ParentRefs[0],
	)

	grInvalidFilterRule := createGRPCMethodMatch("myService", "myMethod", "Exact")

	grInvalidFilterRule.Filters = []v1alpha2.GRPCRouteFilter{
		{
			Type: "RequestHeaderModifier",
		},
	}

	grInvalidFilter := createGRPCRoute(
		"gr",
		gatewayNsName.Name,
		"example.com",
		[]v1alpha2.GRPCRouteRule{grInvalidFilterRule},
	)

	tests := []struct {
		validator *validationfakes.FakeHTTPFieldsValidator
		gr        *v1alpha2.GRPCRoute
		expected  *L7Route
		name      string
	}{
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grBoth,
			expected: &L7Route{
				RouteType:     RouteTypeGRPC,
				Source:        grBoth,
				SrcParentRefs: grBoth.Spec.ParentRefs,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Valid:      true,
				Attachable: true,
				Spec: L7RouteSpec{
					Hostnames: grBoth.Spec.Hostnames,
					Rules: []RouteRule{
						{
							ValidMatches:     true,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grBoth.Spec.Rules[0].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
						{
							ValidMatches:     true,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grBoth.Spec.Rules[1].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
					},
				},
			},
			name: "normal case with both",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grInvalidMatchesEmptyMethodFields,
			expected: &L7Route{
				RouteType:     RouteTypeGRPC,
				Source:        grInvalidMatchesEmptyMethodFields,
				SrcParentRefs: grInvalidMatchesEmptyMethodFields.Spec.ParentRefs,
				Valid:         false,
				Attachable:    true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: ` +
							`[spec.rules[0].matches[0].method.type: Unsupported value: "": supported values: "Exact",` +
							` spec.rules[0].matches[0].method.service: Required value: service is required,` +
							` spec.rules[0].matches[0].method.method: Required value: method is required]`,
					),
				},
				Spec: L7RouteSpec{
					Hostnames: grInvalidMatchesEmptyMethodFields.Spec.Hostnames,
					Rules: []RouteRule{
						{
							ValidMatches:     false,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grInvalidMatchesEmptyMethodFields.Spec.Rules[0].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
					},
				},
			},
			name: "invalid matches with empty method fields",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grDuplicateSectionName,
			expected: &L7Route{
				RouteType: RouteTypeGRPC,
				Source:    grDuplicateSectionName,
			},
			name: "invalid route with duplicate sectionName",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grOneInvalid,
			expected: &L7Route{
				Source:        grOneInvalid,
				SrcParentRefs: grOneInvalid.Spec.ParentRefs,
				RouteType:     RouteTypeGRPC,
				Valid:         true,
				Attachable:    true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRoutePartiallyInvalid(
						`spec.rules[1].matches[0].headers[0].type: Unsupported value: "": supported values: "Exact"`,
					),
				},
				Spec: L7RouteSpec{
					Hostnames: grOneInvalid.Spec.Hostnames,
					Rules: []RouteRule{
						{
							ValidMatches:     true,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grOneInvalid.Spec.Rules[0].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
						{
							ValidMatches:     false,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grOneInvalid.Spec.Rules[1].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
					},
				},
			},
			name: "invalid headers and valid method",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grInvalidHeadersEmptyType,
			expected: &L7Route{
				Source:        grInvalidHeadersEmptyType,
				SrcParentRefs: grInvalidHeadersEmptyType.Spec.ParentRefs,
				RouteType:     RouteTypeGRPC,
				Valid:         false,
				Attachable:    true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].matches[0].headers[0].type: ` +
							`Unsupported value: "": supported values: "Exact"`,
					),
				},
				Spec: L7RouteSpec{
					Hostnames: grInvalidHeadersEmptyType.Spec.Hostnames,
					Rules: []RouteRule{
						{
							ValidMatches:     false,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grInvalidHeadersEmptyType.Spec.Rules[0].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
					},
				},
			},
			name: "invalid headers with empty type",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grInvalidMatchesNilMethodType,
			expected: &L7Route{
				Source:        grInvalidMatchesNilMethodType,
				SrcParentRefs: grInvalidMatchesNilMethodType.Spec.ParentRefs,
				RouteType:     RouteTypeGRPC,
				Valid:         false,
				Attachable:    true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].matches[0].method.type: Required value: cannot be empty`,
					),
				},
				Spec: L7RouteSpec{
					Hostnames: grInvalidMatchesNilMethodType.Spec.Hostnames,
					Rules: []RouteRule{
						{
							ValidMatches:     false,
							ValidFilters:     true,
							Matches:          convertGRPCMatches(grInvalidMatchesNilMethodType.Spec.Rules[0].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
					},
				},
			},
			name: "invalid method with nil type",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grInvalidFilter,
			expected: &L7Route{
				Source:        grInvalidFilter,
				SrcParentRefs: grInvalidFilter.Spec.ParentRefs,
				RouteType:     RouteTypeGRPC,
				Valid:         false,
				Attachable:    true,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`All rules are invalid: spec.rules[0].filters: Unsupported value: []v1alpha2.GRPCRouteFilter{v1alpha2.` +
							`GRPCRouteFilter{Type:"RequestHeaderModifier", RequestHeaderModifier:(*v1.HTTPHeaderFilter)(nil), ` +
							`ResponseHeaderModifier:(*v1.HTTPHeaderFilter)(nil), RequestMirror:(*v1.HTTPRequestMirrorFilter)(nil), ` +
							`ExtensionRef:(*v1.LocalObjectReference)(nil)}}: supported values: "gRPC filters are not yet supported"`,
					),
				},
				Spec: L7RouteSpec{
					Hostnames: grInvalidFilter.Spec.Hostnames,
					Rules: []RouteRule{
						{
							ValidMatches:     true,
							ValidFilters:     false,
							Matches:          convertGRPCMatches(grInvalidFilter.Spec.Rules[0].Matches),
							RouteBackendRefs: []RouteBackendRef{},
						},
					},
				},
			},
			name: "invalid filter",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grNotNGF,
			expected:  nil,
			name:      "not NGF route",
		},
		{
			validator: &validationfakes.FakeHTTPFieldsValidator{},
			gr:        grInvalidHostname,
			expected: &L7Route{
				Source:        grInvalidHostname,
				RouteType:     RouteTypeGRPC,
				SrcParentRefs: grInvalidHostname.Spec.ParentRefs,
				Valid:         false,
				Attachable:    false,
				ParentRefs: []ParentRef{
					{
						Idx:     0,
						Gateway: gatewayNsName,
					},
				},
				Conditions: []conditions.Condition{
					staticConds.NewRouteUnsupportedValue(
						`spec.hostnames[0]: Invalid value: "": cannot be empty string`,
					),
				},
			},
			name: "invalid hostname",
		},
	}

	gatewayNsNames := []types.NamespacedName{gatewayNsName}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			route := buildGRPCRoute(test.validator, test.gr, gatewayNsNames)
			g.Expect(helpers.Diff(test.expected, route)).To(BeEmpty())
		})
	}
}
