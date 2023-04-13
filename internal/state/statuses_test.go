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
				Spec: v1beta1.HTTPRouteSpec{
					CommonRouteSpec: v1beta1.CommonRouteSpec{
						ParentRefs: []v1beta1.ParentReference{
							{
								SectionName: helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
							},
							{
								SectionName: helpers.GetPointer[v1beta1.SectionName]("listener-80-2"),
							},
						},
					},
				},
			},
			ParentRefs: []graph.ParentRef{
				{
					Idx:      0,
					Gateway:  client.ObjectKeyFromObject(gw),
					Attached: true,
				},
				{
					Idx:                       1,
					Gateway:                   client.ObjectKeyFromObject(gw),
					Attached:                  false,
					FailedAttachmentCondition: invalidCondition,
				},
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
					ObservedGeneration: 1,
					Conditions:         conditions.NewDefaultGatewayClassConditions(),
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
						ParentStatuses: []ParentStatus{
							{
								GatewayNsName: client.ObjectKeyFromObject(gw),
								SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
								Conditions:    conditions.NewDefaultRouteConditions(),
							},
							{
								GatewayNsName: client.ObjectKeyFromObject(gw),
								SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-2"),
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
						ParentStatuses: []ParentStatus{
							{
								GatewayNsName: client.ObjectKeyFromObject(gw),
								SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
								),
							},
							{
								GatewayNsName: client.ObjectKeyFromObject(gw),
								SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-2"),
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
					Valid: false,
					Conditions: []conditions.Condition{
						conditions.NewGatewayClassInvalidParameters("error"),
					},
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
					ObservedGeneration: 1,
					Conditions: []conditions.Condition{
						conditions.NewGatewayClassInvalidParameters("error"),
					},
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
						ParentStatuses: []ParentStatus{
							{
								GatewayNsName: client.ObjectKeyFromObject(gw),
								SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
								Conditions: append(
									conditions.NewDefaultRouteConditions(),
									conditions.NewTODO("GatewayClass is invalid or doesn't exist"),
								),
							},
							{
								GatewayNsName: client.ObjectKeyFromObject(gw),
								SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-2"),
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			result := buildStatuses(test.graph)
			g.Expect(helpers.Diff(test.expected, result)).To(BeEmpty())
		})
	}
}
