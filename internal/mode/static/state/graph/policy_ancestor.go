package graph

import (
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func ancestorsFull(
	ancestors []v1alpha2.PolicyAncestorStatus,
	newAncestor v1alpha2.ParentReference,
	ctlrName string,
) bool {
	if len(ancestors) < 16 {
		return false
	}

	status := v1alpha2.PolicyAncestorStatus{AncestorRef: newAncestor, ControllerName: v1.GatewayController(ctlrName)}

	for _, ancestor := range ancestors {
		if ancestorStatusEqual(ancestor, status) {
			return false
		}
	}

	return true
}

func ancestorStatusEqual(curStatus v1alpha2.PolicyAncestorStatus, newStatus v1alpha2.PolicyAncestorStatus) bool {
	if curStatus.ControllerName != newStatus.ControllerName {
		return false
	}

	if !helpers.EqualPointers(curStatus.AncestorRef.Group, newStatus.AncestorRef.Group) {
		return false
	}

	if !helpers.EqualPointers(curStatus.AncestorRef.Kind, newStatus.AncestorRef.Kind) {
		return false
	}

	if curStatus.AncestorRef.Name != newStatus.AncestorRef.Name {
		return false
	}

	if !helpers.EqualPointers(curStatus.AncestorRef.Namespace, newStatus.AncestorRef.Namespace) {
		return false
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
