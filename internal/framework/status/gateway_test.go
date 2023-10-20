package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestPrepareGatewayStatus(t *testing.T) {
	podIP := v1.GatewayStatusAddress{
		Type:  helpers.GetPointer(v1.IPAddressType),
		Value: "1.2.3.4",
	}
	status := GatewayStatus{
		Conditions: CreateTestConditions("GatewayTest"),
		ListenerStatuses: ListenerStatuses{
			"listener": {
				AttachedRoutes: 3,
				Conditions:     CreateTestConditions("ListenerTest"),
				SupportedKinds: []v1.RouteGroupKind{
					{
						Kind: v1.Kind("HTTPRoute"),
					},
				},
			},
		},
		Addresses:          []v1.GatewayStatusAddress{podIP},
		ObservedGeneration: 1,
	}

	transitionTime := metav1.NewTime(time.Now())

	expected := v1.GatewayStatus{
		Conditions: CreateExpectedAPIConditions("GatewayTest", 1, transitionTime),
		Listeners: []v1.ListenerStatus{
			{
				Name: "listener",
				SupportedKinds: []v1.RouteGroupKind{
					{
						Kind: "HTTPRoute",
					},
				},
				AttachedRoutes: 3,
				Conditions:     CreateExpectedAPIConditions("ListenerTest", 1, transitionTime),
			},
		},
		Addresses: []v1.GatewayStatusAddress{podIP},
	}

	g := NewWithT(t)

	result := prepareGatewayStatus(status, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
