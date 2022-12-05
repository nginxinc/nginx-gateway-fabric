package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func TestPrepareHTTPRouteStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	acceptedTrue := conditions.RouteCondition{
		Type:   v1beta1.RouteConditionAccepted,
		Status: metav1.ConditionTrue,
	}
	acceptedFalse := conditions.RouteCondition{
		Type:   v1beta1.RouteConditionAccepted,
		Status: metav1.ConditionFalse,
	}

	status := state.HTTPRouteStatus{
		ObservedGeneration: 1,
		ParentStatuses: map[string]state.ParentStatus{
			"accepted": {
				Conditions: []conditions.RouteCondition{acceptedTrue},
			},
			"not-accepted": {
				Conditions: []conditions.RouteCondition{acceptedFalse},
			},
		},
	}

	gwNsName := types.NamespacedName{Namespace: "test", Name: "gateway"}
	gatewayCtlrName := "test.example.com"

	transitionTime := metav1.NewTime(time.Now())

	expected := v1beta1.HTTPRouteStatus{
		RouteStatus: v1beta1.RouteStatus{
			Parents: []v1beta1.RouteParentStatus{
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
						Name:        "gateway",
						SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("accepted")),
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions: []metav1.Condition{
						{
							Type:               string(v1beta1.RouteConditionAccepted),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
						},
					},
				},
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
						Name:        "gateway",
						SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("not-accepted")),
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions: []metav1.Condition{
						{
							Type:               string(v1beta1.RouteConditionAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 1,
							LastTransitionTime: transitionTime,
						},
					},
				},
			},
		},
	}
	result := prepareHTTPRouteStatus(status, gwNsName, gatewayCtlrName, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}

func TestConvertRouteConditions(t *testing.T) {
	g := NewGomegaWithT(t)

	routeConds := []conditions.RouteCondition{
		{
			Type:    "Test",
			Status:  metav1.ConditionTrue,
			Reason:  "reason1",
			Message: "message1",
		},
		{
			Type:    "Test",
			Status:  metav1.ConditionFalse,
			Reason:  "reason2",
			Message: "message2",
		},
	}

	var generation int64 = 1
	transitionTime := metav1.NewTime(time.Now())

	expected := []metav1.Condition{
		{
			Type:               "Test",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: generation,
			LastTransitionTime: transitionTime,
			Reason:             "reason1",
			Message:            "message1",
		},
		{
			Type:               "Test",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			LastTransitionTime: transitionTime,
			Reason:             "reason2",
			Message:            "message2",
		},
	}

	result := convertRouteConditions(routeConds, generation, transitionTime)

	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
