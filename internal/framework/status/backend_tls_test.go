package status

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestPrepareBackendTLSPolicyStatus(t *testing.T) {
	oldStatus := v1alpha2.PolicyStatus{
		Ancestors: []v1alpha2.PolicyAncestorStatus{
			{
				AncestorRef: v1.ParentReference{
					Namespace: helpers.GetPointer((v1.Namespace)("ns1")),
					Name:      v1alpha2.ObjectName("other-gw"),
				},
				ControllerName: v1alpha2.GatewayController("otherCtlr"),
				Conditions:     []metav1.Condition{{Type: "otherType", Status: "otherStatus"}},
			},
		},
	}

	newStatus := BackendTLSPolicyStatus{
		AncestorStatuses: []AncestorStatus{
			{
				GatewayNsName: types.NamespacedName{
					Namespace: "ns1",
					Name:      "gw1",
				},
				Conditions: []conditions.Condition{{Type: "type1", Status: "status1"}},
			},
		},
		ObservedGeneration: 1,
	}

	transistionTime := metav1.Now()
	ctlrName := "nginx-gateway"

	policyStatus := prepareBackendTLSPolicyStatus(oldStatus, newStatus, ctlrName, transistionTime)

	g := NewWithT(t)

	g.Expect(policyStatus.Ancestors).To(HaveLen(2))
}
