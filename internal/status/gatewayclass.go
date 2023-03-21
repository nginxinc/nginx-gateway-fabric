package status

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// prepareGatewayClassStatus prepares the status for the GatewayClass resource.
func prepareGatewayClassStatus(status state.GatewayClassStatus, transitionTime metav1.Time) v1beta1.GatewayClassStatus {
	return v1beta1.GatewayClassStatus{
		Conditions: convertConditions(status.Conditions, status.ObservedGeneration, transitionTime),
	}
}
