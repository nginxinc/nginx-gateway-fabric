package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuildReferencedServices(t *testing.T) {
	t.Parallel()

	gw1 := types.NamespacedName{Namespace: "test", Name: "gw1"}
	gw2 := types.NamespacedName{Namespace: "test", Name: "gw2"}

	getNormalL7Route := func() *L7Route {
		return &L7Route{
			ParentRefs: []ParentRef{
				{
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
					},
					Gateway: gw1,
				},
			},
			Valid: true,
			Spec: L7RouteSpec{
				Rules: []RouteRule{
					{
						BackendRefs: []BackendRef{
							{
								SvcNsName: types.NamespacedName{Namespace: "banana-ns", Name: "service"},
							},
						},
					},
				},
			},
			RouteType: RouteTypeHTTP,
		}
	}

	getModifiedL7Route := func(mod func(route *L7Route) *L7Route) *L7Route {
		return mod(getNormalL7Route())
	}

	getNormalL4Route := func() *L4Route {
		return &L4Route{
			Spec: L4RouteSpec{
				BackendRef: BackendRef{
					SvcNsName: types.NamespacedName{Namespace: "tlsroute-ns", Name: "service"},
				},
			},
			Valid: true,
			ParentRefs: []ParentRef{
				{
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
					},
					Gateway: gw1,
				},
			},
		}
	}

	getModifiedL4Route := func(mod func(route *L4Route) *L4Route) *L4Route {
		return mod(getNormalL4Route())
	}

	normalRoute := getNormalL7Route()
	normalL4Route := getNormalL4Route()

	validRouteTwoServicesOneRule := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.Spec.Rules[0].BackendRefs = []BackendRef{
			{
				SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
			},
			{
				SvcNsName: types.NamespacedName{Namespace: "service-ns2", Name: "service2"},
			},
		}

		return route
	})

	validRouteTwoServicesTwoRules := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.Spec.Rules = []RouteRule{
			{
				BackendRefs: []BackendRef{
					{
						SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
					},
				},
			},
			{
				BackendRefs: []BackendRef{
					{
						SvcNsName: types.NamespacedName{Namespace: "service-ns2", Name: "service2"},
					},
				},
			},
		}

		return route
	})

	normalL4Route2 := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.Spec.BackendRef.SvcNsName = types.NamespacedName{Namespace: "tlsroute-ns", Name: "service2"}
		return route
	})

	normalL4RouteWithSameSvcAsL7Route := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.Spec.BackendRef.SvcNsName = types.NamespacedName{Namespace: "service-ns", Name: "service"}
		return route
	})

	invalidRoute := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.Valid = false
		return route
	})

	invalidL4Route := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.Valid = false
		return route
	})

	unattachedRoute := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.ParentRefs[0].Attachment.Attached = false
		return route
	})

	unattachedL4Route := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.ParentRefs[0].Attachment.Attached = false
		return route
	})

	nilAttachmentRoute := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.ParentRefs[0].Attachment = nil
		return route
	})

	nilAttachmentL4Route := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.ParentRefs[0].Attachment = nil
		return route
	})

	attachedRouteWithManyParentRefs := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.ParentRefs = []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
				Gateway: gw1,
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
				Gateway: gw2,
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
				Gateway: gw1,
			},
		}

		return route
	})

	attachedL4RoutesWithManyParentRefs := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.ParentRefs = []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
				Gateway: gw2,
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
				Gateway: gw1,
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
				Gateway: gw1,
			},
		}

		return route
	})

	validRouteNoServiceNsName := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.Spec.Rules[0].BackendRefs[0].SvcNsName = types.NamespacedName{}
		return route
	})

	validL4RouteNoServiceNsName := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.Spec.BackendRef.SvcNsName = types.NamespacedName{}
		return route
	})

	tests := []struct {
		l7Routes map[RouteKey]*L7Route
		l4Routes map[L4RouteKey]*L4Route
		exp      map[types.NamespacedName]*ReferencedService
		name     string
	}{
		{
			name: "normal routes",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}: normalRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}: normalL4Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "banana-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "tlsroute-ns", Name: "service"}: {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "l7 route with two services in one Rule", // l4 routes don't support multiple services right now
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "two-svc-one-rule"}}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "service-ns2", Name: "service2"}: {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "route with one service per rule", // l4 routes don't support multiple rules right now
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "service-ns2", Name: "service2"}: {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "multiple valid routes with same services",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
				{NamespacedName: types.NamespacedName{Name: "two-svc-one-rule"}}: validRouteTwoServicesOneRule,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "l4-route-1"}}:                    normalL4Route,
				{NamespacedName: types.NamespacedName{Name: "l4-route-2"}}:                    normalL4Route2,
				{NamespacedName: types.NamespacedName{Name: "l4-route-same-svc-as-l7-route"}}: normalL4RouteWithSameSvcAsL7Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "service-ns2", Name: "service2"}: {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "tlsroute-ns", Name: "service"}:  {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "tlsroute-ns", Name: "service2"}: {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "valid routes with different services",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:     normalRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}: normalL4Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "service-ns2", Name: "service2"}: {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "banana-ns", Name: "service"}:    {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "tlsroute-ns", Name: "service"}:  {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "invalid routes",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-l4-route"}}: invalidL4Route,
			},
			exp: nil,
		},
		{
			name: "unattached route",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "unattached-route"}}: unattachedRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "unattached-l4-route"}}: unattachedL4Route,
			},
			exp: nil,
		},
		{
			name: "route with nil parent attachment status",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "nil-attachment-route"}}: nilAttachmentRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "nil-attachment-l4-route"}}: nilAttachmentL4Route,
			},
			exp: nil,
		},
		{
			name: "combination of valid and invalid routes",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:  normalRoute,
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-l4-route"}}: invalidL4Route,
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}:  normalL4Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "banana-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "tlsroute-ns", Name: "service"}: {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "route with many parentRefs and one is attached",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "multiple-parent-ref-route"}}: attachedRouteWithManyParentRefs,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "multiple-parent-ref-l4-route"}}: attachedL4RoutesWithManyParentRefs,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "banana-ns", Name: "service"}:   {ParentGateways: []types.NamespacedName{gw1}},
				{Namespace: "tlsroute-ns", Name: "service"}: {ParentGateways: []types.NamespacedName{gw1}},
			},
		},
		{
			name: "valid route no service nsname",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "no-service-nsname"}}: validRouteNoServiceNsName,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "no-service-nsname-l4"}}: validL4RouteNoServiceNsName,
			},
			exp: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)
			g.Expect(buildReferencedServices(test.l7Routes, test.l4Routes)).To(Equal(test.exp))
		})
	}
}

