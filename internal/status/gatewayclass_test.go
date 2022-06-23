package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareGatewayClassStatus(t *testing.T) {
	transitionTime := metav1.NewTime(time.Now())

	tests := []struct {
		status   state.GatewayClassStatus
		expected v1alpha2.GatewayClassStatus
		msg      string
	}{
		{
			status: state.GatewayClassStatus{
				Valid:              true,
				ObservedGeneration: 1,
			},
			expected: v1alpha2.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(v1alpha2.GatewayClassConditionStatusAccepted),
						Status:             metav1.ConditionTrue,
						ObservedGeneration: 1,
						LastTransitionTime: transitionTime,
						Reason:             string(v1alpha2.GatewayClassReasonAccepted),
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
			expected: v1alpha2.GatewayClassStatus{
				Conditions: []metav1.Condition{
					{
						Type:               string(v1alpha2.GatewayClassConditionStatusAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: 2,
						LastTransitionTime: transitionTime,
						Reason:             string(v1alpha2.GatewayClassReasonAccepted),
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
