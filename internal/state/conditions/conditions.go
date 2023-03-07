package conditions

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

const (
	// RouteReasonInvalidListener is used with the "Accepted" condition when the Route references an invalid listener.
	RouteReasonInvalidListener v1beta1.RouteConditionReason = "InvalidListener"

	// ListenerReasonUnsupportedValue is used with the "Accepted" condition when a value of a field in a Listener
	// is invalid or not supported.
	ListenerReasonUnsupportedValue v1beta1.ListenerConditionReason = "UnsupportedValue"

	// RouteReasonBackendRefUnsupportedValue is used with the "ResolvedRefs" condition when one of the
	// Route rules has a backendRef with an unsupported value.
	RouteReasonBackendRefUnsupportedValue = "UnsupportedValue"
)

// Condition defines a condition to be reported in the status of resources.
type Condition struct {
	Type    string
	Status  metav1.ConditionStatus
	Reason  string
	Message string
}

// DeduplicateConditions removes duplicate conditions based on the condition type.
// The last condition wins. The order of conditions is preserved.
func DeduplicateConditions(conds []Condition) []Condition {
	type elem struct {
		cond       Condition
		reverseIdx int
	}

	uniqueElems := make(map[string]elem)

	idx := 0
	for i := len(conds) - 1; i >= 0; i-- {
		if _, exist := uniqueElems[conds[i].Type]; exist {
			continue
		}

		uniqueElems[conds[i].Type] = elem{
			cond:       conds[i],
			reverseIdx: idx,
		}
		idx++
	}

	result := make([]Condition, len(uniqueElems))

	for _, el := range uniqueElems {
		result[len(result)-el.reverseIdx-1] = el.cond
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

// NewRouteUnsupportedValue returns a Condition that indicates that the HTTPRoute includes an unsupported value.
func NewRouteUnsupportedValue(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonUnsupportedValue),
		Message: msg,
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
		Message: "Listener is invalid for this parent ref",
	}
}

// NewListenerPortUnavailable returns a Condition that indicates a port is unavailable in a Listener.
func NewListenerPortUnavailable(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.ListenerReasonPortUnavailable),
		Message: msg,
	}
}

// NewDefaultListenerConditions returns the default Conditions that must be present in the status of a Listener.
func NewDefaultListenerConditions() []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1beta1.ListenerReasonAccepted),
			Message: "Listener is accepted",
		},
		{
			Type:    string(v1beta1.ListenerReasonResolvedRefs),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1beta1.ListenerReasonResolvedRefs),
			Message: "All references are resolved",
		},
		{
			Type:    string(v1beta1.ListenerConditionConflicted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonNoConflicts),
			Message: "No conflicts",
		},
	}
}

// NewListenerUnsupportedValue returns a Condition that indicates that a field of a Listener has an unsupported value.
// Unsupported means that the value is not supported by the implementation or invalid.
func NewListenerUnsupportedValue(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(ListenerReasonUnsupportedValue),
		Message: msg,
	}
}

// NewListenerInvalidCertificateRef returns Conditions that indicate that a CertificateRef of a Listener is invalid.
func NewListenerInvalidCertificateRef(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonInvalidCertificateRef),
			Message: msg,
		},
		{
			Type:    string(v1beta1.ListenerReasonResolvedRefs),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonInvalidCertificateRef),
			Message: msg,
		},
	}
}

// NewListenerConflictedHostname returns Conditions that indicate that a hostname of a Listener is conflicted.
func NewListenerConflictedHostname(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonHostnameConflict),
			Message: msg,
		},
		{
			Type:    string(v1beta1.ListenerConditionConflicted),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1beta1.ListenerReasonHostnameConflict),
			Message: msg,
		},
	}
}

// NewListenerUnsupportedAddress returns a Condition that indicates that the address of a Listener is unsupported.
func NewListenerUnsupportedAddress(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.ListenerReasonUnsupportedAddress),
		Message: msg,
	}
}

// NewListenerUnsupportedProtocol returns a Condition that indicates that the protocol of a Listener is unsupported.
func NewListenerUnsupportedProtocol(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.ListenerReasonUnsupportedProtocol),
		Message: msg,
	}
}

// NewRouteBackendRefInvalidKind returns a Condition that indicates that the Route has a backendRef with an
// invalid kind.
func NewRouteBackendRefInvalidKind(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionResolvedRefs),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonInvalidKind),
		Message: msg,
	}
}

// NewRouteBackendRefRefNotPermitted returns a Condition that indicates that the Route has a backendRef that
// is not permitted.
func NewRouteBackendRefRefNotPermitted(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionResolvedRefs),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonRefNotPermitted),
		Message: msg,
	}
}

// NewRouteBackendRefRefBackendNotFound returns a Condition that indicates that the Route has a backendRef that
// points to non-existing backend.
func NewRouteBackendRefRefBackendNotFound(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionResolvedRefs),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonBackendNotFound),
		Message: msg,
	}
}

// NewRouteBackendRefUnsupportedValue returns a Condition that indicates that the Route has a backendRef with
// an unsupported value.
func NewRouteBackendRefUnsupportedValue(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionResolvedRefs),
		Status:  metav1.ConditionFalse,
		Reason:  RouteReasonBackendRefUnsupportedValue,
		Message: msg,
	}
}
