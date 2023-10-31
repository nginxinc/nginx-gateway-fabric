package status

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func CreateTestConditions(condType string) []conditions.Condition {
	return []conditions.Condition{
		{
			Type:    condType,
			Status:  metav1.ConditionTrue,
			Reason:  "TestReason1",
			Message: "Test message1",
		},
		{
			Type:    condType,
			Status:  metav1.ConditionFalse,
			Reason:  "TestReason2",
			Message: "Test message2",
		},
	}
}

func CreateExpectedAPIConditions(
	condType string,
	observedGeneration int64,
	transitionTime metav1.Time,
) []metav1.Condition {
	return []metav1.Condition{
		{
			Type:               condType,
			Status:             metav1.ConditionTrue,
			ObservedGeneration: observedGeneration,
			LastTransitionTime: transitionTime,
			Reason:             "TestReason1",
			Message:            "Test message1",
		},
		{
			Type:               condType,
			Status:             metav1.ConditionFalse,
			ObservedGeneration: observedGeneration,
			LastTransitionTime: transitionTime,
			Reason:             "TestReason2",
			Message:            "Test message2",
		},
	}
}

func TestConvertRouteConditions(t *testing.T) {
	g := NewWithT(t)

	var generation int64 = 1
	transitionTime := metav1.NewTime(time.Now())

	expected := CreateExpectedAPIConditions("Test", generation, transitionTime)

	result := convertConditions(CreateTestConditions("Test"), generation, transitionTime)
	g.Expect(helpers.Diff(expected, result)).To(BeEmpty())
}
