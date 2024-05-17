package graph

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	ngfsort "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/sort"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

// Policy represents an NGF Policy.
type Policy struct {
	// Source is the corresponding Policy resource.
	Source policies.Policy
	// Ancestor is the ancestor object of the Policy. Used in status.
	Ancestor *PolicyAncestor
	// TargetRef is the resource that the Policy targets.
	TargetRef PolicyTargetRef
	// Conditions holds the conditions for the Policy.
	// These conditions apply to the entire Policy.
	// The conditions in the Ancestor apply only to the Policy in regard to the Ancestor.
	Conditions []conditions.Condition
	// Valid indicates whether the Policy is valid.
	Valid bool
}

// PolicyAncestor represents an ancestor of a Policy.
type PolicyAncestor struct {
	// Ancestor is the ancestor object.
	Ancestor v1.ParentReference
	// Conditions contains the list of conditions of the Policy in relation to the ancestor.
	Conditions []conditions.Condition
}

// PolicyTargetRef represents the object that the Policy is targeting.
type PolicyTargetRef struct {
	// Kind is the Kind of the object.
	Kind v1.Kind
	// Group is the Group of the object.
	Group v1.Group
	// Nsname is the NamespacedName of the object.
	Nsname types.NamespacedName
}

// PolicyKey is a unique identifier for an NGF Policy.
type PolicyKey struct {
	// Nsname is the NamespacedName of the Policy.
	NsName types.NamespacedName
	// GVK is the GroupVersionKind of the Policy.
	GVK schema.GroupVersionKind
}

const (
	gatewayGroupKind = v1.GroupName + "/" + kinds.Gateway
	hrGroupKind      = v1.GroupName + "/" + kinds.HTTPRoute
	grpcGroupKind    = v1.GroupName + "/" + kinds.GRPCRoute
)

// attachPolicies attaches the graph's processed policies to the resources they target. It modifies the graph in place.
func (g *Graph) attachPolicies(ctlrName string) {
	if g.Gateway == nil {
		return
	}

	for _, policy := range g.NGFPolicies {
		ref := policy.TargetRef

		switch ref.Kind {
		case kinds.Gateway:
			attachPolicyToGateway(policy, g.Gateway, g.IgnoredGateways, ctlrName)
		case kinds.HTTPRoute, kinds.GRPCRoute:
			route, exists := g.Routes[routeKeyForKind(ref.Kind, ref.Nsname)]
			if !exists {
				continue
			}

			attachPolicyToRoute(policy, route, ctlrName)
		}
	}
}

func attachPolicyToRoute(policy *Policy, route *L7Route, ctlrName string) {
	kind := v1.Kind(kinds.HTTPRoute)
	if route.RouteType == RouteTypeGRPC {
		kind = kinds.GRPCRoute
	}

	routeNsName := types.NamespacedName{Namespace: route.Source.GetNamespace(), Name: route.Source.GetName()}

	ancestor := &PolicyAncestor{
		Ancestor: createParentReference(v1.GroupName, kind, routeNsName),
	}

	curAncestorStatus := policy.Source.GetPolicyStatus().Ancestors
	if ancestorsFull(curAncestorStatus, ancestor.Ancestor, ctlrName) {
		// FIXME (kate-osborn): https://github.com/nginxinc/nginx-gateway-fabric/issues/1987
		return
	}

	policy.Ancestor = ancestor

	if !route.Valid || !route.Attachable || len(route.ParentRefs) == 0 {
		policy.Ancestor.Conditions = []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")}

		return
	}

	route.Policies = append(route.Policies, policy)
}

func attachPolicyToGateway(
	policy *Policy,
	gw *Gateway,
	ignoredGateways map[types.NamespacedName]*v1.Gateway,
	ctlrName string,
) {
	ref := policy.TargetRef

	_, ignored := ignoredGateways[ref.Nsname]

	if !ignored && ref.Nsname != client.ObjectKeyFromObject(gw.Source) {
		return
	}

	ancestor := &PolicyAncestor{
		Ancestor: createParentReference(v1.GroupName, kinds.Gateway, ref.Nsname),
	}

	curAncestorStatus := policy.Source.GetPolicyStatus().Ancestors
	if ancestorsFull(curAncestorStatus, ancestor.Ancestor, ctlrName) {
		// FIXME (kate-osborn): https://github.com/nginxinc/nginx-gateway-fabric/issues/1987
		return
	}

	policy.Ancestor = ancestor

	if ignored {
		policy.Ancestor.Conditions = []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is ignored")}
		return
	}

	if !gw.Valid {
		policy.Ancestor.Conditions = []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")}
		return
	}

	gw.Policies = append(gw.Policies, policy)
}

