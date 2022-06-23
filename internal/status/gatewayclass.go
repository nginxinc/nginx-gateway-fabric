package status

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// prepareGatewayClassStatus prepares the status for the GatewayClass resource.
func prepareGatewayClassStatus(status state.GatewayClassStatus, transitionTime metav1.Time) v1alpha2.GatewayClassStatus {
	var (
		condStatus metav1.ConditionStatus
		msg        string
	)

	if status.Valid {
		condStatus = metav1.ConditionTrue
		msg = "GatewayClass has been accepted"
	} else {
		condStatus = metav1.ConditionFalse
		msg = fmt.Sprintf("GatewayClass has been rejected: %s", status.ErrorMsg)
	}

	cond := metav1.Condition{
		Type:               string(v1alpha2.GatewayClassConditionStatusAccepted),
		Status:             condStatus,
		ObservedGeneration: status.ObservedGeneration,
		LastTransitionTime: transitionTime,
		Reason:             string(v1alpha2.GatewayClassReasonAccepted),
		Message:            msg,
	}

	return v1alpha2.GatewayClassStatus{
		Conditions: []metav1.Condition{cond},
	}
}
