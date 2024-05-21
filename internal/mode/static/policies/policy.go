package policies

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// Policy is an extension of client.Object. It adds methods that are common among all NGF Policies.
//
//counterfeiter:generate . Policy
type Policy interface {
	GetTargetRef() v1alpha2.LocalPolicyTargetReference
	GetPolicyStatus() v1alpha2.PolicyStatus
	SetPolicyStatus(status v1alpha2.PolicyStatus)
	client.Object
}

// ConfigGenerator generates a slice of bytes containing the configuration from a Policy.
//
//counterfeiter:generate . ConfigGenerator
type ConfigGenerator interface {
	Generate(policy Policy) []byte
}

// We generate a mock of ObjectKind so that we can create fake policies and set their GVKs.
//counterfeiter:generate k8s.io/apimachinery/pkg/runtime/schema.ObjectKind
