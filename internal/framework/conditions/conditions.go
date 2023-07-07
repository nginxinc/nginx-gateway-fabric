package conditions

import (
	"fmt"

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

	// ListenerReasonUnsupportedValue is used with the "Accepted" condition when a value of a field in a Listener
	// is invalid or not supported.
	ListenerReasonUnsupportedValue v1beta1.ListenerConditionReason = "UnsupportedValue"

	// ListenerMessageFailedNginxReload is a message used with ListenerConditionProgrammed (false)
	// when nginx fails to reload.
	ListenerMessageFailedNginxReload = "The Listener is not programmed due to a failure to " +
		"reload nginx with the configuration"

	// RouteReasonBackendRefUnsupportedValue is used with the "ResolvedRefs" condition when one of the
	// Route rules has a backendRef with an unsupported value.
	RouteReasonBackendRefUnsupportedValue = "UnsupportedValue"

	// RouteReasonInvalidGateway is used with the "Accepted" (false) condition when the Gateway the Route
	// references is invalid.
	RouteReasonInvalidGateway = "InvalidGateway"

	// RouteReasonInvalidListener is used with the "Accepted" condition when the Route references an invalid listener.
	RouteReasonInvalidListener v1beta1.RouteConditionReason = "InvalidListener"

	// RouteReasonGatewayNotProgrammed is used when the associated Gateway is not programmed.
	// Used with Accepted (false).
	RouteReasonGatewayNotProgrammed v1beta1.RouteConditionReason = "GatewayNotProgrammed"

	// GatewayReasonGatewayConflict indicates there are multiple Gateway resources to choose from,
	// and we ignored the resource in question and picked another Gateway as the winner.
	// This reason is used with GatewayConditionAccepted (false).
	GatewayReasonGatewayConflict v1beta1.GatewayConditionReason = "GatewayConflict"

	// GatewayMessageGatewayConflict is a message that describes GatewayReasonGatewayConflict.
	GatewayMessageGatewayConflict = "The resource is ignored due to a conflicting Gateway resource"

	// GatewayReasonUnsupportedValue is used with GatewayConditionAccepted (false) when a value of a field in a Gateway
	// is invalid or not supported.
	GatewayReasonUnsupportedValue v1beta1.GatewayConditionReason = "UnsupportedValue"

	// GatewayMessageFailedNginxReload is a message used with GatewayConditionProgrammed (false)
	// when nginx fails to reload.
	GatewayMessageFailedNginxReload = "The Gateway is not programmed due to a failure to " +
		"reload nginx with the configuration"

	// RouteMessageFailedNginxReload is a message used with RouteReasonGatewayNotProgrammed
	// when nginx fails to reload.
	RouteMessageFailedNginxReload = GatewayMessageFailedNginxReload + ". NGINX may still be configured " +
		"for this HTTPRoute. However, future updates to this resource will not be configured until the Gateway " +
		"is programmed again"
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

// NewTODO returns a Condition that can be used as a placeholder for a condition that is not yet implemented.
func NewTODO(msg string) Condition {
	return Condition{
		Type:    "TODO",
		Status:  metav1.ConditionTrue,
		Reason:  "TODO",
		Message: fmt.Sprintf("The condition for this has not been implemented yet: %s", msg),
	}
}

// NewDefaultRouteConditions returns the default conditions that must be present in the status of an HTTPRoute.
func NewDefaultRouteConditions() []Condition {
	return []Condition{
		NewRouteAccepted(),
		NewRouteResolvedRefs(),
	}
}

// NewRouteNotAllowedByListeners returns a Condition that indicates that the HTTPRoute is not allowed by
// any listener.
func NewRouteNotAllowedByListeners() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonNotAllowedByListeners),
		Message: "HTTPRoute is not allowed by any listener",
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

// NewRouteResolvedRefs returns a Condition that indicates that all the references on the Route are resolved.
func NewRouteResolvedRefs() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionResolvedRefs),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.RouteReasonResolvedRefs),
		Message: "All references are resolved",
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

// NewRouteInvalidGateway returns a Condition that indicates that the Route is not Accepted because the Gateway it
// references is invalid.
func NewRouteInvalidGateway() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  RouteReasonInvalidGateway,
		Message: "Gateway is invalid",
	}
}

