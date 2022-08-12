package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareHTTPRouteStatus(t *testing.T) {
	status := state.HTTPRouteStatus{
		ParentStatuses: map[string]state.ParentStatus{
			"attached": {
				Attached: true,
			},
			"not-attached": {
				Attached: false,
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
						SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("attached")),
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions: []metav1.Condition{
						{
							Type:               string(v1beta1.RouteConditionAccepted),
							Status:             metav1.ConditionTrue,
							ObservedGeneration: 123,
							LastTransitionTime: transitionTime,
							Reason:             "Accepted",
						},
					},
				},
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
						Name:        "gateway",
						SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("not-attached")),
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions: []metav1.Condition{
						{
							Type:               string(v1beta1.RouteConditionAccepted),
							Status:             metav1.ConditionFalse,
							ObservedGeneration: 123,
							LastTransitionTime: transitionTime,
							Reason:             "NotAttached",
						},
					},
				},
			},
		},
	}

	result := prepareHTTPRouteStatus(status, gwNsName, gatewayCtlrName, transitionTime)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("prepareHTTPRouteStatus() mismatch (-want +got):\n%s", diff)
	}
}