func processPolicies(
	policies map[PolicyKey]policies.Policy,
	validator validation.PolicyValidator,
	gateways processedGateways,
	routes map[RouteKey]*L7Route,
) map[PolicyKey]*Policy {
	if len(policies) == 0 || gateways.Winner == nil {
		return nil
	}

	processedPolicies := make(map[PolicyKey]*Policy)

	for key, policy := range policies {
		ref := policy.GetTargetRef()
		refNsName := types.NamespacedName{Name: string(ref.Name), Namespace: policy.GetNamespace()}

		refGroupKind := fmt.Sprintf("%s/%s", ref.Group, ref.Kind)

		switch refGroupKind {
		case gatewayGroupKind:
			if !gatewayExists(refNsName, gateways.Winner, gateways.Ignored) {
				continue
			}
		case hrGroupKind, grpcGroupKind:
			if _, exists := routes[routeKeyForKind(ref.Kind, refNsName)]; !exists {
				continue
			}
		default:
			continue
		}

		var conds []conditions.Condition

		if err := validator.Validate(policy); err != nil {
			conds = append(conds, staticConds.NewPolicyInvalid(err.Error()))
		}

		processedPolicies[key] = &Policy{
			Source:     policy,
			Valid:      len(conds) == 0,
			Conditions: conds,
			TargetRef: PolicyTargetRef{
				Kind:   ref.Kind,
				Group:  ref.Group,
				Nsname: refNsName,
			},
		}
	}

	markConflictedPolicies(processedPolicies, validator)

	return processedPolicies
}

// markConflictedPolicies marks policies that conflict with a policy of greater precedence as invalid.
// Policies are sorted by timestamp and then alphabetically.
func markConflictedPolicies(policies map[PolicyKey]*Policy, validator validation.PolicyValidator) {
	// Policies can only conflict if they are the same policy type (gvk) and they target the same resource.
	type key struct {
		policyGVK schema.GroupVersionKind
		PolicyTargetRef
	}

	possibles := make(map[key][]*Policy)

	for pk, p := range policies {
		// If a policy is invalid, it cannot conflict with another policy.
		if p.Valid {
			ak := key{
				PolicyTargetRef: p.TargetRef,
				policyGVK:       pk.GVK,
			}
			if possibles[ak] == nil {
				possibles[ak] = make([]*Policy, 0)
			}
			possibles[ak] = append(possibles[ak], p)
		}
	}

	for _, policyList := range possibles {
		if len(policyList) == 1 {
			// if the policyList only has one entry, then we don't need to check for conflicts.
			continue
		}

		// First, we sort the policyList according to the rules in the spec.
		// This will put them in priority-order.
		sort.Slice(
			policyList, func(i, j int) bool {
				return ngfsort.ClientObject(policyList[i].Source, policyList[j].Source)
			},
		)

		// Second, we range over the policyList, starting with the highest priority policy.
		for i := range policyList {
			if !policyList[i].Valid {
				// Ignore policy that has already been marked as invalid.
				continue
			}

			// Next, we compare the ith policy (policyList[i]) to the rest of the policies in the list.
			// The ith policy takes precedence over polices that follow it, so if there is a conflict between
			// it and a subsequent policy, the ith policy wins, and we mark the subsequent policy as invalid.
			// Example: policyList = [A, B, C] where B conflicts with A.
			// i=A, j=B => conflict, B's marked as invalid.
			// i=A, j=C => no conflict.
			// i=B, j=C => B's already invalid, so we hit the continue.
			// i=C => j loop terminates.
			// Results: A, and C are valid. B is invalid.
			for j := i + 1; j < len(policyList); j++ {
				if !policyList[j].Valid {
					// Ignore policy that has already been marked as invalid.
					continue
				}

				if validator.Conflicts(policyList[i].Source, policyList[j].Source) {
					conflicted := policyList[j]
					conflicted.Valid = false
					conflicted.Conditions = append(conflicted.Conditions, staticConds.NewPolicyConflicted(
						fmt.Sprintf(
							"Conflicts with another %s",
							conflicted.Source.GetObjectKind().GroupVersionKind().Kind,
						),
					))
				}
			}
		}
	}
}