// NewRouteNoMatchingParent returns a Condition that indicates that the Route is not Accepted because
// it specifies a Port and/or SectionName that does not match any Listeners in the Gateway.
func NewRouteNoMatchingParent() Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.RouteReasonNoMatchingParent),
		Message: "Listener is not found for this parent ref",
	}
}

// NewRouteGatewayNotProgrammed returns a Condition that indicates that the Gateway it references is not programmed,
// which does not guarantee that the HTTPRoute has been configured.
func NewRouteGatewayNotProgrammed(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.RouteConditionAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(RouteReasonGatewayNotProgrammed),
		Message: msg,
	}
}

// NewDefaultListenerConditions returns the default Conditions that must be present in the status of a Listener.
func NewDefaultListenerConditions() []Condition {
	return []Condition{
		NewListenerAccepted(),
		NewListenerProgrammed(),
		NewListenerResolvedRefs(),
		NewListenerNoConflicts(),
	}
}

// NewListenerAccepted returns a Condition that indicates that the Listener is accepted.
func NewListenerAccepted() Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionAccepted),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.ListenerReasonAccepted),
		Message: "Listener is accepted",
	}
}

// NewListenerProgrammed returns a Condition that indicates the Listener is programmed.
func NewListenerProgrammed() Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionProgrammed),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.ListenerReasonProgrammed),
		Message: "Listener is programmed",
	}
}

// NewListenerResolvedRefs returns a Condition that indicates that all references in a Listener are resolved.
func NewListenerResolvedRefs() Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionResolvedRefs),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.ListenerReasonResolvedRefs),
		Message: "All references are resolved",
	}
}

// NewListenerNoConflicts returns a Condition that indicates that there are no conflicts in a Listener.
func NewListenerNoConflicts() Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionConflicted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.ListenerReasonNoConflicts),
		Message: "No conflicts",
	}
}

// NewListenerNotProgrammedInvalid returns a Condition that indicates the Listener is not programmed because it is
// semantically or syntactically invalid. The provided message contains the details of why the Listener is invalid.
func NewListenerNotProgrammedInvalid(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.ListenerConditionProgrammed),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.ListenerReasonInvalid),
		Message: msg,
	}
}

// NewListenerUnsupportedValue returns Conditions that indicate that a field of a Listener has an unsupported value.
// Unsupported means that the value is not supported by the implementation or invalid.
func NewListenerUnsupportedValue(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(ListenerReasonUnsupportedValue),
			Message: msg,
		},
		NewListenerNotProgrammedInvalid(msg),
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
		NewListenerNotProgrammedInvalid(msg),
	}
}

// NewListenerInvalidRouteKinds returns Conditions that indicate that an invalid or unsupported Route kind is
// specified by the Listener.
func NewListenerInvalidRouteKinds(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerReasonResolvedRefs),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonInvalidRouteKinds),
			Message: msg,
		},
		NewListenerNotProgrammedInvalid(msg),
	}
}

// NewListenerProtocolConflict returns Conditions that indicate multiple Listeners are specified with the same
// Listener port number, but have conflicting protocol specifications.
func NewListenerProtocolConflict(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonProtocolConflict),
			Message: msg,
		},
		{
			Type:    string(v1beta1.ListenerConditionConflicted),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1beta1.ListenerReasonProtocolConflict),
			Message: msg,
		},
		NewListenerNotProgrammedInvalid(msg),
	}
}

// NewListenerUnsupportedProtocol returns Conditions that indicate that the protocol of a Listener is unsupported.
func NewListenerUnsupportedProtocol(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonUnsupportedProtocol),
			Message: msg,
		},
		NewListenerNotProgrammedInvalid(msg),
	}
}

// NewListenerRefNotPermitted returns Conditions that indicates that the Listener references a TLS secret that is not
// permitted by a ReferenceGrant.
func NewListenerRefNotPermitted(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.ListenerConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonRefNotPermitted),
			Message: msg,
		},
		{
			Type:    string(v1beta1.ListenerReasonResolvedRefs),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.ListenerReasonRefNotPermitted),
			Message: msg,
		},
		NewListenerNotProgrammedInvalid(msg),
	}
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

