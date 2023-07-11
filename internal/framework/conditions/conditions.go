package conditions

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	// GatewayClassReasonGatewayClassConflict indicates there are multiple GatewayClass resources
	// that reference this controller, and we ignored the resource in question and picked the
	// GatewayClass that is referenced in the command-line argument.
	// This reason is used with GatewayClassConditionAccepted (false).
	GatewayClassReasonGatewayClassConflict v1beta1.GatewayClassConditionReason = "GatewayClassConflict"

	// GatewayClassMessageGatewayClassConflict is a message that describes GatewayClassReasonGatewayClassConflict.
	GatewayClassMessageGatewayClassConflict = "The resource is ignored due to a conflicting GatewayClass resource"
)

// Condition defines a condition to be reported in the status of resources.
type Condition struct {
	Type    string
	Status  metav1.ConditionStatus
	Reason  string
	Message string
}

// NewDefaultGatewayClassConditions returns the default Conditions that must be present in the status of a GatewayClass.
func NewDefaultGatewayClassConditions() []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1beta1.GatewayClassReasonAccepted),
			Message: "GatewayClass is accepted",
		},
	}
}

// NewGatewayClassConflict returns a Condition that indicates that the GatewayClass is not accepted
// due to a conflict with another GatewayClass.
func NewGatewayClassConflict() Condition {
	return Condition{
		Type:    string(v1beta1.GatewayClassConditionStatusAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(GatewayClassReasonGatewayClassConflict),
		Message: GatewayClassMessageGatewayClassConflict,
	}
}
