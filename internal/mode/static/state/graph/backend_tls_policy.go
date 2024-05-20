package graph

import (
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/validation/field"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha3"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

type BackendTLSPolicy struct {
	// Source is the source resource.
	Source *v1alpha3.BackendTLSPolicy
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
	// Ignored shows whether the BackendTLSPolicy is ignored.
	Ignored bool
}

func processBackendTLSPolicies(
	backendTLSPolicies map[types.NamespacedName]*v1alpha3.BackendTLSPolicy,
	configMapResolver *configMapResolver,
	ctlrName string,
	gateway *Gateway,
) map[types.NamespacedName]*BackendTLSPolicy {
	if len(backendTLSPolicies) == 0 || gateway == nil {
		return nil
	}

	processedBackendTLSPolicies := make(map[types.NamespacedName]*BackendTLSPolicy, len(backendTLSPolicies))
	for nsname, backendTLSPolicy := range backendTLSPolicies {
		var caCertRef types.NamespacedName
		valid, ignored, conds := validateBackendTLSPolicy(
			backendTLSPolicy,
			configMapResolver,
			ctlrName,
			gateway,
		)

		if valid && !ignored && backendTLSPolicy.Spec.Validation.CACertificateRefs != nil {
			caCertRef = types.NamespacedName{
				Namespace: backendTLSPolicy.Namespace, Name: string(backendTLSPolicy.Spec.Validation.CACertificateRefs[0].Name),
			}
		}

		processedBackendTLSPolicies[nsname] = &BackendTLSPolicy{
			Source:     backendTLSPolicy,
			Valid:      valid,
			Conditions: conds,
			Gateway: types.NamespacedName{
				Namespace: gateway.Source.Namespace,
				Name:      gateway.Source.Name,
			},
			CaCertRef: caCertRef,
			Ignored:   ignored,
		}
	}
	return processedBackendTLSPolicies
}

func validateBackendTLSPolicy(
	backendTLSPolicy *v1alpha3.BackendTLSPolicy,
	configMapResolver *configMapResolver,
	ctlrName string,
	gateway *Gateway,
) (valid, ignored bool, conds []conditions.Condition) {
	valid = true
	ignored = false
	if err := validateAncestorMaxCount(backendTLSPolicy, ctlrName, gateway); err != nil {
		valid = false
		ignored = true
	}
	if err := validateBackendTLSHostname(backendTLSPolicy); err != nil {
		valid = false
		conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(fmt.Sprintf("invalid hostname: %s", err.Error())))
	}

	caCertRefs := backendTLSPolicy.Spec.Validation.CACertificateRefs
	wellKnownCerts := backendTLSPolicy.Spec.Validation.WellKnownCACertificates
	if len(caCertRefs) > 0 && wellKnownCerts != nil {
		valid = false
		msg := "CACertificateRefs and WellKnownCACertificates are mutually exclusive"
		conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(msg))
	} else if len(caCertRefs) > 0 {
		if err := validateBackendTLSCACertRef(backendTLSPolicy, configMapResolver); err != nil {
			valid = false
			conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(
				fmt.Sprintf("invalid CACertificateRef: %s", err.Error())))
		}
	} else if wellKnownCerts != nil {
		if err := validateBackendTLSWellKnownCACerts(backendTLSPolicy); err != nil {
			valid = false
			conds = append(conds, staticConds.NewBackendTLSPolicyInvalid(
				fmt.Sprintf("invalid WellKnownCACertificates: %s", err.Error())))
		}
	} else {
		valid = false
		conds = append(conds, staticConds.NewBackendTLSPolicyInvalid("CACertRefs and WellKnownCACerts are both nil"))
	}
	return valid, ignored, conds
}

func validateAncestorMaxCount(backendTLSPolicy *v1alpha3.BackendTLSPolicy, ctlrName string, gateway *Gateway) error {
	var err error
	if len(backendTLSPolicy.Status.Ancestors) >= 16 {
		// check if we already are an ancestor on this policy. If we are, we are safe to continue.
		ancestorRef := v1.ParentReference{
			Namespace: helpers.GetPointer((v1.Namespace)(gateway.Source.Namespace)),
			Name:      v1.ObjectName(gateway.Source.Name),
		}
		var alreadyAncestor bool
		for _, ancestor := range backendTLSPolicy.Status.Ancestors {
			if string(ancestor.ControllerName) == ctlrName && ancestor.AncestorRef.Name == ancestorRef.Name &&
				ancestor.AncestorRef.Namespace != nil && *ancestor.AncestorRef.Namespace == *ancestorRef.Namespace {
				alreadyAncestor = true
				break
			}
		}
		if !alreadyAncestor {
			err = errors.New("too many ancestors, cannot attach a new Gateway")
		}
	}
	return err
}

func validateBackendTLSHostname(btp *v1alpha3.BackendTLSPolicy) error {
	h := string(btp.Spec.Validation.Hostname)

	if err := validateHostname(h); err != nil {
		path := field.NewPath("tls.hostname")
		valErr := field.Invalid(path, btp.Spec.Validation.Hostname, err.Error())
		return valErr
	}
	return nil
}

func validateBackendTLSCACertRef(btp *v1alpha3.BackendTLSPolicy, configMapResolver *configMapResolver) error {
	if len(btp.Spec.Validation.CACertificateRefs) != 1 {
		path := field.NewPath("tls.cacertrefs")
		valErr := field.TooMany(path, len(btp.Spec.Validation.CACertificateRefs), 1)
		return valErr
	}
	if btp.Spec.Validation.CACertificateRefs[0].Kind != "ConfigMap" {
		path := field.NewPath("tls.cacertrefs[0].kind")
		valErr := field.NotSupported(path, btp.Spec.Validation.CACertificateRefs[0].Kind, []string{"ConfigMap"})
		return valErr
	}
	if btp.Spec.Validation.CACertificateRefs[0].Group != "" && btp.Spec.Validation.CACertificateRefs[0].Group != "core" {
		path := field.NewPath("tls.cacertrefs[0].group")
		valErr := field.NotSupported(path, btp.Spec.Validation.CACertificateRefs[0].Group, []string{"", "core"})
		return valErr
	}
	nsName := types.NamespacedName{Namespace: btp.Namespace, Name: string(btp.Spec.Validation.CACertificateRefs[0].Name)}
	if err := configMapResolver.resolve(nsName); err != nil {
		path := field.NewPath("tls.cacertrefs[0]")
		return field.Invalid(path, btp.Spec.Validation.CACertificateRefs[0], err.Error())
	}
	return nil
}

func validateBackendTLSWellKnownCACerts(btp *v1alpha3.BackendTLSPolicy) error {
	if *btp.Spec.Validation.WellKnownCACertificates != v1alpha3.WellKnownCACertificatesSystem {
		path := field.NewPath("tls.wellknowncacertificates")
		return field.NotSupported(
			path,
			btp.Spec.Validation.WellKnownCACertificates,
			[]string{string(v1alpha3.WellKnownCACertificatesSystem)},
		)
	}
	return nil
}
