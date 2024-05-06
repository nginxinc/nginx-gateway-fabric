package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuildReferencedServices(t *testing.T) {
	normalRoute := &L7Route{
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
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
		RouteType: RouteTypeHTTP,
	}

	validRouteTwoServicesOneRule := &L7Route{
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
							SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
							Weight:    1,
						},
						{
							SvcNsName: types.NamespacedName{Namespace: "service-ns2", Name: "service2"},
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
	}

	validRouteTwoServicesTwoRules := &L7Route{
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
							SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
				{
					BackendRefs: []BackendRef{
						{
							SvcNsName: types.NamespacedName{Namespace: "service-ns2", Name: "service2"},
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
	}

	invalidRoute := &L7Route{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Valid: false,
		Spec: L7RouteSpec{
			Rules: []RouteRule{
				{
					BackendRefs: []BackendRef{
						{
							SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
	}

	unattachedRoute := &L7Route{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
			},
		},
		Valid: true,
		Spec: L7RouteSpec{
			Rules: []RouteRule{
				{
					BackendRefs: []BackendRef{
						{
							SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
	}

	attachedRouteWithManyParentRefs := &L7Route{
		ParentRefs: []ParentRef{
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
		},
		Valid: true,
		Spec: L7RouteSpec{
			Rules: []RouteRule{
				{
					BackendRefs: []BackendRef{
						{
							SvcNsName: types.NamespacedName{Namespace: "service-ns", Name: "service"},
							Weight:    1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
	}
	validRouteNoServiceNsName := &L7Route{
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
							Weight: 1,
						},
					},
					ValidMatches: true,
					ValidFilters: true,
				},
			},
		},
	}

	tests := []struct {
		routes map[RouteKey]*L7Route
		exp    map[types.NamespacedName]struct{}
		name   string
	}{
		{
			name: "normal route",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}: normalRoute,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}: {},
			},
		},
		{
			name: "route with two services in one Rule",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "two-svc-one-rule"}}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "route with one service per rule",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "two valid routes with same services",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
				{NamespacedName: types.NamespacedName{Name: "two-svc-one-rule"}}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "two valid routes with different services",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "one-svc-per-rule"}}: validRouteTwoServicesTwoRules,
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:     normalRoute,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
				{Namespace: "banana-ns", Name: "service"}:    {},
			},
		},
		{
			name: "invalid routes",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			exp: nil,
		},
		{
			name: "unattached route",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "unattached-route"}}: unattachedRoute,
			},
			exp: nil,
		},
		{
			name: "combination of valid and invalid routes",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "normal-route"}}:  normalRoute,
				{NamespacedName: types.NamespacedName{Name: "invalid-route"}}: invalidRoute,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}: {},
			},
		},
		{
			name: "route with many parentRefs and one is attached",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "multiple-parent-ref-route"}}: attachedRouteWithManyParentRefs,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}: {},
			},
		},
		{
			name: "valid route no service nsname",
			routes: map[RouteKey]*L7Route{
				{NamespacedName: types.NamespacedName{Name: "no-service-nsname"}}: validRouteNoServiceNsName,
			},
			exp: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(buildReferencedServices(test.routes)).To(Equal(test.exp))
		})
	}
}
