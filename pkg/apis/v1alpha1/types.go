package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:validation:Optional
// +kubebuilder:resource:shortName=gcfg,scope=Cluster
type GatewayConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GatewayConfigSpec `json:"spec"`
}

type GatewayConfigSpec struct {
	Worker *Worker `json:"worker,omitempty"`
	HTTP   *HTTP   `json:"http,omitempty"`
}

type Worker struct {
	Processes *int `json:"processes,omitempty"`
}

type HTTP struct {
	AccessLogs []AccessLog `json:"accessLogs,omitempty"`
}

type AccessLog struct {
	Format      string `json:"format"`
	Destination string `json:"destination"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// GatewayConfigList is a list of the GatewayConfig resources.
type GatewayConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []GatewayConfig `json:"items"`
}
