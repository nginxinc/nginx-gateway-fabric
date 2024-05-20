package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestAncestorsFull(t *testing.T) {
	createCurStatus := func(numAncestors int, ctlrName string) []v1alpha2.PolicyAncestorStatus {
		statuses := make([]v1alpha2.PolicyAncestorStatus, 0, numAncestors)

		for i := 0; i < numAncestors; i++ {
			statuses = append(statuses, v1alpha2.PolicyAncestorStatus{
				ControllerName: v1.GatewayController(ctlrName),
			})
		}

		return statuses
	}

	tests := []struct {
		name      string
		curStatus []v1alpha2.PolicyAncestorStatus
		expFull   bool
	}{
		{
			name:      "not full",
			curStatus: createCurStatus(15, "controller"),
			expFull:   false,
		},
		{
			name:      "full; ancestor does not exist in current status",
			curStatus: createCurStatus(16, "controller"),
			expFull:   true,
		},
		{
			name:      "full, but ancestor does exist in current status",
			curStatus: createCurStatus(16, "nginx-gateway"),
			expFull:   false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			full := ancestorsFull(test.curStatus, "nginx-gateway")
			g.Expect(full).To(Equal(test.expFull))
		})
	}
}
