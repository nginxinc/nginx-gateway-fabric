package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

func TestPrepareGatewayClassStatus(t *testing.T) {
	transitionTime := metav1.NewTime(time.Now())

	status := state.GatewayClassStatus{
		ObservedGeneration: 1,
		Conditions:         CreateTestConditions(),
	}
	expected := v1beta1.GatewayClassStatus{
		Conditions: CreateExpectedAPIConditions(1, transitionTime),
	}

	g := NewGomegaWithT(t)

	result := prepareGatewayClassStatus(status, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
