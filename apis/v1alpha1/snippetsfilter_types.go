package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +kubebuilder:resource:categories=nginx-gateway-fabric,shortName=snippetsfilter
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// SnippetsFilter is a filter that allows inserting NGINX configuration into the
// generated NGINX config for HTTPRoute and GRPCRoute resources.
type SnippetsFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the SnippetsFilter.
	Spec SnippetsFilterSpec `json:"spec"`

	// Status defines the state of the SnippetsFilter.
	Status SnippetsFilterStatus `json:"status,omitempty"`
}

// SnippetsFilterSpec defines the desired state of the SnippetsFilter.
type SnippetsFilterSpec struct {
	// Snippets is a list of NGINX configuration snippets.
	// There can only be one snippet per context.
	// Allowed contexts: http, http.server, http.server.location.
	Snippets []Snippet `json:"snippets"`
}

// Snippet represents an NGINX configuration snippet.
type Snippet struct {
	// Context is the NGINX context to insert the snippet into.
	Context NginxContext `json:"context"`

	// Value is the NGINX configuration snippet.
	Value string `json:"value"`
}

// SnippetsFilterList contains a list of SnippetFilters.
type SnippetsFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Snippet `json:"items"`
}

// SnippetsFilterStatus defines the state of SnippetsFilter.
type SnippetsFilterStatus struct {
	// Conditions describes the state of the SnippetsFilter.
	// +optional
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=8
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// SnippetsFilterConditionType is a type of condition associated with SnippetsFilter.
type SnippetsFilterConditionType string

// SnippetsFilterConditionReason is a reason for a SnippetsFilter condition type.
type SnippetsFilterConditionReason string

const (
	// SnippetsFilterConditionTypeAccepted indicates that the SnippetsFilter is accepted.
	//
	// Possible reasons for this condition to be True:
	//
	// * Accepted
	//
	// Possible reasons for this condition to be False:
	//
	// * Invalid.
	SnippetsFilterConditionTypeAccepted SnippetsFilterConditionType = "Accepted"

	// SnippetsFilterConditionReasonAccepted is used with the Accepted condition type when
	// the condition is true.
	SnippetsFilterConditionReasonAccepted SnippetsFilterConditionReason = "Accepted"

	// SnippetsFilterConditionTypeInvalid is used with the Accepted condition type when
	// SnippetsFilter is invalid.
	SnippetsFilterConditionTypeInvalid SnippetsFilterConditionType = "Invalid"
)

// NginxContext represents the NGINX configuration context.
//
// +kubebuilder:validation:Enum=http;http.server;http.server.location;
type NginxContext string

const (
	// NginxContextHTTP is the http context of the NGINX configuration.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#http
	NginxContextHTTP NginxContext = "http"

	// NginxContextHTTPServer is the server context of the NGINX configuration.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#server
	NginxContextHTTPServer NginxContext = "http.server"

	// NginxContextHTTPServerLocation is the location context of the NGINX configuration.
	// https://nginx.org/en/docs/http/ngx_http_core_module.html#location
	NginxContextHTTPServerLocation NginxContext = "http.server.location"
)
