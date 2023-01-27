package conditions

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// RouteReasonInvalidListener is used with the "Accepted" condition when the route references an invalid listener.
const RouteReasonInvalidListener v1beta1.RouteConditionReason = "InvalidListener"

// Condition defines a condition to be reported in the status of resources.
type Condition struct {
	Type    string
	Status  metav1.ConditionStatus
	Reason  string
	Message string
}

// DeduplicateConditions removes duplicate conditions based on the condition type.
// The last condition wins.
func DeduplicateConditions(conds []Condition) []Condition {
	uniqueConds := make(map[string]Condition)

	for _, cond := range conds {
		uniqueConds[cond.Type] = cond
	}

	result := make([]Condition, 0, len(uniqueConds))

	for _, cond := range uniqueConds {
		result = append(result, cond)
	}

	return result
}

// NewDefaultRouteConditions returns the default conditions that must be present in the status of an HTTPRoute.
func NewDefaultRouteConditions() []Condition {
	return []Condition{
		NewRouteAccepted(),
	}
}

// NewRouteNoMatchingListenerHostname returns a Condition that indicates that the hostname of the listener
// does not match the hostnames of the HTTPRoute.
func NewRouteNoMatchingListenerHostname() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonNoMatchingListenerHostname),
		Message: "Listener hostname does not match the HTTPRoute hostnames",
	}
}

// NewRouteAccepted returns a Condition that indicates that the HTTPRoute is accepted.
func NewRouteAccepted() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.RouteReasonAccepted),
		Message: "The route is accepted",
	}
}

// NewTODO returns a Condition that can be used as a placeholder for a condition that is not yet implemented.
func NewTODO(msg string) Condition {
	return Condition{
		Type:    "TODO",
		Status:  metav1.ConditionTrue,
		Reason:  "TODO",
		Message: fmt.Sprintf("The condition for this has not been implemented yet: %s", msg),
	}
}

// NewRouteInvalidListener returns a Condition that indicates that the HTTPRoute is not accepted because of an
// invalid listener.
func NewRouteInvalidListener() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(RouteReasonInvalidListener),
		Message: "The listener is invalid for this parent ref",
	}
}
