package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/types"
)

func TestBuildReferencedServices(t *testing.T) {
	normalRoute := &Route{
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

	validRouteTwoServicesOneRule := &Route{
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

	validRouteTwoServicesTwoRules := &Route{
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

	invalidRoute := &Route{
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

	unattachedRoute := &Route{
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

	attachedRouteWithManyParentRefs := &Route{
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
	validRouteNoServiceNsName := &Route{
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

	services := map[types.NamespacedName]*v1.Service{
		{Namespace: "banana-ns", Name: "service"}:    {ObjectMeta: metav1.ObjectMeta{Name: "banana-service"}},
		{Namespace: "service-ns", Name: "service"}:   {ObjectMeta: metav1.ObjectMeta{Name: "service"}},
		{Namespace: "service-ns2", Name: "service2"}: {ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
	}

	tests := []struct {
		routes map[types.NamespacedName]*Route
		exp    map[types.NamespacedName]*v1.Service
		name   string
	}{
		{
			name: "normal route",
			routes: map[types.NamespacedName]*Route{
				{Name: "normal-route"}: normalRoute,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "banana-ns", Name: "service"}: {ObjectMeta: metav1.ObjectMeta{Name: "banana-service"}},
			},
		},
		{
			name: "route with two services in one Rule",
			routes: map[types.NamespacedName]*Route{
				{Name: "two-svc-one-rule"}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "service-ns", Name: "service"}:   {ObjectMeta: metav1.ObjectMeta{Name: "service"}},
				{Namespace: "service-ns2", Name: "service2"}: {ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
			},
		},
		{
			name: "route with one service per rule",
			routes: map[types.NamespacedName]*Route{
				{Name: "one-svc-per-rule"}: validRouteTwoServicesTwoRules,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "service-ns", Name: "service"}:   {ObjectMeta: metav1.ObjectMeta{Name: "service"}},
				{Namespace: "service-ns2", Name: "service2"}: {ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
			},
		},
		{
			name: "two valid routes with same services",
			routes: map[types.NamespacedName]*Route{
				{Name: "one-svc-per-rule"}: validRouteTwoServicesTwoRules,
				{Name: "two-svc-one-rule"}: validRouteTwoServicesOneRule,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "service-ns", Name: "service"}:   {ObjectMeta: metav1.ObjectMeta{Name: "service"}},
				{Namespace: "service-ns2", Name: "service2"}: {ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
			},
		},
		{
			name: "two valid routes with different services",
			routes: map[types.NamespacedName]*Route{
				{Name: "one-svc-per-rule"}: validRouteTwoServicesTwoRules,
				{Name: "normal-route"}:     normalRoute,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "service-ns", Name: "service"}:   {ObjectMeta: metav1.ObjectMeta{Name: "service"}},
				{Namespace: "service-ns2", Name: "service2"}: {ObjectMeta: metav1.ObjectMeta{Name: "service2"}},
				{Namespace: "banana-ns", Name: "service"}:    {ObjectMeta: metav1.ObjectMeta{Name: "banana-service"}},
			},
		},
		{
			name: "invalid route",
			routes: map[types.NamespacedName]*Route{
				{Name: "invalid-route"}: invalidRoute,
			},
			exp: nil,
		},
		{
			name: "unattached route",
			routes: map[types.NamespacedName]*Route{
				{Name: "unattached-route"}: unattachedRoute,
			},
			exp: nil,
		},
		{
			name: "combination of valid and invalid routes",
			routes: map[types.NamespacedName]*Route{
				{Name: "normal-route"}:  normalRoute,
				{Name: "invalid-route"}: invalidRoute,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "banana-ns", Name: "service"}: {ObjectMeta: metav1.ObjectMeta{Name: "banana-service"}},
			},
		},
		{
			name: "route with many parentRefs and one is attached",
			routes: map[types.NamespacedName]*Route{
				{Name: "multiple-parent-ref-route"}: attachedRouteWithManyParentRefs,
			},
			exp: map[types.NamespacedName]*v1.Service{
				{Namespace: "service-ns", Name: "service"}: {ObjectMeta: metav1.ObjectMeta{Name: "service"}},
			},
		},
		{
			name: "valid route no service nsname",
			routes: map[types.NamespacedName]*Route{
				{Name: "no-service-nsname"}: validRouteNoServiceNsName,
			},
			exp: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(buildReferencedServices(test.routes, services)).To(Equal(test.exp))
		})
	}
}
