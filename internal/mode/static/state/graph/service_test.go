package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestBuildReferencedServices(t *testing.T) {
	t.Parallel()

	gwNsname := types.NamespacedName{Namespace: "test", Name: "gwNsname"}
	gw := &Gateway{
		Source: &v1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: gwNsname.Namespace,
				Name:      gwNsname.Name,
			},
		},
	}
	ignoredGw := types.NamespacedName{Namespace: "test", Name: "ignoredGw"}

	getNormalL7Route := func() *L7Route {
		return &L7Route{
			ParentRefs: []ParentRef{
				{
					Gateway: gwNsname,
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
					Gateway: gwNsname,
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

	validRouteNoServiceNsName := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.Spec.Rules[0].BackendRefs[0].SvcNsName = types.NamespacedName{}
		return route
	})

	validL4RouteNoServiceNsName := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.Spec.BackendRef.SvcNsName = types.NamespacedName{}
		return route
	})

	normalL4RouteWinningAndIgnoredGws := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.ParentRefs = []ParentRef{
			{
				Gateway: ignoredGw,
			},
			{
				Gateway: ignoredGw,
			},
			{
				Gateway: gwNsname,
			},
		}
		return route
	})

	normalRouteWinningAndIgnoredGws := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.ParentRefs = []ParentRef{
			{
				Gateway: ignoredGw,
			},
			{
				Gateway: gwNsname,
			},
			{
				Gateway: ignoredGw,
			},
		}
		return route
	})

	normalL4RouteIgnoredGw := getModifiedL4Route(func(route *L4Route) *L4Route {
		route.ParentRefs[0].Gateway = ignoredGw
		return route
	})

	normalL7RouteIgnoredGw := getModifiedL7Route(func(route *L7Route) *L7Route {
		route.ParentRefs[0].Gateway = ignoredGw
		return route
	})

	tests := []struct {
		l7Routes map[RouteKey]*L7Route
		l4Routes map[L4RouteKey]*L4Route
		exp      map[types.NamespacedName]*ReferencedService
		gw       *Gateway
		name     string
	}{
		{
			name: "normal routes",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}: normalRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}: normalL4Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "banana-ns", Name: "service"}:   {},
				{Namespace: "tlsroute-ns", Name: "service"}: {},
			},
		},
		{
			name: "l7 route with two services in one Rule", // l4 routes don't support multiple services right now
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "two-svc-one-rule"}}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "route with one service per rule", // l4 routes don't support multiple rules right now
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "multiple valid routes with same services",
			gw:   gw,
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
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
				{Namespace: "tlsroute-ns", Name: "service"}:  {},
				{Namespace: "tlsroute-ns", Name: "service2"}: {},
			},
		},
		{
			name: "valid routes that do not belong to winning gateway",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "belongs-to-ignored-gws"}}: normalL7RouteIgnoredGw,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "belongs-to-ignored-gw"}}: normalL4RouteIgnoredGw,
			},
			exp: nil,
		},
		{
			name: "valid routes that belong to both winning and ignored gateways",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "belongs-to-ignored-gws"}}: normalRouteWinningAndIgnoredGws,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "ignored-gw"}}: normalL4RouteWinningAndIgnoredGws,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "banana-ns", Name: "service"}:   {},
				{Namespace: "tlsroute-ns", Name: "service"}: {},
			},
		},
		{
			name: "valid routes with different services",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:     normalRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}: normalL4Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
				{Namespace: "banana-ns", Name: "service"}:    {},
				{Namespace: "tlsroute-ns", Name: "service"}:  {},
			},
		},
		{
			name: "invalid routes",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-l4-route"}}: invalidL4Route,
			},
			exp: nil,
		},
		{
			name: "combination of valid and invalid routes",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:  normalRoute,
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-l4-route"}}: invalidL4Route,
				{NamespacedName: types.NamespacedName{Name: "normal-l4-route"}}:  normalL4Route,
			},
			exp: map[types.NamespacedName]*ReferencedService{
				{Namespace: "banana-ns", Name: "service"}:   {},
				{Namespace: "tlsroute-ns", Name: "service"}: {},
			},
		},
		{
			name: "valid route no service nsname",
			gw:   gw,
			l7Routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "no-service-nsname"}}: validRouteNoServiceNsName,
			},
			l4Routes: map[L4RouteKey]*L4Route{
				{NamespacedName: types.NamespacedName{Name: "no-service-nsname-l4"}}: validL4RouteNoServiceNsName,
			},
			exp: nil,
		},
		{
			name: "nil gateway",
			gw:   nil,
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

			g.Expect(buildReferencedServices(test.l7Routes, test.l4Routes, test.gw)).To(Equal(test.exp))
		})
	}
}
