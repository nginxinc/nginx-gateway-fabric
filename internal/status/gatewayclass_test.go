package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareGatewayClassStatus(t *testing.T) {
	transitionTime := metav1.NewTime(time.Now())

	tests := []struct {
		msg      string
		expected v1beta1.GatewayClassStatus
		status   state.GatewayClassStatus
	}{
		{
			status: state.GatewayClassStatus{
				Valid:              true,
				ObservedGeneration: 1,
			},
			expected: v1beta1.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(v1beta1.GatewayClassConditionStatusAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 1,
						LastTransitionTime: transitionTime,
						Reason:             string(v1beta1.GatewayClassReasonAccepted),
						Message:            "GatewayClass has been accepted",
					},
				},
			},
			msg: "valid GatewayClass",
		},
		{
			status: state.GatewayClassStatus{
				Valid:              false,
				ErrorMsg:           "error",
				ObservedGeneration: 2,
			},
			expected: v1beta1.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(v1beta1.GatewayClassConditionStatusAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: 2,
						LastTransitionTime: transitionTime,
						Reason:             string(v1beta1.GatewayClassReasonAccepted),
						Message:            "GatewayClass has been rejected: error",
					},
				},
			},
			msg: "invalid GatewayClass",
		},
	}

	for _, test := range tests {
		result := prepareGatewayClassStatus(test.status, transitionTime)
		if diff := cmp.Diff(test.expected, result); diff != "" {
			t.Errorf("prepareGatewayClassStatus() '%s' mismatch (-want +got):\n%s", test.msg, diff)
		}
	}
}
