package v1alpha1

import (
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// FIXME(kate-osborn): Figure out a way to generate these methods for all our policies.
// These methods implement the policies.Policy interface which extends client.Object to add the following methods.

func (p *ClientSettingsPolicy) GetTargetRef() v1alpha2.PolicyTargetReference {
	return p.Spec.TargetRef
}

func (p *ClientSettingsPolicy) GetPolicyStatus() v1alpha2.PolicyStatus {
	return p.Status
}

func (p *ClientSettingsPolicy) SetPolicyStatus(status v1alpha2.PolicyStatus) {
	p.Status = status
}
