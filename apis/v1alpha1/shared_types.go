package v1alpha1

// Duration is a string value representing a duration in time.
// Duration can be specified in milliseconds (ms), seconds (s), minutes (m), hours (h).
// A value without a suffix is seconds.
// Examples: 120s, 50ms, 5m, 1h.
//
// +kubebuilder:validation:Pattern=`^[0-9]{1,4}(ms|s|m|h)?$`
type Duration string

// SpanAttribute is a key value pair to be added to a tracing span.
type SpanAttribute struct {
	// Key is the key for a span attribute.
	// Format: must have all '"' escaped and must not contain any '$' or end with an unescaped '\'
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Pattern=`^([^"$\\]|\\[^$])*$`
	Key string `json:"key"`

	// Value is the value for a span attribute.
	// Format: must have all '"' escaped and must not contain any '$' or end with an unescaped '\'
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// +kubebuilder:validation:Pattern=`^([^"$\\]|\\[^$])*$`
	Value string `json:"value"`
}
