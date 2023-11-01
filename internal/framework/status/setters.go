package status

import (
	"slices"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

// setter is a function that takes an object and sets the status on that object if the status has changed.
// If the status has not changed, and the setter does not set the status, it returns false.
type setter func(client.Object) bool

func newNginxGatewayStatusSetter(clock Clock, status NginxGatewayStatus) func(client.Object) bool {
	return func(object client.Object) bool {
		ng := object.(*ngfAPI.NginxGateway)
		conds := convertConditions(
			status.Conditions,
			status.ObservedGeneration,
			clock.Now(),
		)

		if conditionsEqual(ng.Status.Conditions, conds) {
			return false
		}

		ng.Status = ngfAPI.NginxGatewayStatus{
			Conditions: conds,
		}

		return true
	}
}

func newNginxProxyStatusSetter(clock Clock, status NginxProxyStatus) func(client.Object) bool {
	return func(object client.Object) bool {
		ng := object.(*ngfAPI.NginxProxy)
		conds := convertConditions(
			status.Conditions,
			status.ObservedGeneration,
			clock.Now(),
		)

		if conditionsEqual(ng.Status.Conditions, conds) {
			return false
		}

		ng.Status = ngfAPI.NginxProxyStatus{
			Conditions: conds,
		}

		return true
	}
}

func newGatewayClassStatusSetter(clock Clock, gcs GatewayClassStatus) func(client.Object) bool {
	return func(object client.Object) bool {
		gc := object.(*v1beta1.GatewayClass)
		status := prepareGatewayClassStatus(gcs, clock.Now())

		if conditionsEqual(gc.Status.Conditions, status.Conditions) {
			return false
		}

		gc.Status = status

		return true
	}
}

func newGatewayStatusSetter(clock Clock, gs GatewayStatus) func(client.Object) bool {
	return func(object client.Object) bool {
		gw := object.(*v1beta1.Gateway)
		status := prepareGatewayStatus(gs, clock.Now())

		if gwStatusEqual(gw.Status, status) {
			return false
		}

		gw.Status = status

		return true
	}
}

func newHTTPRouteStatusSetter(gatewayCtlrName string, clock Clock, rs HTTPRouteStatus) func(client.Object) bool {
	return func(object client.Object) bool {
		hr := object.(*v1beta1.HTTPRoute)
		status := prepareHTTPRouteStatus(
			rs,
			gatewayCtlrName,
			clock.Now(),
		)

		if hrStatusEqual(gatewayCtlrName, hr.Status, status) {
			return false
		}

		hr.Status = status

		return true
	}
}

func gwStatusEqual(prev, cur v1beta1.GatewayStatus) bool {
	addressesEqual := slices.EqualFunc(prev.Addresses, cur.Addresses, func(a1, a2 v1beta1.GatewayStatusAddress) bool {
		if !equalPointers[v1beta1.AddressType](a1.Type, a2.Type) {
			return false
		}

		return a1.Value == a2.Value
	})

	if !addressesEqual {
		return false
	}

	if !conditionsEqual(prev.Conditions, cur.Conditions) {
		return false
	}

	return slices.EqualFunc(prev.Listeners, cur.Listeners, func(s1, s2 v1beta1.ListenerStatus) bool {
		if s1.Name != s2.Name {
			return false
		}

		if s1.AttachedRoutes != s2.AttachedRoutes {
			return false
		}

		if !conditionsEqual(s1.Conditions, s2.Conditions) {
			return false
		}

		return slices.EqualFunc(s1.SupportedKinds, s2.SupportedKinds, func(k1, k2 v1beta1.RouteGroupKind) bool {
			if k1.Kind != k2.Kind {
				return false
			}

			return equalPointers(k1.Group, k2.Group)
		})
	})
}

func hrStatusEqual(gatewayCtlrName string, prev, cur v1beta1.HTTPRouteStatus) bool {
	// Since other controllers may update HTTPRoute status we can't assume anything about the order of the statuses,
	// and we have to ignore statuses written by other controllers when checking for equality.
	// Therefore, we can't use slices.EqualFunc here because it cares about the order.

	// First, we check if the prev status has any RouteParentStatuses that are no longer present in the cur status.
	for _, prevParent := range prev.Parents {
		if prevParent.ControllerName != v1beta1.GatewayController(gatewayCtlrName) {
			continue
		}

		exists := slices.ContainsFunc(cur.Parents, func(curParent v1beta1.RouteParentStatus) bool {
			return routeParentStatusEqual(prevParent, curParent)
		})

		if !exists {
			return false
		}
	}

	// Then, we check if the cur status has any RouteParentStatuses that are no longer present in the prev status.
	for _, curParent := range cur.Parents {
		exists := slices.ContainsFunc(prev.Parents, func(prevParent v1beta1.RouteParentStatus) bool {
			return routeParentStatusEqual(curParent, prevParent)
		})

		if !exists {
			return false
		}
	}

	return true
}

func routeParentStatusEqual(p1, p2 v1beta1.RouteParentStatus) bool {
	if p1.ControllerName != p2.ControllerName {
		return false
	}

	if p1.ParentRef.Name != p2.ParentRef.Name {
		return false
	}

	if !equalPointers(p1.ParentRef.Namespace, p2.ParentRef.Namespace) {
		return false
	}

	if !equalPointers(p1.ParentRef.SectionName, p2.ParentRef.SectionName) {
		return false
	}

	// we ignore the rest of the ParentRef fields because we do not set them

	return conditionsEqual(p1.Conditions, p2.Conditions)
}

func conditionsEqual(prev, cur []v1.Condition) bool {
	return slices.EqualFunc(prev, cur, func(c1, c2 v1.Condition) bool {
		if c1.ObservedGeneration != c2.ObservedGeneration {
			return false
		}

		if c1.Type != c2.Type {
			return false
		}

		if c1.Status != c2.Status {
			return false
		}

		if c1.Message != c2.Message {
			return false
		}

		return c1.Reason == c2.Reason
	})
}

// equalPointers returns whether two pointers are equal.
// Pointers are considered equal if one of the following is true:
// - They are both nil.
// - One is nil and the other is empty (e.g. nil string and "").
// - They are both non-nil, and their values are the same.
func equalPointers[T comparable](p1, p2 *T) bool {
	if p1 == nil && p2 == nil {
		return true
	}

	var p1Val, p2Val T

	if p1 != nil {
		p1Val = *p1
	}

	if p2 != nil {
		p2Val = *p2
	}

	return p1Val == p2Val
}
