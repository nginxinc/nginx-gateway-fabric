package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func refGrantAllowsGatewayToSecret(
	refGrants map[types.NamespacedName]*v1beta1.ReferenceGrant,
	gwNs string,
	secretNsName types.NamespacedName,
) bool {
	for nsname, grant := range refGrants {
		if nsname.Namespace != secretNsName.Namespace {
			continue
		}

		if fromIncludesGatewayNs(grant.Spec.From, gwNs) && toIncludesSecret(grant.Spec.To, secretNsName.Name) {
			return true
		}
	}

	return false
}

func fromIncludesGatewayNs(fromList []v1beta1.ReferenceGrantFrom, gwNs string) bool {
	for _, from := range fromList {
		if from.Group != v1beta1.GroupName {
			continue
		}

		if from.Kind != "Gateway" {
			continue
		}

		if string(from.Namespace) != gwNs {
			continue
		}

		return true
	}

	return false
}

func toIncludesSecret(toList []v1beta1.ReferenceGrantTo, secretName string) bool {
	for _, to := range toList {
		if to.Group != "" && to.Group != "core" {
			continue
		}

		if to.Kind != "Secret" {
			continue
		}

		if to.Name != nil && string(*to.Name) != secretName {
			continue
		}

		return true
	}

	return false
}
