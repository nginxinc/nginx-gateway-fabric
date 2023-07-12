package conditions

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
)

func TestDeduplicateConditions(t *testing.T) {
	g := NewGomegaWithT(t)

	conds := []conditions.Condition{
		{
			Type:    "Type1",
			Status:  metav1.ConditionTrue,
			Message: "0",
		},
		{
			Type:    "Type1",
			Status:  metav1.ConditionFalse,
			Message: "1",
		},
		{
			Type:    "Type2",
			Status:  metav1.ConditionFalse,
			Message: "2",
		},
		{
			Type:    "Type2",
			Status:  metav1.ConditionTrue,
			Message: "3",
		},
		{
			Type:    "Type3",
			Status:  metav1.ConditionTrue,
			Message: "4",
		},
	}

	expected := []conditions.Condition{
		{
			Type:    "Type1",
			Status:  metav1.ConditionFalse,
			Message: "1",
		},
		{
			Type:    "Type2",
			Status:  metav1.ConditionTrue,
			Message: "3",
		},
		{
			Type:    "Type3",
			Status:  metav1.ConditionTrue,
			Message: "4",
		},
	}

	result := DeduplicateConditions(conds)
	g.Expect(result).Should(Equal(expected))
}
