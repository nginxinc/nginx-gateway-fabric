package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareHTTPRouteStatus(t *testing.T) {
	g := NewGomegaWithT(t)

	status := state.HTTPRouteStatus{
		ObservedGeneration: 1,
		ParentStatuses: map[string]state.ParentStatus{
			"parent": {
				Conditions: CreateTestConditions(),
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
						SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("parent")),
					},
					ControllerName: v1beta1.GatewayController(gatewayCtlrName),
					Conditions:     CreateExpectedAPIConditions(1, transitionTime),
				},
			},
		},
	}
	result := prepareHTTPRouteStatus(status, gwNsName, gatewayCtlrName, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