func TestGetUniqueAttachedParentGateways(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	parentRefs := []ParentRef{
		{
			Attachment: &ParentRefAttachmentStatus{
				Attached: true,
			},
			Gateway: types.NamespacedName{Name: "attached-1", Namespace: "test"},
		},
		{
			Attachment: &ParentRefAttachmentStatus{
				Attached: true,
			},
			Gateway: types.NamespacedName{Name: "attached-2", Namespace: "test2"},
		},
		{
			Attachment: &ParentRefAttachmentStatus{
				Attached: false,
			},
			Gateway: types.NamespacedName{Name: "not-attached-1", Namespace: "test"},
		},
		{
			Attachment: nil,
			Gateway:    types.NamespacedName{Name: "nil-attachment", Namespace: "test"},
		},
		{
			Attachment: &ParentRefAttachmentStatus{
				Attached: true,
			},
			Gateway: types.NamespacedName{Name: "attached-1", Namespace: "test"}, // dupe
		},
	}

	expectedNsNames := []types.NamespacedName{
		{Namespace: "test", Name: "attached-1"},
		{Namespace: "test2", Name: "attached-2"},
	}

	uniqueAttachedRefs := getUniqueAttachedParentGateways(parentRefs)
	g.Expect(uniqueAttachedRefs).To(Equal(expectedNsNames))
}
