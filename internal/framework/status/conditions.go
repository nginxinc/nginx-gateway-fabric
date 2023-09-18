package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
)

func convertConditions(
	conds []conditions.Condition,
	observedGeneration int64,
	transitionTime metav1.Time,
) []metav1.Condition {
	apiConds := make([]metav1.Condition, len(conds))

	for i := range conds {
		apiConds[i] = metav1.Condition{
			Type:               conds[i].Type,
			Status:             conds[i].Status,
			ObservedGeneration: observedGeneration,
			LastTransitionTime: transitionTime,
			Reason:             conds[i].Reason,
			Message:            conds[i].Message,
		}
	}

	return apiConds
}
