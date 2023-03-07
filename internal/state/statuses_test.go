package state

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/graph"
)

func TestBuildStatuses(t *testing.T) {
	invalidCondition := conditions.Condition{
		Type:   "Test",
		Status: metav1.ConditionTrue,
	}

	listeners := map[string]*graph.Listener{
		"listener-80-1": {
			Valid: true,
			Routes: map[types.NamespacedName]*graph.Route{
				{Namespace: "test", Name: "hr-1"}: {},
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

	routes := map[types.NamespacedName]*graph.Route{
		{Namespace: "test", Name: "hr-1"}: {
			Source: &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 3,
				},
			},
			SectionNameRefs: map[string]graph.ParentRef{
				"listener-80-1": {
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
				},
				"listener-80-2": {
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(ignoredGw),
				},
			},
			UnboundSectionNameRefs: map[string]conditions.Condition{
				"listener-80-2": invalidCondition,
			},
		},
	}

	routesAllRefsInvalid := map[types.NamespacedName]*graph.Route{
		{Namespace: "test", Name: "hr-1"}: {
			Source: &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Generation: 4,
				},
			},
			SectionNameRefs: map[string]graph.ParentRef{
				"listener-80-1": {
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
				},
				"listener-80-2": {
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(ignoredGw),
				},
			},
			UnboundSectionNameRefs: map[string]conditions.Condition{
				"listener-80-1": invalidCondition,
				"listener-80-2": invalidCondition,
			},
		},
	}

	tests := []struct {
		graph    *graph.Graph
		expected Statuses
		name     string
	}{
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid: true,
				},
				Gateway: &graph.Gateway{
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
							AttachedRoutes: 1,
							Conditions:     conditions.NewDefaultListenerConditions(),
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
			name: "normal case",
		},
		{
			graph: &graph.Graph{
				GatewayClass: nil,
				Gateway: &graph.Gateway{
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
							AttachedRoutes: 1,
							Conditions: append(
								conditions.NewDefaultListenerConditions(),
								conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
							),
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
									conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
								),
							},
							"listener-80-2": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
									invalidCondition,
								),
							},
						},
					},
				},
			},
			name: "gatewayclass doesn't exist",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: &v1beta1.GatewayClass{
						ObjectMeta: metav1.ObjectMeta{Generation: 1},
					},
					Valid:    false,
					ErrorMsg: "error",
				},
				Gateway: &graph.Gateway{
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
							AttachedRoutes: 1,
							Conditions: append(
								conditions.NewDefaultListenerConditions(),
								conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
							),
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
									conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
								),
							},
							"listener-80-2": {
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
									invalidCondition,
								),
							},
						},
					},
				},
			},
			name: "gatewayclass is not valid",
		},
		{
			graph: &graph.Graph{
				GatewayClass: &graph.GatewayClass{
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
			name: "gateway and ignored gateways don't exist",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := buildStatuses(test.graph)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
