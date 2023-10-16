package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestPrepareGatewayStatus(t *testing.T) {
	podIP := v1beta1.GatewayStatusAddress{
		Type:  helpers.GetPointer(v1beta1.IPAddressType),
		Value: "1.2.3.4",
	}
	status := GatewayStatus{
		Conditions: CreateTestConditions("GatewayTest"),
		ListenerStatuses: ListenerStatuses{
			"listener": {
				AttachedRoutes: 3,
				Conditions:     CreateTestConditions("ListenerTest"),
				SupportedKinds: []v1beta1.RouteGroupKind{
					{
						Kind: v1beta1.Kind("HTTPRoute"),
					},
				},
			},
		},
		Addresses:          []v1beta1.GatewayStatusAddress{podIP},
		ObservedGeneration: 1,
	}

	transitionTime := metav1.NewTime(time.Now())

	expected := v1beta1.GatewayStatus{
		Conditions: CreateExpectedAPIConditions("GatewayTest", 1, transitionTime),
		Listeners: []v1beta1.ListenerStatus{
			{
				Name: "listener",
				SupportedKinds: []v1beta1.RouteGroupKind{
					{
						Kind: "HTTPRoute",
					},
				},
				AttachedRoutes: 3,
				Conditions:     CreateExpectedAPIConditions("ListenerTest", 1, transitionTime),
			},
		},
		Addresses: []v1beta1.GatewayStatusAddress{podIP},
	}

	g := NewWithT(t)

	result := prepareGatewayStatus(status, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
