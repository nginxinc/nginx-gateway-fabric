package graph

import (
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

const maxAncestors = 16

func ancestorsFull(
	ancestors []v1alpha2.PolicyAncestorStatus,
	ctlrName string,
) bool {
	if len(ancestors) < maxAncestors {
		return false
	}

	for _, ancestor := range ancestors {
		if string(ancestor.ControllerName) == ctlrName {
			return false
		}
	}

	return true
}

func createParentReference(
	group v1.Group,
	kind v1.Kind,
	nsname types.NamespacedName,
) v1.ParentReference {
	return v1.ParentReference{
		Group:     &group,
		Kind:      &kind,
		Namespace: (*v1.Namespace)(&nsname.Namespace),
		Name:      v1.ObjectName(nsname.Name),
	}
}
