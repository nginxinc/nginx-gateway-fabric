package conditions

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

// RouteReasonInvalidListener is used with the "Accepted" condition when the route references an invalid listener.
const RouteReasonInvalidListener v1beta1.RouteConditionReason = "InvalidListener"

// RouteCondition defines a condition to be reported in the status of an HTTPRoute.
type RouteCondition struct {
	Type    v1beta1.RouteConditionType
	Status  metav1.ConditionStatus
	Reason  v1beta1.RouteConditionReason
	Message string
}

// DeduplicateRouteConditions removes duplicate conditions based on the condition type.
// The last condition wins.
func DeduplicateRouteConditions(conds []RouteCondition) []RouteCondition {
	uniqueConds := make(map[v1beta1.RouteConditionType]RouteCondition)

	for _, cond := range conds {
		uniqueConds[cond.Type] = cond
	}

	result := make([]RouteCondition, 0, len(uniqueConds))

	for _, cond := range uniqueConds {
		result = append(result, cond)
	}

	return result
}

// NewDefaultRouteConditions returns the default conditions that must be present in the status of an HTTPRoute.
func NewDefaultRouteConditions() []RouteCondition {
	return []RouteCondition{
		NewRouteAccepted(),
	}
}

// NewRouteNoMatchingListenerHostname returns a RouteCondition that indicates that the hostname of the listener
// does not match the hostnames of the HTTPRoute.
func NewRouteNoMatchingListenerHostname() RouteCondition {
	return RouteCondition{
		Type:    v1beta1.RouteConditionAccepted,
		Status:  metav1.ConditionFalse,
		Reason:  v1beta1.RouteReasonNoMatchingListenerHostname,
		Message: "Listener hostname does not match the HTTPRoute hostnames",
	}
}

// NewRouteAccepted returns a RouteCondition that indicates that the HTTPRoute is accepted.
func NewRouteAccepted() RouteCondition {
	return RouteCondition{
		Type:    v1beta1.RouteConditionAccepted,
		Status:  metav1.ConditionTrue,
		Reason:  v1beta1.RouteReasonAccepted,
		Message: "The route is accepted",
	}
}

// NewRouteTODO returns a RouteCondition that can be used as a placeholder for a condition that is not yet implemented.
func NewRouteTODO(msg string) RouteCondition {
	return RouteCondition{
		Type:    "TODO",
		Status:  metav1.ConditionTrue,
		Reason:  "TODO",
		Message: fmt.Sprintf("The condition for this has not been implemented yet: %s", msg),
	}
}

// NewRouteInvalidListener returns a RouteCondition that indicates that the HTTPRoute is not accepted because of an
// invalid listener.
func NewRouteInvalidListener() RouteCondition {
	return RouteCondition{
		Type:    v1beta1.RouteConditionAccepted,
		Status:  metav1.ConditionFalse,
		Reason:  RouteReasonInvalidListener,
		Message: "The listener is invalid for this parent ref",
	}
}
