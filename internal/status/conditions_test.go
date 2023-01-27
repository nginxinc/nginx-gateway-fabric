package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func TestConvertRouteConditions(t *testing.T) {
	g := NewGomegaWithT(t)

	routeConds := []conditions.Condition{
		{
			Type:    "Test",
			Status:  metav1.ConditionTrue,
			Reason:  "reason1",
			Message: "message1",
		},
		{
			Type:    "Test",
			Status:  metav1.ConditionFalse,
			Reason:  "reason2",
			Message: "message2",
		},
	}

	var generation int64 = 1
	transitionTime := metav1.NewTime(time.Now())

	expected := []metav1.Condition{
		{
			Type:               "Test",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: generation,
			LastTransitionTime: transitionTime,
			Reason:             "reason1",
			Message:            "message1",
		},
		{
			Type:               "Test",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: generation,
			LastTransitionTime: transitionTime,
			Reason:             "reason2",
			Message:            "message2",
		},
	}

	result := convertConditions(routeConds, generation, transitionTime)

	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
