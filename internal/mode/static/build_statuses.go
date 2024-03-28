package static

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status2"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

type nginxReloadResult struct {
	error error
}

func buildRouteStatuses(
	routes map[types.NamespacedName]*graph.Route,
	nginxReloadRes nginxReloadResult,
) []status2.UpdateRequest {
	reqs := make([]status2.UpdateRequest, 0, len(routes))

	for nsname, r := range routes {
		parents := make([]v1.RouteParentStatus, 0, len(r.ParentRefs))

		defaultConds := staticConds.NewDefaultRouteConditions()

		for _, ref := range r.ParentRefs {
			failedAttachmentCondCount := 0
			if ref.Attachment != nil && !ref.Attachment.Attached {
				failedAttachmentCondCount = 1
			}
			allConds := make([]conditions.Condition, 0, len(r.Conditions)+len(defaultConds)+failedAttachmentCondCount)

			// We add defaultConds first, so that any additional conditions will override them, which is
			// ensured by DeduplicateConditions.
			allConds = append(allConds, defaultConds...)
			allConds = append(allConds, r.Conditions...)
			if failedAttachmentCondCount == 1 {
				allConds = append(allConds, ref.Attachment.FailedCondition)
			}

			if nginxReloadRes.error != nil {
				allConds = append(
					allConds,
					staticConds.NewRouteGatewayNotProgrammed(staticConds.RouteMessageFailedNginxReload),
				)
			}

			routeRef := r.Source.Spec.ParentRefs[ref.Idx]

			conds := conditions.DeduplicateConditions(allConds)

			apiConds := status2.ConvertConditions(conds, r.Source.Generation, metav1.Now())

			ps := v1.RouteParentStatus{
				ParentRef: v1.ParentReference{
					Namespace:   (*v1.Namespace)(&ref.Gateway.Namespace),
					Name:        v1.ObjectName(ref.Gateway.Name),
					SectionName: routeRef.SectionName,
				},
				ControllerName: v1.GatewayController("todo.controller/path"),
				Conditions:     apiConds,
			}

			parents = append(parents, ps)
		}

		status := v1.HTTPRouteStatus{
			RouteStatus: v1.RouteStatus{
				Parents: parents,
			},
		}

		req := status2.UpdateRequest{
			NsName:       nsname,
			ResourceType: &v1.HTTPRoute{},
			Setter:       newHTTPRouteStatusSetter(status, "todo.controller/path"),
		}

		reqs = append(reqs, req)
	}

	return reqs
}

func buildGatewayClassStatuses(
	gc *graph.GatewayClass,
	ignoredGwClasses map[types.NamespacedName]*v1.GatewayClass,
) []status2.UpdateRequest {
	var reqs []status2.UpdateRequest

	if gc != nil {
		defaultConds := conditions.NewDefaultGatewayClassConditions()

		conds := make([]conditions.Condition, 0, len(gc.Conditions)+len(defaultConds))

		// We add default conds first, so that any additional conditions will override them, which is
		// ensured by DeduplicateConditions.
		conds = append(conds, defaultConds...)
		conds = append(conds, gc.Conditions...)

		conds = conditions.DeduplicateConditions(conds)

		req := status2.UpdateRequest{
			NsName:       client.ObjectKeyFromObject(gc.Source),
			ResourceType: &v1.GatewayClass{},
			Setter:       newGatewayClassStatusSetter(gc.Source.Generation, conds),
		}

		reqs = append(reqs, req)
	}

	for nsname, gwClass := range ignoredGwClasses {
		req := status2.UpdateRequest{
			NsName:       nsname,
			ResourceType: &v1.GatewayClass{},
			Setter:       newGatewayClassStatusSetter(gwClass.Generation, []conditions.Condition{conditions.NewGatewayClassConflict()}),
		}

		reqs = append(reqs, req)
	}

	return reqs
}

