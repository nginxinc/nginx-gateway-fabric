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
	invalidRouteCondition := conditions.Condition{
		Type:   "TestInvalidRoute",
		Status: metav1.ConditionTrue,
	}
	invalidAttachmentCondition := conditions.Condition{
		Type:   "TestInvalidAttachment",
		Status: metav1.ConditionTrue,
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
		{Namespace: "test", Name: "hr-valid"}: {
			Valid: true,
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
					Idx:     0,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &graph.ParentRefAttachmentStatus{
						Attached: true,
					},
				},
				{
					Idx:     1,
					Gateway: client.ObjectKeyFromObject(gw),
					Attachment: &graph.ParentRefAttachmentStatus{
						Attached:        false,
						FailedCondition: invalidAttachmentCondition,
					},
				},
			},
		},
		{Namespace: "test", Name: "hr-invalid"}: {
			Valid:      false,
			Conditions: []conditions.Condition{invalidRouteCondition},
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
						},
					},
				},
			},
			ParentRefs: []graph.ParentRef{
				{
					Idx:        0,
					Gateway:    client.ObjectKeyFromObject(gw),
					Attachment: nil,
				},
			},
		},
	}

	graph := &graph.Graph{
		GatewayClass: &graph.GatewayClass{
			Source: &v1beta1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{Generation: 1},
			},
			Valid: true,
		},
		Gateway: &graph.Gateway{
			Source: gw,
			Listeners: map[string]*graph.Listener{
				"listener-80-1": {
					Valid: true,
					Routes: map[types.NamespacedName]*graph.Route{
						{Namespace: "test", Name: "hr-1"}: {},
					},
				},
			},
		},
		IgnoredGateways: map[types.NamespacedName]*v1beta1.Gateway{
			client.ObjectKeyFromObject(ignoredGw): ignoredGw,
		},
		Routes: routes,
	}

	expected := Statuses{
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
			{Namespace: "test", Name: "hr-valid"}: {
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
							invalidAttachmentCondition,
						),
					},
				},
			},
			{Namespace: "test", Name: "hr-invalid"}: {
				ObservedGeneration: 3,
				ParentStatuses: []ParentStatus{
					{
						GatewayNsName: client.ObjectKeyFromObject(gw),
						SectionName:   helpers.GetPointer[v1beta1.SectionName]("listener-80-1"),
						Conditions: append(
							conditions.NewDefaultRouteConditions(),
							invalidRouteCondition,
						),
					},
				},
			},
		},
	}

	g := NewGomegaWithT(t)

	result := buildStatuses(graph)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
