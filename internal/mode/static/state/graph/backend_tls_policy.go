package graph

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

type BackendTLSPolicy struct {
	// Source is the source resource.
	Source *v1alpha2.BackendTLSPolicy
	// CaCertRef is the name of the ConfigMap that contains the CA certificate.
	CaCertRef types.NamespacedName
	// Gateway is the name of the Gateway that is being checked for this BackendTLSPolicy.
	Gateway types.NamespacedName
	// Conditions include Conditions for the BackendTLSPolicy.
	Conditions []conditions.Condition
	// Valid shows whether the BackendTLSPolicy is valid.
	Valid bool
	// IsReferenced shows whether the BackendTLSPolicy is referenced by a BackendRef.
	IsReferenced bool
}

func processBackendTLSPolicies(
	backendTLSPolicies map[types.NamespacedName]*v1alpha2.BackendTLSPolicy,
	configMapResolver *configMapResolver,
	ctlrName string,
	gateway *Gateway,
) map[types.NamespacedName]*BackendTLSPolicy {
	if len(backendTLSPolicies) == 0 {
		return nil
	}
	processedBackendTLSPolicies := make(map[types.NamespacedName]*BackendTLSPolicy, len(backendTLSPolicies))
	for nsname, backendTLSPolicy := range backendTLSPolicies {
		processedBackendTLSPolicies[nsname] = &BackendTLSPolicy{
			Source: backendTLSPolicy,
		}
		valid, caCertRef, conds := validateBackendTLSPolicy(backendTLSPolicy, configMapResolver, ctlrName, gateway)
		processedBackendTLSPolicies[nsname].Valid = valid
		processedBackendTLSPolicies[nsname].Conditions = conds
		processedBackendTLSPolicies[nsname].Gateway = types.NamespacedName{
			Namespace: gateway.Source.Namespace,
			Name:      gateway.Source.Name,
		}
		processedBackendTLSPolicies[nsname].CaCertRef = caCertRef
	}
	return processedBackendTLSPolicies
}

func validateBackendTLSPolicy(
	backendTLSPolicy *v1alpha2.BackendTLSPolicy,
	configMapResolver *configMapResolver,
	ctlrName string,
	gateway *Gateway,
) (bool, types.NamespacedName, []conditions.Condition) {
	conds := make([]conditions.Condition, 0)
	valid := true
	caCertRef := types.NamespacedName{}
	if len(backendTLSPolicy.Status.Ancestors) >= 16 {
		// check if we already are an ancestor on this policy. If we are, we are safe to continue.
		ancestorRef := v1.ParentReference{
			Namespace: helpers.GetPointer((v1.Namespace)(gateway.Source.Namespace)),
			Name:      v1.ObjectName(gateway.Source.Name),
		}
		var alreadyAncestor bool
		for _, ancestor := range backendTLSPolicy.Status.Ancestors {
			if string(ancestor.ControllerName) == ctlrName && ancestor.AncestorRef.Name == ancestorRef.Name &&
				*ancestor.AncestorRef.Namespace == *ancestorRef.Namespace {
				alreadyAncestor = true
				break
			}
		}
		if !alreadyAncestor {
			valid = false
			conds = append(conds, staticConds.NewBackendTLSPolicyIgnored("too many ancestors, cannot attach a new Gateway"))
		}
	}
	if err := validateBackendTLSHostname(backendTLSPolicy); err != nil {
		valid = false
		conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(fmt.Sprintf("invalid hostname: %s", err.Error())))
	}
	if backendTLSPolicy.Spec.TLS.CACertRefs != nil && len(backendTLSPolicy.Spec.TLS.CACertRefs) > 0 {
		if err := validateBackendTLSCACertRef(backendTLSPolicy, configMapResolver); err != nil {
			valid = false
			conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(
				fmt.Sprintf("invalid CACertRef: %s", err.Error())))
		} else {
			caCertRef = types.NamespacedName{
				Namespace: backendTLSPolicy.Namespace, Name: string(backendTLSPolicy.Spec.TLS.CACertRefs[0].Name),
			}
		}
	} else if backendTLSPolicy.Spec.TLS.WellKnownCACerts != nil {
		if err := validateBackendTLSWellKnownCACerts(backendTLSPolicy); err != nil {
			valid = false
			conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(
				fmt.Sprintf("invalid WellKnownCACerts: %s", err.Error())))
		}
	} else {
		valid = false
		conds = append(conds, staticConds.NewBackendTLSPolicyInvalid("CACertRefs and WellKnownCACerts are both nil"))
	}
	return valid, caCertRef, conds
}

func validateBackendTLSHostname(btp *v1alpha2.BackendTLSPolicy) error {
	h := string(btp.Spec.TLS.Hostname)

	if err := validateHostname(h); err != nil {
		path := field.NewPath("tls.hostname")
		valErr := field.Invalid(path, btp.Spec.TLS.Hostname, err.Error())
		return valErr
	}
	return nil
}

func validateBackendTLSCACertRef(btp *v1alpha2.BackendTLSPolicy, configMapResolver *configMapResolver) error {
	if len(btp.Spec.TLS.CACertRefs) != 1 {
		path := field.NewPath("tls.cacertrefs")
		valErr := field.TooMany(path, len(btp.Spec.TLS.CACertRefs), 1)
		return valErr
	}
	if btp.Spec.TLS.CACertRefs[0].Kind != "ConfigMap" {
		path := field.NewPath("tls.cacertrefs[0].kind")
		valErr := field.NotSupported(path, btp.Spec.TLS.CACertRefs[0].Kind, []string{"ConfigMap"})
		return valErr
	}
	if btp.Spec.TLS.CACertRefs[0].Group != "" && btp.Spec.TLS.CACertRefs[0].Group != "core" {
		path := field.NewPath("tls.cacertrefs[0].group")
		valErr := field.NotSupported(path, btp.Spec.TLS.CACertRefs[0].Group, []string{"", "core"})
		return valErr
	}
	nsName := types.NamespacedName{Namespace: btp.Namespace, Name: string(btp.Spec.TLS.CACertRefs[0].Name)}
	if err := configMapResolver.resolve(nsName); err != nil {
		path := field.NewPath("tls.cacertrefs[0]")
		return field.Invalid(path, btp.Spec.TLS.CACertRefs[0], err.Error())
	}
	return nil
}

func validateBackendTLSWellKnownCACerts(btp *v1alpha2.BackendTLSPolicy) error {
	if *btp.Spec.TLS.WellKnownCACerts != v1alpha2.WellKnownCACertSystem {
		path := field.NewPath("tls.wellknowncacerts")
		return field.Invalid(path, btp.Spec.TLS.WellKnownCACerts, "unsupported value")
	}
	return nil
}