func buildGatewayStatuses(
	gateway *graph.Gateway,
	ignoredGateways map[types.NamespacedName]*v1.Gateway,
	gwAddresses []v1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) []status2.UpdateRequest {
	reqs := make([]status2.UpdateRequest, 0, 1+len(ignoredGateways))

	if gateway != nil {
		reqs = append(reqs, buildGatewayStatus(gateway, gwAddresses, nginxReloadRes))
	}

	for nsname, gw := range ignoredGateways {
		apiConds := status2.ConvertConditions(staticConds.NewGatewayConflict(), gw.Generation, metav1.Now())
		reqs = append(reqs, status2.UpdateRequest{
			NsName:       nsname,
			ResourceType: &v1.Gateway{},
			Setter: newGatewayStatusSetter(v1.GatewayStatus{
				Conditions: apiConds,
			}),
		})
	}

	return reqs
}

func buildGatewayStatus(
	gateway *graph.Gateway,
	gwAddresses []v1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) status2.UpdateRequest {
	if !gateway.Valid {
		conds := status2.ConvertConditions(conditions.DeduplicateConditions(gateway.Conditions), gateway.Source.Generation, metav1.Now())
		return status2.UpdateRequest{
			NsName:       client.ObjectKeyFromObject(gateway.Source),
			ResourceType: &v1.Gateway{},
			Setter: newGatewayStatusSetter(v1.GatewayStatus{
				Conditions: conds,
			}),
		}
	}

	listenerStatuses := make([]v1.ListenerStatus, 0, len(gateway.Listeners))

	validListenerCount := 0
	for _, l := range gateway.Listeners {
		var conds []conditions.Condition

		if l.Valid {
			conds = staticConds.NewDefaultListenerConditions()
			validListenerCount++
		} else {
			conds = l.Conditions
		}

		if nginxReloadRes.error != nil {
			conds = append(
				conds,
				staticConds.NewListenerNotProgrammedInvalid(staticConds.ListenerMessageFailedNginxReload),
			)
		}

		apiConds := status2.ConvertConditions(conditions.DeduplicateConditions(conds), gateway.Source.Generation, metav1.Now())

		listenerStatuses = append(listenerStatuses, v1.ListenerStatus{
			Name:           v1.SectionName(l.Name),
			SupportedKinds: l.SupportedKinds,
			AttachedRoutes: int32(len(l.Routes)),
			Conditions:     apiConds,
		})
	}

	gwConds := staticConds.NewDefaultGatewayConditions()
	if validListenerCount == 0 {
		gwConds = append(gwConds, staticConds.NewGatewayNotAcceptedListenersNotValid()...)
	} else if validListenerCount < len(gateway.Listeners) {
		gwConds = append(gwConds, staticConds.NewGatewayAcceptedListenersNotValid())
	}

	if nginxReloadRes.error != nil {
		gwConds = append(
			gwConds,
			staticConds.NewGatewayNotProgrammedInvalid(staticConds.GatewayMessageFailedNginxReload),
		)
	}

	apiGwConds := status2.ConvertConditions(gwConds, gateway.Source.Generation, metav1.Now())

	return status2.UpdateRequest{
		NsName:       client.ObjectKeyFromObject(gateway.Source),
		ResourceType: &v1.Gateway{},
		Setter: newGatewayStatusSetter(v1.GatewayStatus{
			Listeners:  listenerStatuses,
			Conditions: apiGwConds,
			Addresses:  gwAddresses,
		}),
	}
}

func buildBackendTLSPolicyStatuses(backendTLSPolicies map[types.NamespacedName]*graph.BackendTLSPolicy,
) []status2.UpdateRequest {
	reqs := make([]status2.UpdateRequest, 0, len(backendTLSPolicies))

	for nsname, backendTLSPolicy := range backendTLSPolicies {
		if backendTLSPolicy.IsReferenced {
			if !backendTLSPolicy.Ignored {
				reqs = append(reqs, status2.UpdateRequest{
					NsName:       nsname,
					ResourceType: &v1alpha2.BackendTLSPolicy{},
					Setter:       newBackendTLSPolicyStatusSetter("todo.controller/path", backendTLSPolicy),
				})
			}
		}
	}
	return reqs
}
