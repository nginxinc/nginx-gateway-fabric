package state

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestBuildStatuses(t *testing.T) {
	listeners := map[string]*listener{
		"listener-80-1": {
			Valid: true,
			Routes: map[types.NamespacedName]*route{
				{Namespace: "test", Name: "hr-1"}: {},
			},
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

	routesAllRefsInvalid := map[types.NamespacedName]*route{
		{Namespace: "test", Name: "hr-1"}: {
			InvalidSectionNameRefs: map[string]struct{}{
				"listener-80-2": {},
				"listener-80-1": {},
			},
		},
	}

	gw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "gateway",
		},
	}

	ignoredGw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  "test",
			Name:       "ignored-gateway",
			Generation: 1,
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
					Source: &v1beta1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid: true,
				},
				Gateway: &gateway{
					Source:    gw,
					Listeners: listeners,
				},
				IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{
					{Namespace: "test", Name: "ignored-gateway"}: ignoredGw,
				},
				Routes: routes,
			},
			expected: Statuses{
				GatewayClassStatus: &GatewayClassStatus{
					Valid:              true,
					ObservedGeneration: 1,
				},
				GatewayStatus: &GatewayStatus{
					NsName: types.NamespacedName{Namespace: "test", Name: "gateway"},
					ListenerStatuses: map[string]ListenerStatus{
						"listener-80-1": {
							Valid:          true,
							AttachedRoutes: 1,
						},
					},
				},
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{
					{Namespace: "test", Name: "ignored-gateway"}: {ObservedGeneration: 1},
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
				Gateway: &gateway{
					Source:    gw,
					Listeners: listeners,
				},
				IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{
					{Namespace: "test", Name: "ignored-gateway"}: ignoredGw,
				},
				Routes: routes,
			},
			expected: Statuses{
				GatewayClassStatus: nil,
				GatewayStatus: &GatewayStatus{
					NsName: types.NamespacedName{Namespace: "test", Name: "gateway"},
					ListenerStatuses: map[string]ListenerStatus{
						"listener-80-1": {
							Valid:          false,
							AttachedRoutes: 1,
						},
					},
				},
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{
					{Namespace: "test", Name: "ignored-gateway"}: {ObservedGeneration: 1},
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
					Source: &v1beta1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid:    false,
					ErrorMsg: "error",
				},
				Gateway: &gateway{
					Source:    gw,
					Listeners: listeners,
				},
				IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{
					{Namespace: "test", Name: "ignored-gateway"}: ignoredGw,
				},
				Routes: routes,
			},
			expected: Statuses{
				GatewayClassStatus: &GatewayClassStatus{
					Valid:              false,
					ErrorMsg:           "error",
					ObservedGeneration: 1,
				},
				GatewayStatus: &GatewayStatus{
					NsName: types.NamespacedName{Namespace: "test", Name: "gateway"},
					ListenerStatuses: map[string]ListenerStatus{
						"listener-80-1": {
							Valid:          false,
							AttachedRoutes: 1,
						},
					},
				},
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{
					{Namespace: "test", Name: "ignored-gateway"}: {ObservedGeneration: 1},
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
		{
			graph: &graph{
				GatewayClass: &gatewayClass{
					Source: &v1beta1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid: true,
				},
				Gateway:         nil,
				IgnoredGateways: nil,
				Routes:          routesAllRefsInvalid,
			},
			expected: Statuses{
				GatewayClassStatus: &GatewayClassStatus{
					Valid:              true,
					ObservedGeneration: 1,
				},
				GatewayStatus:          nil,
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{},
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
			msg: "gateway and ignored gateways don't exist",
		},
	}

	for _, test := range tests {
		result := buildStatuses(test.graph)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildStatuses() '%v' mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}