// NewGatewayClassInvalidParameters returns a Condition that indicates that the GatewayClass has invalid parameters.
func NewGatewayClassInvalidParameters(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.GatewayClassConditionStatusAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.GatewayClassReasonInvalidParameters),
		Message: msg,
	}
}

// NewDefaultGatewayConditions returns the default Conditions that must be present in the status of a Gateway.
func NewDefaultGatewayConditions() []Condition {
	return []Condition{
		NewGatewayAccepted(),
		NewGatewayProgrammed(),
	}
}

// NewGatewayAccepted returns a Condition that indicates the Gateway is accepted.
func NewGatewayAccepted() Condition {
	return Condition{
		Type:    string(v1beta1.GatewayConditionAccepted),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.GatewayReasonAccepted),
		Message: "Gateway is accepted",
	}
}

// NewGatewayConflict returns Conditions that indicate the Gateway has a conflict with another Gateway.
func NewGatewayConflict() []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.GatewayConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(GatewayReasonGatewayConflict),
			Message: GatewayMessageGatewayConflict,
		},
		NewGatewayConflictNotProgrammed(),
	}
}

// NewGatewayAcceptedListenersNotValid returns a Condition that indicates the Gateway is accepted,
// but has at least one listener that is invalid.
func NewGatewayAcceptedListenersNotValid() Condition {
	return Condition{
		Type:    string(v1beta1.GatewayConditionAccepted),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.GatewayReasonListenersNotValid),
		Message: "Gateway has at least one valid listener",
	}
}

// NewGatewayNotAcceptedListenersNotValid returns Conditions that indicate the Gateway is not accepted,
// because all listeners are invalid.
func NewGatewayNotAcceptedListenersNotValid() []Condition {
	msg := "Gateway has no valid listeners"
	return []Condition{
		{
			Type:    string(v1beta1.GatewayConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.GatewayReasonListenersNotValid),
			Message: msg,
		},
		NewGatewayNotProgrammedInvalid(msg),
	}
}

// NewGatewayInvalid returns Conditions that indicate the Gateway is not accepted and programmed because it is
// semantically or syntactically invalid. The provided message contains the details of why the Gateway is invalid.
func NewGatewayInvalid(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.GatewayConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(v1beta1.GatewayReasonInvalid),
			Message: msg,
		},
		NewGatewayNotProgrammedInvalid(msg),
	}
}

// NewGatewayUnsupportedValue returns Conditions that indicate that a field of the Gateway has an unsupported value.
// Unsupported means that the value is not supported by the implementation or invalid.
func NewGatewayUnsupportedValue(msg string) []Condition {
	return []Condition{
		{
			Type:    string(v1beta1.GatewayConditionAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  string(GatewayReasonUnsupportedValue),
			Message: msg,
		},
		{
			Type:    string(v1beta1.GatewayConditionProgrammed),
			Status:  metav1.ConditionFalse,
			Reason:  string(GatewayReasonUnsupportedValue),
			Message: msg,
		},
	}
}

// NewGatewayProgrammed returns a Condition that indicates the Gateway is programmed.
func NewGatewayProgrammed() Condition {
	return Condition{
		Type:    string(v1beta1.GatewayConditionProgrammed),
		Status:  metav1.ConditionTrue,
		Reason:  string(v1beta1.GatewayReasonProgrammed),
		Message: "Gateway is programmed",
	}
}

// NewGatewayInvalid returns a Condition that indicates the Gateway is not programmed because it is
// semantically or syntactically invalid. The provided message contains the details of why the Gateway is invalid.
func NewGatewayNotProgrammedInvalid(msg string) Condition {
	return Condition{
		Type:    string(v1beta1.GatewayConditionProgrammed),
		Status:  metav1.ConditionFalse,
		Reason:  string(v1beta1.GatewayReasonInvalid),
		Message: msg,
	}
}

// NewGatewayConflictNotProgrammed returns a custom Programmed Condition that indicates the Gateway has a
// conflict with another Gateway.
func NewGatewayConflictNotProgrammed() Condition {
	return Condition{
		Type:    string(v1beta1.GatewayConditionProgrammed),
		Status:  metav1.ConditionFalse,
		Reason:  string(GatewayReasonGatewayConflict),
		Message: GatewayMessageGatewayConflict,
	}
}
