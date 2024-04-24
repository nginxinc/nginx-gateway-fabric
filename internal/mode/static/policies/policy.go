package policies

import (
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// Policy is an extension of client.Object. It adds methods that are common among all NGF Policies.
type Policy interface {
	GetTargetRef() v1alpha2.PolicyTargetReference
	GetPolicyStatus() v1alpha2.PolicyStatus
	SetPolicyStatus(status v1alpha2.PolicyStatus)
	client.Object
}
