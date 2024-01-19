package static

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

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
	gwAddresses []v1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) status.GatewayAPIStatuses {
	statuses := status.GatewayAPIStatuses{
		HTTPRouteStatuses: make(status.HTTPRouteStatuses),
	}

	statuses.GatewayClassStatuses = buildGatewayClassStatuses(graph.GatewayClass, graph.IgnoredGatewayClasses)

	statuses.GatewayStatuses = buildGatewayStatuses(graph.Gateway, graph.IgnoredGateways, gwAddresses, nginxReloadRes)

	statuses.BackendTLSPolicyStatuses = buildBackendTLSPolicyStatuses(graph.BackendTLSPolicies)

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
				Conditions:    conditions.DeduplicateConditions(allConds),
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
	ignoredGwClasses map[types.NamespacedName]*v1.GatewayClass,
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
			Conditions:         conditions.DeduplicateConditions(conds),
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
	ignoredGateways map[types.NamespacedName]*v1.Gateway,
	gwAddresses []v1.GatewayStatusAddress,
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
	gwAddresses []v1.GatewayStatusAddress,
	nginxReloadRes nginxReloadResult,
) status.GatewayStatus {
	if !gateway.Valid {
		return status.GatewayStatus{
			Conditions:         conditions.DeduplicateConditions(gateway.Conditions),
			ObservedGeneration: gateway.Source.Generation,
		}
	}

	listenerStatuses := make([]status.ListenerStatus, 0, len(gateway.Listeners))

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

		listenerStatuses = append(listenerStatuses, status.ListenerStatus{
			Name:           v1.SectionName(l.Name),
			AttachedRoutes: int32(len(l.Routes)),
			Conditions:     conditions.DeduplicateConditions(conds),
			SupportedKinds: l.SupportedKinds,
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

	return status.GatewayStatus{
		Conditions:         conditions.DeduplicateConditions(gwConds),
		ListenerStatuses:   listenerStatuses,
		Addresses:          gwAddresses,
		ObservedGeneration: gateway.Source.Generation,
	}
}

func buildBackendTLSPolicyStatuses(backendTLSPolicies map[types.NamespacedName]*graph.BackendTLSPolicy,
) status.BackendTLSPolicyStatuses {
	statuses := make(status.BackendTLSPolicyStatuses, len(backendTLSPolicies))
	ignoreStatus := false

	for nsname, backendTLSPolicy := range backendTLSPolicies {
		if backendTLSPolicy.IsReferenced {
			if !backendTLSPolicy.Valid {
				for i := range backendTLSPolicy.Conditions {
					if backendTLSPolicy.Conditions[i].Reason == string(staticConds.BackendTLSPolicyReasonIgnored) {
						// We should not report the status of an ignored BackendTLSPolicy.
						ignoreStatus = true
					}
				}
			}
			if !ignoreStatus {
				statuses[nsname] = status.BackendTLSPolicyStatus{
					AncestorStatuses: []status.AncestorStatus{
						{
							GatewayNsName: backendTLSPolicy.Gateway,
							Conditions:    conditions.DeduplicateConditions(backendTLSPolicy.Conditions),
						},
					},
				}
			}
		}
	}
	return statuses
}
