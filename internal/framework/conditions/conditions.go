package conditions

import (
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

const (
	// GatewayClassReasonGatewayClassConflict indicates there are multiple GatewayClass resources
	// that reference this controller, and we ignored the resource in question and picked the
	// GatewayClass that is referenced in the command-line argument.
	// This reason is used with GatewayClassConditionAccepted (false).
	GatewayClassReasonGatewayClassConflict v1.GatewayClassConditionReason = "GatewayClassConflict"

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

// NewDefaultGatewayClassConditions returns Conditions that indicate that the GatewayClass is accepted and that the
// Gateway API CRD versions are supported.
func NewDefaultGatewayClassConditions() []Condition {
	return []Condition{
		{
			Type:    string(v1.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1.GatewayClassReasonAccepted),
			Message: "GatewayClass is accepted",
		},
		{
			Type:    string(v1.GatewayClassConditionStatusSupportedVersion),
			Status:  metav1.ConditionTrue,
			Reason:  string(v1.GatewayClassReasonSupportedVersion),
			Message: "Gateway API CRD versions are supported",
		},
	}
}

// NewGatewayClassSupportedVersionBestEffort returns a Condition that indicates that the GatewayClass is accepted,
// but the Gateway API CRD versions are not supported. This means NGF will attempt to generate configuration,
// but it does not guarantee support.
func NewGatewayClassSupportedVersionBestEffort(recommendedVersion string) []Condition {
	return []Condition{
		{
			Type:   string(v1.GatewayClassConditionStatusSupportedVersion),
			Status: metav1.ConditionFalse,
			Reason: string(v1.GatewayClassReasonUnsupportedVersion),
			Message: fmt.Sprintf(
				"Gateway API CRD versions are not recommended. Recommended version is %s",
				recommendedVersion,
			),
		},
	}
}

// NewGatewayClassUnsupportedVersion returns Conditions that indicate that the GatewayClass is not accepted because
// the Gateway API CRD versions are not supported. NGF will not generate configuration in this case.
func NewGatewayClassUnsupportedVersion(recommendedVersion string) []Condition {
	return []Condition{
		{
			Type:   string(v1.GatewayClassConditionStatusAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(v1.GatewayClassReasonUnsupportedVersion),
			Message: fmt.Sprintf(
				"Gateway API CRD versions are not supported. Please install version %s",
				recommendedVersion,
			),
		},
		{
			Type:   string(v1.GatewayClassConditionStatusSupportedVersion),
			Status: metav1.ConditionFalse,
			Reason: string(v1.GatewayClassReasonUnsupportedVersion),
			Message: fmt.Sprintf(
				"Gateway API CRD versions are not supported. Please install version %s",
				recommendedVersion,
			),
		},
	}
}

// NewGatewayClassConflict returns a Condition that indicates that the GatewayClass is not accepted
// due to a conflict with another GatewayClass.
func NewGatewayClassConflict() Condition {
	return Condition{
		Type:    string(v1.GatewayClassConditionStatusAccepted),
		Status:  metav1.ConditionFalse,
		Reason:  string(GatewayClassReasonGatewayClassConflict),
		Message: GatewayClassMessageGatewayClassConflict,
	}
}
