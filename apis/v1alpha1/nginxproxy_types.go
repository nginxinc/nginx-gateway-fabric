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
type NginxProxySpec struct {
	HTTP *HTTP `json:"http,omitempty"`
}

// HTTP defines the NGINX HTTP block configuration.
type HTTP struct {
	Telemetry *Telemetry `json:"telemetry,omitempty"`
}

// Telemetry defines the telemetry configuration.
type Telemetry struct {
	Tracing *Tracing `json:"tracing,omitempty"`
}

// Tracing defines the tracing configuration.
type Tracing struct {
	//
	// Interval specifies the tracing interval. Default is 5s.
	//
	// +optional
	// +kubebuilder:default="5s"
	// +kubebuilder:validation:Pattern=`^(\d+y)??\s*(\d+M)??\s*(\d+w)??\s*(\d+d)??\s*(\d+h)??\s*(\d+m)??\s*(\d+s?)??\s*(\d+ms)??$`
	Interval *string `json:"interval,omitempty"`
	//
	// BatchSize specifies the maximum number of spans to be sent in one batch per worker. Default is 512.
	//
	// +optional
	// +kubebuilder:default=512
	// +kubebuilder:validation:Minimum=1
	BatchSize *int32 `json:"batchSize,omitempty"`
	//
	// BatchCount specifies the number of pending batches per worker, spans exceeding the limit are dropped. Default is 4.
	//
	// +optional
	// +kubebuilder:default=4
	// +kubebuilder:validation:Minimum=1
	BatchCount *int32 `json:"batchCount,omitempty"`
	//
	// Enable enables or disables OpenTelemetry tracing at the HTTP context. Default is false.
	//
	// +optional
	// +kubebuilder:default=false
	Enable *bool `json:"enable,omitempty"`
	//
	// Endpoint specifies the address of OTLP/gRPC endpoint that will accept telemetry data.
	// It should be in the form 'hostname.my.domain:<port>' or '<IP:<port>'
	// e.g. 'simplest-collector.default.svc.cluster.local:4317' or '10.7.8.9:4318'
	//
	// +required
	// ++kubebuilder:validation:Required
	Endpoint string `json:"endpoint"`
}

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
