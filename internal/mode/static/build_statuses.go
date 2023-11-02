package static

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
)

type nginxReloadResult struct {
	error error
}

// buildGatewayAPIStatuses builds status.Statuses from a Graph.
func buildGatewayAPIStatuses(
	graph *graph.Graph,
	gwAddresses []v1beta1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) status.GatewayAPIStatuses {
	statuses := status.GatewayAPIStatuses{
		HTTPRouteStatuses: make(status.HTTPRouteStatuses),
	}

	statuses.GatewayClassStatuses = buildGatewayClassStatuses(graph.GatewayClass, graph.IgnoredGatewayClasses)

	statuses.GatewayStatuses = buildGatewayStatuses(graph.Gateway, graph.IgnoredGateways, gwAddresses, nginxReloadRes)

	for nsname, r := range graph.Routes {
		parentStatuses := make([]status.ParentStatus, 0, len(r.ParentRefs))

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

			parentStatuses = append(parentStatuses, status.ParentStatus{
				GatewayNsName: ref.Gateway,
				SectionName:   routeRef.SectionName,
				Conditions:    staticConds.DeduplicateConditions(allConds),
			})
		}

		statuses.HTTPRouteStatuses[nsname] = status.HTTPRouteStatus{
			ObservedGeneration: r.Source.Generation,
			ParentStatuses:     parentStatuses,
		}
	}

	return statuses
}

func buildGatewayClassStatuses(
	gc *graph.GatewayClass,
	ignoredGwClasses map[types.NamespacedName]*v1beta1.GatewayClass,
) status.GatewayClassStatuses {
	statuses := make(status.GatewayClassStatuses)

	if gc != nil {
		defaultConds := conditions.NewDefaultGatewayClassConditions()

		conds := make([]conditions.Condition, 0, len(gc.Conditions)+len(defaultConds))

		// We add default conds first, so that any additional conditions will override them, which is
		// ensured by DeduplicateConditions.
		conds = append(conds, defaultConds...)
		conds = append(conds, gc.Conditions...)

		statuses[client.ObjectKeyFromObject(gc.Source)] = status.GatewayClassStatus{
			Conditions:         staticConds.DeduplicateConditions(conds),
			ObservedGeneration: gc.Source.Generation,
		}
	}

	for nsname, gwClass := range ignoredGwClasses {
		statuses[nsname] = status.GatewayClassStatus{
			Conditions:         []conditions.Condition{conditions.NewGatewayClassConflict()},
			ObservedGeneration: gwClass.Generation,
		}
	}

	return statuses
}

func buildGatewayStatuses(
	gateway *graph.Gateway,
	ignoredGateways map[types.NamespacedName]*v1beta1.Gateway,
	gwAddresses []v1beta1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) status.GatewayStatuses {
	statuses := make(status.GatewayStatuses)

	if gateway != nil {
		statuses[client.ObjectKeyFromObject(gateway.Source)] = buildGatewayStatus(gateway, gwAddresses, nginxReloadRes)
	}

	for nsname, gw := range ignoredGateways {
		statuses[nsname] = status.GatewayStatus{
			Conditions:         staticConds.NewGatewayConflict(),
			ObservedGeneration: gw.Generation,
			Ignored:            true,
		}
	}

	return statuses
}

func buildGatewayStatus(
	gateway *graph.Gateway,
	gwAddresses []v1beta1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) status.GatewayStatus {
	if !gateway.Valid {
		return status.GatewayStatus{
			Conditions:         staticConds.DeduplicateConditions(gateway.Conditions),
			ObservedGeneration: gateway.Source.Generation,
		}
	}

	listenerStatuses := make(map[string]status.ListenerStatus)

	validListenerCount := 0
	for name, l := range gateway.Listeners {
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

		listenerStatuses[name] = status.ListenerStatus{
			AttachedRoutes: int32(len(l.Routes)),
			Conditions:     staticConds.DeduplicateConditions(conds),
			SupportedKinds: l.SupportedKinds,
		}
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

	return status.GatewayStatus{
		Conditions:         staticConds.DeduplicateConditions(gwConds),
		ListenerStatuses:   listenerStatuses,
		Addresses:          gwAddresses,
		ObservedGeneration: gateway.Source.Generation,
	}
}

func buildNginxProxyStatus(np *ngfAPI.NginxProxy, nginxReloadRes nginxReloadResult) *status.NginxProxyStatus {
	if np == nil {
		return nil
	}

	conds := []conditions.Condition{
		staticConds.NewNginxProxyAccepted(),
	}

	if nginxReloadRes.error != nil {
		conds = append(conds, staticConds.NewNginxProxyNotProgrammed(staticConds.NginxProxyMessageFailedNginxReload))
	} else {
		conds = append(conds, staticConds.NewNginxProxyProgrammed())
	}

	npStatus := &status.NginxProxyStatus{
		NsName:             client.ObjectKeyFromObject(np),
		Conditions:         conds,
		ObservedGeneration: np.Generation,
	}

	return npStatus
}
