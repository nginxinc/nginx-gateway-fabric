package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestPrepareGatewayClassStatus(t *testing.T) {
	transitionTime := metav1.NewTime(time.Now())

	status := GatewayClassStatus{
		ObservedGeneration: 1,
		Conditions:         CreateTestConditions("Test"),
	}
	expected := v1.GatewayClassStatus{
		Conditions: CreateExpectedAPIConditions("Test", 1, transitionTime),
	}

	g := NewWithT(t)

	result := prepareGatewayClassStatus(status, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
