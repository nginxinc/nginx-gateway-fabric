package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
)

func TestPrepareGatewayStatus(t *testing.T) {
	statuses := newstate.ListenerStatuses{
		"valid-listener": {
			Valid:          true,
			AttachedRoutes: 2,
		},
		"invalid-listener": {
			Valid:          false,
			AttachedRoutes: 1,
		},
	}

	expectedTime := time.Now()
	clock := NewFakeClock(expectedTime)

	expected := v1alpha2.GatewayStatus{
		Listeners: []v1alpha2.ListenerStatus{
			{
				Name: "invalid-listener",
				SupportedKinds: []v1alpha2.RouteGroupKind{
					{
						Kind: "HTTPRoute",
					},
				},
				AttachedRoutes: 1,
				Conditions: []metav1.Condition{
					{
						Type:               string(v1alpha2.ListenerConditionReady),
						Status:             "False",
						ObservedGeneration: 123,
						LastTransitionTime: metav1.Time{Time: expectedTime},
						Reason:             string(v1alpha2.ListenerReasonInvalid),
					},
				},
			},
			{
				Name: "valid-listener",
				SupportedKinds: []v1alpha2.RouteGroupKind{
					{
						Kind: "HTTPRoute",
					},
				},
				AttachedRoutes: 2,
				Conditions: []metav1.Condition{
					{
						Type:               string(v1alpha2.ListenerConditionReady),
						Status:             "True",
						ObservedGeneration: 123,
						LastTransitionTime: metav1.Time{Time: expectedTime},
						Reason:             string(v1alpha2.ListenerReasonReady),
					},
				},
			},
		},
	}

	result := prepareGatewayStatus(statuses, clock)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("prepareGatewayStatus() mismatch (-want +got):\n%s", diff)
	}
}
