package status

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareGatewayStatus(t *testing.T) {
	t.Setenv("POD_IP", "1.2.3.4")
	ipAddrType := v1beta1.IPAddressType
	podIP := v1beta1.GatewayAddress{
		Type:  &ipAddrType,
		Value: "1.2.3.4",
	}

	status := state.GatewayStatus{
		ListenerStatuses: state.ListenerStatuses{
			"listener": {
				AttachedRoutes: 3,
				Conditions:     CreateTestConditions(),
			},
		},
		ObservedGeneration: 1,
	}

	transitionTime := metav1.NewTime(time.Now())

	expected := v1beta1.GatewayStatus{
		Listeners: []v1beta1.ListenerStatus{
			{
				Name: "listener",
				SupportedKinds: []v1beta1.RouteGroupKind{
					{
						Kind: "HTTPRoute",
					},
				},
				AttachedRoutes: 3,
				Conditions:     CreateExpectedAPIConditions(1, transitionTime),
			},
		},
		Addresses: []v1beta1.GatewayAddress{podIP},
	}

	g := NewGomegaWithT(t)

	result := prepareGatewayStatus(status, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}

func TestPrepareIgnoredGatewayStatus(t *testing.T) {
	status := state.IgnoredGatewayStatus{
		ObservedGeneration: 1,
	}

	transitionTime := metav1.NewTime(time.Now())

	expected := v1beta1.GatewayStatus{
		Conditions: []metav1.Condition{
			{
				Type:               string(v1beta1.GatewayConditionReady),
				Status:             metav1.ConditionFalse,
				ObservedGeneration: status.ObservedGeneration,
				LastTransitionTime: transitionTime,
				Reason:             string(GetawayReasonGatewayConflict),
				Message:            GatewayMessageGatewayConflict,
			},
		},
	}

	result := prepareIgnoredGatewayStatus(status, transitionTime)
	if diff := cmp.Diff(expected, result); diff != "" {
		t.Errorf("prepareIgnoredGatewayStatus() mismatch (-want +got):\n%s", diff)
	}
}
