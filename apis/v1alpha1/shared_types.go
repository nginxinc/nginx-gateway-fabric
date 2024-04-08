package v1alpha1

// Duration is a string value representing a duration in time.
// Duration can be specified in milliseconds (ms) or seconds (s) A value without a suffix is seconds.
// Examples: 120s, 50ms.
//
// +kubebuilder:validation:Pattern=`^\d{1,4}(ms|s)?$`
type Duration string
