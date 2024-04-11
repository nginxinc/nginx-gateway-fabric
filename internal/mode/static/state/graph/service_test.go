package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
)

func TestBuildReferencedServices(t *testing.T) {
	normalRoute := &HTTPRoute{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Valid: true,
		Rules: []Rule{
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
	}

	validRouteTwoServicesOneRule := &HTTPRoute{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Valid: true,
		Rules: []Rule{
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
	}

	validRouteTwoServicesTwoRules := &HTTPRoute{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Valid: true,
		Rules: []Rule{
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
	}

	invalidRoute := &HTTPRoute{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Valid: false,
		Rules: []Rule{
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
	}

	unattachedRoute := &HTTPRoute{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: false,
				},
			},
		},
		Valid: true,
		Rules: []Rule{
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
	}

	attachedRouteWithManyParentRefs := &HTTPRoute{
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
		Rules: []Rule{
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
	}
	validRouteNoServiceNsName := &HTTPRoute{
		ParentRefs: []ParentRef{
			{
				Attachment: &ParentRefAttachmentStatus{
					Attached: true,
				},
			},
		},
		Valid: true,
		Rules: []Rule{
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
	}

	tests := []struct {
		routes map[types.NamespacedName]*HTTPRoute
		exp    map[types.NamespacedName]struct{}
		name   string
	}{
		{
			name: "normal route",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "normal-route"}: normalRoute,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}: {},
			},
		},
		{
			name: "route with two services in one Rule",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "two-svc-one-rule"}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "route with one service per rule",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "one-svc-per-rule"}: validRouteTwoServicesTwoRules,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "two valid routes with same services",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "one-svc-per-rule"}: validRouteTwoServicesTwoRules,
				{Name: "two-svc-one-rule"}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
			},
		},
		{
			name: "two valid routes with different services",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "one-svc-per-rule"}: validRouteTwoServicesTwoRules,
				{Name: "normal-route"}:     normalRoute,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}:   {},
				{Namespace: "service-ns2", Name: "service2"}: {},
				{Namespace: "banana-ns", Name: "service"}:    {},
			},
		},
		{
			name: "invalid route",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "invalid-route"}: invalidRoute,
			},
			exp: nil,
		},
		{
			name: "unattached route",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "unattached-route"}: unattachedRoute,
			},
			exp: nil,
		},
		{
			name: "combination of valid and invalid routes",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "normal-route"}:  normalRoute,
				{Name: "invalid-route"}: invalidRoute,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "banana-ns", Name: "service"}: {},
			},
		},
		{
			name: "route with many parentRefs and one is attached",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "multiple-parent-ref-route"}: attachedRouteWithManyParentRefs,
			},
			exp: map[types.NamespacedName]struct{}{
				{Namespace: "service-ns", Name: "service"}: {},
			},
		},
		{
			name: "valid route no service nsname",
			routes: map[types.NamespacedName]*HTTPRoute{
				{Name: "no-service-nsname"}: validRouteNoServiceNsName,
			},
			exp: nil,
		},
	}

	grpcRoutes := map[types.NamespacedName]*GRPCRoute{}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(buildReferencedServices(test.routes, grpcRoutes)).To(Equal(test.exp))
		})
	}
}
