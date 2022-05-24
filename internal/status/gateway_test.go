package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareGatewayStatus(t *testing.T) {
	statuses := state.ListenerStatuses{
		"valid-listener": {
			Valid:          true,
			AttachedRoutes: 2,
		},
		"invalid-listener": {
			Valid:          false,
			AttachedRoutes: 1,
		},
	}

	transitionTime := metav1.NewTime(time.Now())

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
						Status:             metav1.ConditionFalse,
						ObservedGeneration: 123,
						LastTransitionTime: transitionTime,
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
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 123,
						LastTransitionTime: transitionTime,
						Reason:             string(v1alpha2.ListenerReasonReady),
					},
				},
			},
		},
	}

	result := prepareGatewayStatus(statuses, transitionTime)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("prepareGatewayStatus() mismatch (-want +got):\n%s", diff)
	}
}
