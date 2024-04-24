package graph

import (
	"fmt"
	"sort"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	ngfsort "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/sort"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

// Policy represents an NGF Policy.
type Policy struct {
	// Source is the NGF Policy object.
	Source policies.Policy
	// TargetRef is the TargetRef of the Policy.
	TargetRef PolicyTargetRef
	// Valid indicates whether the Policy is semantically and syntactically valid.
	Valid bool
	// Conditions contains the conditions of the Policy. Conditions will be nil if the Policy is valid.
	Conditions []conditions.Condition
	// Ancestors is a list of the ancestors of the Policies.
	Ancestors []PolicyAncestor
}

// PolicyAncestor represents an ancestor of a Policy.
type PolicyAncestor struct {
	// Ancestor is the ancestor object.
	Ancestor PolicyAncestorRef
	// Conditions contains the list of conditions of the Policy in relation to the ancestor.
	Conditions []conditions.Condition
}

// PolicyAncestorRef contains the identifying information of the ancestor.
type PolicyAncestorRef struct {
	// Kind is the Kind of the object.
	Kind v1.Kind
	// Group is the Group of the object.
	Group v1.Group
	// Nsname is the NamespacedName of the object.
	Nsname types.NamespacedName
	// SectionName is the SectionName of the object. For example, a listener name. This may be empty.
	SectionName string
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
	gatewayGroupKind = v1.GroupName + "/" + "Gateway"
	hrGroupKind      = v1.GroupName + "/" + "HTTPRoute"
)

// attachPolicies attaches the graph's processed policies to the resources they target. It modifies the graph in place.
func (g *Graph) attachPolicies() {
	if g.Gateway == nil {
		return
	}

	for _, policy := range g.NGFPolicies {
		ref := policy.TargetRef

		switch ref.Kind {
		case "Gateway":
			ancestor, attached := attachPolicyToGateway(policy, g.Gateway, g.IgnoredGateways)
			if attached {
				if g.Gateway.Policies == nil {
					g.Gateway.Policies = make([]*Policy, 0, len(g.NGFPolicies))
				}

				g.Gateway.Policies = append(g.Gateway.Policies, policy)
			}

			policy.Ancestors = []PolicyAncestor{ancestor}
		case "HTTPRoute":
			route, exists := g.Routes[ref.Nsname]
			if !exists {
				return
			}

			ancestors, attached := attachPolicyToRoute(policy, route)

			if attached {
				if route.Policies == nil {
					route.Policies = make([]*Policy, 0, len(g.NGFPolicies))
				}

				route.Policies = append(route.Policies, policy)
			}

			policy.Ancestors = ancestors
		}
	}
}

func attachPolicyToRoute(policy *Policy, route *Route) (ancestors []PolicyAncestor, attached bool) {
	ancestors = make([]PolicyAncestor, 0, len(route.ParentRefs))

	if len(route.ParentRefs) == 0 {
		// this is an edge case that only happens when there are duplicate section names in the route
		routeNsName := types.NamespacedName{Namespace: route.Source.Namespace, Name: route.Source.Name}
		ancestors = append(ancestors, PolicyAncestor{
			Ancestor: PolicyAncestorRef{
				Kind:   "HTTPRoute",
				Group:  v1.GroupName,
				Nsname: routeNsName,
			},
			Conditions: []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")},
		})

		return ancestors, false
	}

	for _, pr := range route.ParentRefs {
		ancestor := PolicyAncestor{
			Ancestor: PolicyAncestorRef{
				Kind:        "Gateway",
				Group:       v1.GroupName,
				Nsname:      pr.Gateway,
				SectionName: pr.SectionName,
			},
			Conditions: make([]conditions.Condition, 0, 1),
		}

		if !parentRefAttached(route, pr) {
			ancestor.Conditions = append(
				ancestor.Conditions,
				staticConds.NewPolicyTargetNotFound("TargetRef is invalid"),
			)
		} else if policy.Valid {
			ancestor.Conditions = append(ancestor.Conditions, staticConds.NewPolicyAccepted())
			attached = true
		}

		ancestors = append(ancestors, ancestor)
	}

	return ancestors, attached
}

func parentRefAttached(route *Route, parent ParentRef) bool {
	return route.Valid && parent.Attachment != nil && parent.Attachment.Attached
}

func attachPolicyToGateway(
	policy *Policy,
	gw *Gateway,
	ignoredGateways map[types.NamespacedName]*v1.Gateway,
) (ancestor PolicyAncestor, attached bool) {
	ref := policy.TargetRef

	_, ignored := ignoredGateways[ref.Nsname]

	if !ignored && ref.Nsname != client.ObjectKeyFromObject(gw.Source) {
		return PolicyAncestor{}, false
	}

	ancestor = PolicyAncestor{
		Ancestor: PolicyAncestorRef{
			Kind:   "Gateway",
			Group:  v1.GroupName,
			Nsname: ref.Nsname,
		},
		Conditions: make([]conditions.Condition, 0),
	}

	if ignored {
		ancestor.Conditions = append(ancestor.Conditions, staticConds.NewPolicyTargetNotFound("TargetRef is ignored"))

		return ancestor, false
	}

	if !gw.Valid {
		ancestor.Conditions = append(ancestor.Conditions, staticConds.NewPolicyTargetNotFound("TargetRef is invalid"))

		return ancestor, false
	}

	if !policy.Valid {
		return ancestor, false
	}

	ancestor.Conditions = append(ancestor.Conditions, staticConds.NewPolicyAccepted())

	return ancestor, true
}

func processPolicies(
	policies map[PolicyKey]policies.Policy,
	validator validation.PolicyValidator,
	gateways processedGateways,
	routes map[types.NamespacedName]*Route,
) map[PolicyKey]*Policy {
	if len(policies) == 0 || gateways.Winner == nil {
		return nil
	}

	processedPolicies := make(map[PolicyKey]*Policy)

	for key, policy := range policies {
		ref := policy.GetTargetRef()
		refNsName := targetRefNsName(ref, policy.GetNamespace())

		refGroupKind := fmt.Sprintf("%s/%s", ref.Group, ref.Kind)

		switch refGroupKind {
		case gatewayGroupKind:
			if !gatewayExists(refNsName, gateways.Winner, gateways.Ignored) {
				continue
			}
		case hrGroupKind:
			if _, exists := routes[refNsName]; !exists {
				continue
			}
		}

		conds := make([]conditions.Condition, 0, 2)

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
	type key struct {
		policyGVK schema.GroupVersionKind
		PolicyTargetRef
	}

	possibles := make(map[key][]*Policy)

	for pk, p := range policies {
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
		if len(policyList) > 1 {
			sort.SliceStable(
				policyList, func(i, j int) bool {
					return ngfsort.ClientObject(policyList[i].Source, policyList[j].Source)
				},
			)

			for i := range policyList {
				if !policyList[i].Valid {
					continue
				}

				for j := i + 1; j < len(policyList); j++ {
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
}
