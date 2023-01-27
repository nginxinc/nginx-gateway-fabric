package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
)

func CreateTestConditions() []conditions.Condition {
	return []conditions.Condition{
		{
			Type:    "Test",
			Status:  metav1.ConditionTrue,
			Reason:  "TestReason1",
			Message: "Test message1",
		},
		{
			Type:    "Test",
			Status:  metav1.ConditionFalse,
			Reason:  "TestReason2",
			Message: "Test message2",
		},
	}
}

func CreateExpectedAPIConditions(observedGeneration int64, transitionTime metav1.Time) []metav1.Condition {
	return []metav1.Condition{
		{
			Type:               "Test",
			Status:             metav1.ConditionTrue,
			ObservedGeneration: observedGeneration,
			LastTransitionTime: transitionTime,
			Reason:             "TestReason1",
			Message:            "Test message1",
		},
		{
			Type:               "Test",
			Status:             metav1.ConditionFalse,
			ObservedGeneration: observedGeneration,
			LastTransitionTime: transitionTime,
			Reason:             "TestReason2",
			Message:            "Test message2",
		},
	}
}

func TestConvertRouteConditions(t *testing.T) {
	g := NewGomegaWithT(t)

	var generation int64 = 1
	transitionTime := metav1.NewTime(time.Now())

	expected := CreateExpectedAPIConditions(generation, transitionTime)

	result := convertConditions(CreateTestConditions(), generation, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
