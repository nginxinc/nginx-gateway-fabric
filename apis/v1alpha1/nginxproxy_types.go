package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status

// NginxProxy represents the dynamic configuration for an NGINX Gateway Fabric data plane.
type NginxProxy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// NginxProxySpec defines the desired state of the NginxProxy.
	Spec NginxProxySpec `json:"spec"`

	// NginxProxyStatus defines the state of the NginxProxy.
	Status NginxProxyStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// NginxProxyList contains a list of NginxProxies.
type NginxProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxProxy `json:"items"`
}

// NginxProxySpec defines the desired state of the NginxProxy.
type NginxProxySpec struct{}

// NginxProxyStatus defines the state of the NginxProxy.
type NginxProxyStatus struct {
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// NginxProxyConditionType is a type of condition associated with an
// NginxProxy. This type should be used with the NginxProxyStatus.Conditions field.
type NginxProxyConditionType string

// NginxProxyConditionReason defines the set of reasons that explain why a
// particular NginxProxy condition type has been raised.
type NginxProxyConditionReason string

const (
	// NginxProxyConditionAccepted is a condition that is true when the NginxProxy
	// configuration is syntactically and semantically valid.
	NginxProxyConditionAccepted NginxProxyConditionType = "Accepted"

	// NginxProxyReasonAccepted is a reason that is used with the "Accepted" condition when the condition is True.
	NginxProxyReasonAccepted NginxProxyConditionReason = "Accepted"

	// NginxProxyConditionProgrammed is a condition that is true when the NginxProxy has resulted in
	// successful nginx configuration.
	NginxProxyConditionProgrammed NginxProxyConditionType = "Programmed"

	// NginxProxyReasonProgrammed is a reason that is used with the "Programmed" condition when the condition is True.
	NginxProxyReasonProgrammed NginxProxyConditionReason = "Programmed"

	// NginxProxyReasonInvalid is a reason that is used with the "Accepted" or "Programmed" condition when
	// an error occurs with validation or reloading nginx.
	NginxProxyReasonInvalid NginxProxyConditionReason = "Invalid"
)
