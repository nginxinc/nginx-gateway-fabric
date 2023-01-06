package state

import (
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func TestBuildStatuses(t *testing.T) {
	invalidCondition := conditions.RouteCondition{
		Type:   "Test",
		Status: metav1.ConditionTrue,
	}

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
			Source: &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 3,
				},
			},
			ValidSectionNameRefs: map[string]struct{}{
				"listener-80-1": {},
			},
			InvalidSectionNameRefs: map[string]conditions.RouteCondition{
				"listener-80-2": invalidCondition,
			},
		},
	}

	routesAllRefsInvalid := map[types.NamespacedName]*route{
		{Namespace: "test", Name: "hr-1"}: {
			Source: &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 4,
				},
			},
			InvalidSectionNameRefs: map[string]conditions.RouteCondition{
				"listener-80-2": invalidCondition,
				"listener-80-1": invalidCondition,
			},
		},
	}

	gw := &v1beta1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:  "test",
			Name:       "gateway",
			Generation: 2,
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
					ObservedGeneration: 2,
				},
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{
					{Namespace: "test", Name: "ignored-gateway"}: {ObservedGeneration: 1},
				},
				HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
					{Namespace: "test", Name: "hr-1"}: {
						ObservedGeneration: 3,
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Conditions: conditions.NewDefaultRouteConditions(),
							},
							"listener-80-2": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									invalidCondition,
								),
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
					ObservedGeneration: 2,
				},
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{
					{Namespace: "test", Name: "ignored-gateway"}: {ObservedGeneration: 1},
				},
				HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
					{Namespace: "test", Name: "hr-1"}: {
						ObservedGeneration: 3,
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewRouteTODO("GatewayClass is invalid or doesn't exist"),
								),
							},
							"listener-80-2": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewRouteTODO("GatewayClass is invalid or doesn't exist"),
									invalidCondition,
								),
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
					ObservedGeneration: 2,
				},
				IgnoredGatewayStatuses: map[types.NamespacedName]IgnoredGatewayStatus{
					{Namespace: "test", Name: "ignored-gateway"}: {ObservedGeneration: 1},
				},
				HTTPRouteStatuses: map[types.NamespacedName]HTTPRouteStatus{
					{Namespace: "test", Name: "hr-1"}: {
						ObservedGeneration: 3,
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewRouteTODO("GatewayClass is invalid or doesn't exist"),
								),
							},
							"listener-80-2": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewRouteTODO("GatewayClass is invalid or doesn't exist"),
									invalidCondition,
								),
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
						ObservedGeneration: 4,
						ParentStatuses: map[string]ParentStatus{
							"listener-80-1": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									invalidCondition,
								),
							},
							"listener-80-2": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									invalidCondition,
								),
							},
						},
					},
				},
			},
			msg: "gateway and ignored gateways don't exist",
		},
	}

	sortConditions := func(statuses Statuses) {
		for _, rs := range statuses.HTTPRouteStatuses {
			for _, ps := range rs.ParentStatuses {
				sort.Slice(ps.Conditions, func(i, j int) bool {
					return ps.Conditions[i].Type < ps.Conditions[j].Type
				})
			}
		}
	}

	for _, test := range tests {
		result := buildStatuses(test.graph)

		sortConditions(result)
		sortConditions(test.expected)

		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("buildStatuses() '%v' mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}
