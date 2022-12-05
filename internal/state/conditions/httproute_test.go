package conditions

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeduplicateDeduplicateRouteConditions(t *testing.T) {
	g := NewGomegaWithT(t)

	conds := []RouteCondition{
		{
			Type:   "Type1",
			Status: metav1.ConditionTrue,
		},
		{
			Type:   "Type1",
			Status: metav1.ConditionFalse,
		},
		{
			Type:   "Type2",
			Status: metav1.ConditionFalse,
		},
		{
			Type:   "Type2",
			Status: metav1.ConditionTrue,
		},
		{
			Type:   "Type3",
			Status: metav1.ConditionTrue,
		},
	}

	expected := []RouteCondition{
		{
			Type:   "Type1",
			Status: metav1.ConditionFalse,
		},
		{
			Type:   "Type2",
			Status: metav1.ConditionTrue,
		},
		{
			Type:   "Type3",
			Status: metav1.ConditionTrue,
		},
	}

	result := DeduplicateRouteConditions(conds)
	g.Expect(result).Should(ConsistOf(expected))
}
