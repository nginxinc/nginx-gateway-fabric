package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
)

func TestPrepareHTTPRouteStatus(t *testing.T) {
	status := newstate.HTTPRouteStatus{
		ParentStatuses: map[string]newstate.ParentStatus{
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

	expectedTime := time.Now()
	clock := NewFakeClock(expectedTime)

	expected := v1alpha2.HTTPRouteStatus{
		RouteStatus: v1alpha2.RouteStatus{
			Parents: []v1alpha2.RouteParentStatus{
				{
					ParentRef: v1alpha2.ParentRef{
						Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
						Name:        "gateway",
						SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("attached")),
					},
					ControllerName: v1alpha2.GatewayController(gatewayCtlrName),
					Conditions: []metav1.Condition{
						{
							Type:               string(v1alpha2.ConditionRouteAccepted),
							Status:             "True",
							ObservedGeneration: 123,
							LastTransitionTime: metav1.Time{Time: expectedTime},
							Reason:             "Attached",
						},
					},
				},
				{
					ParentRef: v1alpha2.ParentRef{
						Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
						Name:        "gateway",
						SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("not-attached")),
					},
					ControllerName: v1alpha2.GatewayController(gatewayCtlrName),
					Conditions: []metav1.Condition{
						{
							Type:               string(v1alpha2.ConditionRouteAccepted),
							Status:             "False",
							ObservedGeneration: 123,
							LastTransitionTime: metav1.Time{Time: expectedTime},
							Reason:             "Not attached",
						},
					},
				},
			},
		},
	}

	result := prepareHTTPRouteStatus(status, gwNsName, gatewayCtlrName, clock)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("prepareHTTPRouteStatus() mismatch (-want +got):\n%s", diff)
	}
}
