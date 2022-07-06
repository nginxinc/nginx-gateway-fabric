package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestBuildStatuses(t *testing.T) {
	listeners := map[string]*listener{
		"listener-80-1": {
			Valid: true,
			Routes: map[types.NamespacedName]*route{
				{Namespace: "test", Name: "hr-1"}: {}},
		},
	}

	routes := map[types.NamespacedName]*route{
		{Namespace: "test", Name: "hr-1"}: {
			ValidSectionNameRefs: map[string]struct{}{
				"listener-80-1": {},
			},
			InvalidSectionNameRefs: map[string]struct{}{
				"listener-80-2": {},
			},
		},
	}

	tests := []struct {
		graph    *graph
		expected Statuses
		msg      string
	}{
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1alpha2.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid: true,
				},
				Listeners: listeners,
				Routes:    routes,
			},
			expected: Statuses{
				GatewayClassStatus: &GatewayClassStatus{
					Valid:              true,
					ObservedGeneration: 1,
				},
				ListenerStatuses: map[string]ListenerStatus{
					"listener-80-1": {
						Valid:          true,
						AttachedRoutes: 1,
					},
				},
				HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
					{Namespace: "test", Name: "hr-1"}: {
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Attached: true,
							},
							"listener-80-2": {
								Attached: false,
							},
						},
					},
				},
			},
			msg: "normal case",
		},
		{
			graph: &graph{
				GatewayClass: nil,
				Listeners:    listeners,
				Routes:       routes,
			},
			expected: Statuses{
				ListenerStatuses: map[string]ListenerStatus{
					"listener-80-1": {
						Valid:          false,
						AttachedRoutes: 1,
					},
				},
				HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
					{Namespace: "test", Name: "hr-1"}: {
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Attached: false,
							},
							"listener-80-2": {
								Attached: false,
							},
						},
					},
				},
			},
			msg: "gatewayclass doesn't exist",
		},
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1alpha2.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid:    false,
					ErrorMsg: "error",
				},
				Listeners: listeners,
				Routes:    routes,
			},
			expected: Statuses{
				GatewayClassStatus: &GatewayClassStatus{
					Valid:              false,
					ErrorMsg:           "error",
					ObservedGeneration: 1,
				},
				ListenerStatuses: map[string]ListenerStatus{
					"listener-80-1": {
						Valid:          false,
						AttachedRoutes: 1,
					},
				},
				HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
					{Namespace: "test", Name: "hr-1"}: {
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Attached: false,
							},
							"listener-80-2": {
								Attached: false,
							},
						},
					},
				},
			},
			msg: "gatewayclass is not valid",
		},
	}

	for _, test := range tests {
		result := buildStatuses(test.graph)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildStatuses() '%v' mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}
