package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:resource:categories=nginx-gateway-fabric,scope=Cluster
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// NginxProxy is a configuration object that is attached to a GatewayClass parametersRef. It provides a way
// to configure global settings for all Gateways defined from the GatewayClass.
type NginxProxy struct { //nolint:govet // standard field alignment, don't change it
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the NginxProxy.
	Spec NginxProxySpec `json:"spec"`
}

// +kubebuilder:object:root=true

// NginxProxyList contains a list of NginxProxies.
type NginxProxyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NginxProxy `json:"items"`
}

// IPFamilyType specifies the IP family to be used by NGINX.
//
// +kubebuilder:validation:Enum=dual;ipv4;ipv6
// +kubebuilder:default:=dual
type IPFamilyType string

const (
	// Dual specifies that the server will use both IPv4 and IPv6.
	Dual IPFamilyType = "dual"
	// IPv4 specifies that the server will use only IPv4.
	IPv4 IPFamilyType = "ipv4"
	// IPv6 specifies that the server will use only IPv6.
	IPv6 IPFamilyType = "ipv6"
)

// NginxProxySpec defines the desired state of the NginxProxy.
type NginxProxySpec struct {
	// IPFamily specifies the IP family to be used by the NGINX.
	// Default is "dual", meaning the server will use both IPv4 and IPv6.
	//
	// +optional
	IPFamily *IPFamilyType `json:"ipFamily,omitempty"`
	// Telemetry specifies the OpenTelemetry configuration.
	//
	// +optional
	Telemetry *Telemetry `json:"telemetry,omitempty"`
	// DisableHTTP2 defines if http2 should be disabled for all servers.
	// Default is false, meaning http2 will be enabled for all servers.
	//
	// +optional
	DisableHTTP2 bool `json:"disableHTTP2,omitempty"`
}

// Telemetry specifies the OpenTelemetry configuration.
type Telemetry struct {
	// Exporter specifies OpenTelemetry export parameters.
	//
	// +optional
	Exporter *TelemetryExporter `json:"exporter,omitempty"`

	// ServiceName is the "service.name" attribute of the OpenTelemetry resource.
	// Default is 'ngf:<gateway-namespace>:<gateway-name>'. If a value is provided by the user,
	// then the default becomes a prefix to that value.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=127
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_-]+$`
	ServiceName *string `json:"serviceName,omitempty"`

	// SpanAttributes are custom key/value attributes that are added to each span.
	//
	// +optional
	// +listType=map
	// +listMapKey=key
	// +kubebuilder:validation:MaxItems=64
	SpanAttributes []SpanAttribute `json:"spanAttributes,omitempty"`
}

// TelemetryExporter specifies OpenTelemetry export parameters.
type TelemetryExporter struct {
	// Interval is the maximum interval between two exports.
	// Default: https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter
	//
	// +optional
	Interval *Duration `json:"interval,omitempty"`

	// BatchSize is the maximum number of spans to be sent in one batch per worker.
	// Default: https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	BatchSize *int32 `json:"batchSize,omitempty"`

	// BatchCount is the number of pending batches per worker, spans exceeding the limit are dropped.
	// Default: https://nginx.org/en/docs/ngx_otel_module.html#otel_exporter
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	BatchCount *int32 `json:"batchCount,omitempty"`

	// Endpoint is the address of OTLP/gRPC endpoint that will accept telemetry data.
	// Format: alphanumeric hostname with optional http scheme and optional port.
	//
	//nolint:lll
	// +kubebuilder:validation:Pattern=`^(?:http?:\/\/)?[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?(?:\.[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?)*(?::\d{1,5})?$`
	Endpoint string `json:"endpoint"`
}
