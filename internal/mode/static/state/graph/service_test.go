package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuildReferencedServices(t *testing.T) {
	t.Parallel()
	getNormalL7Route := func() *L7Route {
		return &L7Route{
			ParentRefs: []ParentRef{
				{
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
					},
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

	attachedRouteWithManyParentRefs := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.ParentRefs = []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
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
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
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
		exp      map[types.NamespacedName]struct{}
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
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}:   {},
				{Namespace: "tlsroute-ns", Name: "service"}: {},
			},
		},
		{
			name: "l7 route with two services in one Rule", // l4 routes don't support multiple services right now
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "two-svc-one-rule"}}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "route with one service per rule", // l4 routes don't support multiple rules right now
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
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
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
				{Namespace: "tlsroute-ns", Name: "service"}:  {},
				{Namespace: "tlsroute-ns", Name: "service2"}: {},
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
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
				{Namespace: "banana-ns", Name: "service"}:    {},
				{Namespace: "tlsroute-ns", Name: "service"}:  {},
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
			name: "combination of valid and invalid routes",
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:  normalRoute,
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-l4-route"}}: invalidL4Route,
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}:  normalL4Route,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}:   {},
				{Namespace: "tlsroute-ns", Name: "service"}: {},
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
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}:   {},
				{Namespace: "tlsroute-ns", Name: "service"}: {},
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
