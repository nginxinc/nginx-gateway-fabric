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

	acceptedTrue := conditions.Condition{
		Type:   string(v1beta1.RouteConditionAccepted),
		Status: metav1.ConditionTrue,
	}
	acceptedFalse := conditions.Condition{
		Type:   string(v1beta1.RouteConditionAccepted),
		Status: metav1.ConditionFalse,
	}

	status := state.HTTPRouteStatus{
		ObservedGeneration: 1,
		ParentStatuses: map[string]state.ParentStatus{
			"accepted": {
				Conditions: []conditions.Condition{acceptedTrue},
			},
			"not-accepted": {
				Conditions: []conditions.Condition{acceptedFalse},
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
