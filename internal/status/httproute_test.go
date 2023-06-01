package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestPrepareHTTPRouteStatus(t *testing.T) {
	gwNsName1 := types.NamespacedName{Namespace: "test", Name: "gateway-1"}
	gwNsName2 := types.NamespacedName{Namespace: "test", Name: "gateway-2"}

	status := HTTPRouteStatus{
		ObservedGeneration: 1,
		ParentStatuses: []ParentStatus{
			{
				GatewayNsName: gwNsName1,
				SectionName:   helpers.GetPointer[v1beta1.SectionName]("http"),
				Conditions:    CreateTestConditions("Test"),
			},
			{
				GatewayNsName: gwNsName2,
				SectionName:   nil,
				Conditions:    CreateTestConditions("Test"),
			},
		},
	}

	gatewayCtlrName := "test.example.com"
	transitionTime := metav1.NewTime(time.Now())

	expected := v1beta1.HTTPRouteStatus{
		RouteStatus: v1beta1.RouteStatus{
			Parents: []v1beta1.RouteParentStatus{
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   helpers.GetPointer(v1beta1.Namespace(gwNsName1.Namespace)),
						Name:        v1beta1.ObjectName(gwNsName1.Name),
						SectionName: helpers.GetPointer[v1beta1.SectionName]("http"),
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions:     CreateExpectedAPIConditions("Test", 1, transitionTime),
				},
				{
					ParentRef: v1beta1.ParentReference{
						Namespace:   helpers.GetPointer(v1beta1.Namespace(gwNsName2.Namespace)),
						Name:        v1beta1.ObjectName(gwNsName2.Name),
						SectionName: nil,
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions:     CreateExpectedAPIConditions("Test", 1, transitionTime),
				},
			},
		},
	}

	g := NewGomegaWithT(t)

	result := prepareHTTPRouteStatus(status, gatewayCtlrName, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
