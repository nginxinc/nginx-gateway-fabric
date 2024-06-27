package graph

import (
	"slices"
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/policiesfakes"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/validation"
)

var testNs = "test"

func TestAttachPolicies(t *testing.T) {
	policyGVK := schema.GroupVersionKind{Group: "Group", Version: "Version", Kind: "Policy"}

	gwPolicyKey := createTestPolicyKey(policyGVK, "gw-policy")
	gwPolicy := &Policy{
		Valid:  true,
		Source: &policiesfakes.FakePolicy{},
		TargetRefs: []PolicyTargetRef{
			{
				Kind:   kinds.Gateway,
				Group:  v1.GroupName,
				Nsname: types.NamespacedName{Namespace: testNs, Name: "gateway"},
			},
			{
				Kind:   kinds.Gateway,
				Group:  v1.GroupName,
				Nsname: types.NamespacedName{Namespace: testNs, Name: "gateway2"}, // ignored
			},
		},
	}

	routePolicyKey := createTestPolicyKey(policyGVK, "route-policy")
	routePolicy := &Policy{
		Valid:  true,
		Source: &policiesfakes.FakePolicy{},
		TargetRefs: []PolicyTargetRef{
			{
				Kind:   kinds.HTTPRoute,
				Group:  v1.GroupName,
				Nsname: types.NamespacedName{Namespace: testNs, Name: "hr-route"},
			},
			{
				Kind:   kinds.HTTPRoute,
				Group:  v1.GroupName,
				Nsname: types.NamespacedName{Namespace: testNs, Name: "hr2-route"},
			},
		},
	}

	grpcRoutePolicyKey := createTestPolicyKey(policyGVK, "grpc-route-policy")
	grpcRoutePolicy := &Policy{
		Valid:  true,
		Source: &policiesfakes.FakePolicy{},
		TargetRefs: []PolicyTargetRef{
			{
				Kind:   kinds.GRPCRoute,
				Group:  v1.GroupName,
				Nsname: types.NamespacedName{Namespace: testNs, Name: "grpc-route"},
			},
		},
	}

	ngfPolicies := map[PolicyKey]*Policy{
		gwPolicyKey:        gwPolicy,
		routePolicyKey:     routePolicy,
		grpcRoutePolicyKey: grpcRoutePolicy,
	}

	createRouteKey := func(name string, routeType RouteType) RouteKey {
		return RouteKey{
			NamespacedName: types.NamespacedName{Name: name, Namespace: testNs},
			RouteType:      routeType,
		}
	}

	newGraph := func() *Graph {
		return &Graph{
			Gateway: &Gateway{
				Source: &v1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "gateway",
						Namespace: testNs,
					},
				},
				Valid: true,
			},
			Routes: map[RouteKey]*L7Route{
				createRouteKey("hr-route", RouteTypeHTTP): {
					Source: &v1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hr-route",
							Namespace: testNs,
						},
					},
					ParentRefs: []ParentRef{
						{
							Attachment: &ParentRefAttachmentStatus{
								Attached: true,
							},
						},
					},
					Valid:      true,
					Attachable: true,
				},
				createRouteKey("hr2-route", RouteTypeHTTP): {
					Source: &v1.HTTPRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "hr2-route",
							Namespace: testNs,
						},
					},
					ParentRefs: []ParentRef{
						{
							Attachment: &ParentRefAttachmentStatus{
								Attached: true,
							},
						},
					},
					Valid:      true,
					Attachable: true,
				},
				createRouteKey("grpc-route", RouteTypeGRPC): {
					Source: &v1alpha2.GRPCRoute{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "grpc-route",
							Namespace: testNs,
						},
					},
					ParentRefs: []ParentRef{
						{
							Attachment: &ParentRefAttachmentStatus{
								Attached: true,
							},
						},
					},
					Valid:      true,
					Attachable: true,
				},
			},

			NGFPolicies: ngfPolicies,
		}
	}

	newModifiedGraph := func(mod func(g *Graph) *Graph) *Graph {
		return mod(newGraph())
	}

	expectNoPolicyAttachment := func(g *WithT, graph *Graph) {
		if graph.Gateway != nil {
			g.Expect(graph.Gateway.Policies).To(BeNil())
		}

		for _, r := range graph.Routes {
			g.Expect(r.Policies).To(BeNil())
		}
	}

	expectPolicyAttachment := func(g *WithT, graph *Graph) {
		if graph.Gateway != nil {
			g.Expect(graph.Gateway.Policies).To(HaveLen(1))
		}

		for _, r := range graph.Routes {
			g.Expect(r.Policies).To(HaveLen(1))
		}
	}

	expectGatewayPolicyAttachment := func(g *WithT, graph *Graph) {
		if graph.Gateway != nil {
			g.Expect(graph.Gateway.Policies).To(HaveLen(1))
		}

		for _, r := range graph.Routes {
			g.Expect(r.Policies).To(BeNil())
		}
	}

	tests := []struct {
		graph  *Graph
		expect func(g *WithT, graph *Graph)
		name   string
	}{
		{
			name: "nil Gateway",
			graph: newModifiedGraph(func(g *Graph) *Graph {
				g.Gateway = nil
				return g
			}),
			expect: expectNoPolicyAttachment,
		},
		{
			name: "nil routes",
			graph: newModifiedGraph(func(g *Graph) *Graph {
				g.Routes = nil
				return g
			}),
			expect: expectGatewayPolicyAttachment,
		},
		{
			name:   "normal",
			graph:  newGraph(),
			expect: expectPolicyAttachment,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			test.graph.attachPolicies("nginx-gateway")
			test.expect(g, test.graph)
		})
	}
}

func TestAttachPolicyToRoute(t *testing.T) {
	routeNsName := types.NamespacedName{Namespace: testNs, Name: "hr-route"}

	createRoute := func(routeType RouteType, valid, attachable, parentRefs bool) *L7Route {
		route := &L7Route{
			Source: &v1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Name:      routeNsName.Name,
					Namespace: routeNsName.Namespace,
				},
			},
			Valid:      valid,
			Attachable: attachable,
			RouteType:  routeType,
		}

		if parentRefs {
			route.ParentRefs = []ParentRef{
				{
					Attachment: &ParentRefAttachmentStatus{
						Attached: true,
					},
				},
			}
		}

		return route
	}

	createGRPCRoute := func(valid, attachable, parentRefs bool) *L7Route {
		return createRoute(RouteTypeGRPC, valid, attachable, parentRefs)
	}

	createHTTPRoute := func(valid, attachable, parentRefs bool) *L7Route {
		return createRoute(RouteTypeHTTP, valid, attachable, parentRefs)
	}

	createExpAncestor := func(kind v1.Kind) v1.ParentReference {
		return v1.ParentReference{
			Group:     helpers.GetPointer[v1.Group](v1.GroupName),
			Kind:      helpers.GetPointer[v1.Kind](kind),
			Namespace: (*v1.Namespace)(&routeNsName.Namespace),
			Name:      v1.ObjectName(routeNsName.Name),
		}
	}

	tests := []struct {
		route        *L7Route
		policy       *Policy
		name         string
		expAncestors []PolicyAncestor
		expAttached  bool
	}{
		{
			name:   "policy attaches to http route",
			route:  createHTTPRoute(true /*valid*/, true /*attachable*/, true /*parentRefs*/),
			policy: &Policy{Source: &policiesfakes.FakePolicy{}},
			expAncestors: []PolicyAncestor{
				{Ancestor: createExpAncestor(kinds.HTTPRoute)},
			},
			expAttached: true,
		},
		{
			name:   "policy attaches to grpc route",
			route:  createGRPCRoute(true /*valid*/, true /*attachable*/, true /*parentRefs*/),
			policy: &Policy{Source: &policiesfakes.FakePolicy{}},
			expAncestors: []PolicyAncestor{
				{Ancestor: createExpAncestor(kinds.GRPCRoute)},
			},
			expAttached: true,
		},
		{
			name:  "attachment with existing ancestor",
			route: createHTTPRoute(true /*valid*/, true /*attachable*/, true /*parentRefs*/),
			policy: &Policy{
				Source: &policiesfakes.FakePolicy{},
				Ancestors: []PolicyAncestor{
					{Ancestor: createExpAncestor(kinds.HTTPRoute)},
				},
			},
			expAncestors: []PolicyAncestor{
				{Ancestor: createExpAncestor(kinds.HTTPRoute)},
				{Ancestor: createExpAncestor(kinds.HTTPRoute)},
			},
			expAttached: true,
		},
		{
			name:   "no attachment; unattachable route",
			route:  createHTTPRoute(true /*valid*/, false /*attachable*/, true /*parentRefs*/),
			policy: &Policy{Source: &policiesfakes.FakePolicy{}},
			expAncestors: []PolicyAncestor{
				{
					Ancestor:   createExpAncestor(kinds.HTTPRoute),
					Conditions: []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")},
				},
			},
			expAttached: false,
		},
		{
			name:   "no attachment; missing parentRefs",
			route:  createHTTPRoute(true /*valid*/, true /*attachable*/, false /*parentRefs*/),
			policy: &Policy{Source: &policiesfakes.FakePolicy{}},
			expAncestors: []PolicyAncestor{
				{
					Ancestor:   createExpAncestor(kinds.HTTPRoute),
					Conditions: []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")},
				},
			},
			expAttached: false,
		},
		{
			name:   "no attachment; invalid route",
			route:  createHTTPRoute(false /*valid*/, true /*attachable*/, true /*parentRefs*/),
			policy: &Policy{Source: &policiesfakes.FakePolicy{}},
			expAncestors: []PolicyAncestor{
				{
					Ancestor:   createExpAncestor(kinds.HTTPRoute),
					Conditions: []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")},
				},
			},
			expAttached: false,
		},
		{
			name:         "no attachment; max ancestors",
			route:        createHTTPRoute(true /*valid*/, true /*attachable*/, true /*parentRefs*/),
			policy:       &Policy{Source: createTestPolicyWithAncestors(16)},
			expAncestors: nil,
			expAttached:  false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			attachPolicyToRoute(test.policy, test.route, "nginx-gateway")

			if test.expAttached {
				g.Expect(test.route.Policies).To(HaveLen(1))
			} else {
				g.Expect(test.route.Policies).To(BeEmpty())
			}

			g.Expect(test.policy.Ancestors).To(BeEquivalentTo(test.expAncestors))
		})
	}
}

func TestAttachPolicyToGateway(t *testing.T) {
	gatewayNsName := types.NamespacedName{Namespace: testNs, Name: "gateway"}
	gateway2NsName := types.NamespacedName{Namespace: testNs, Name: "gateway2"}
	ignoredGatewayNsName := types.NamespacedName{Namespace: testNs, Name: "ignored"}

	newGateway := func(valid bool, nsname types.NamespacedName) *Gateway {
		return &Gateway{
			Source: &v1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: nsname.Namespace,
					Name:      nsname.Name,
				},
			},
			Valid: valid,
		}
	}

	getGatewayParentRef := func(gwNsName types.NamespacedName) v1.ParentReference {
		return v1.ParentReference{
			Group:     helpers.GetPointer[v1.Group](v1.GroupName),
			Kind:      helpers.GetPointer[v1.Kind]("Gateway"),
			Namespace: (*v1.Namespace)(&gwNsName.Namespace),
			Name:      v1.ObjectName(gwNsName.Name),
		}
	}

	tests := []struct {
		policy       *Policy
		gw           *Gateway
		name         string
		expAncestors []PolicyAncestor
		expAttached  bool
	}{
		{
			name: "attached",
			policy: &Policy{
				Source: &policiesfakes.FakePolicy{},
				TargetRefs: []PolicyTargetRef{
					{
						Nsname: gatewayNsName,
						Kind:   "Gateway",
					},
				},
			},
			gw: newGateway(true, gatewayNsName),
			expAncestors: []PolicyAncestor{
				{Ancestor: getGatewayParentRef(gatewayNsName)},
			},
			expAttached: true,
		},
		{
			name: "attached with existing ancestor",
			policy: &Policy{
				Source: &policiesfakes.FakePolicy{},
				TargetRefs: []PolicyTargetRef{
					{
						Nsname: gatewayNsName,
						Kind:   "Gateway",
					},
				},
				Ancestors: []PolicyAncestor{
					{Ancestor: getGatewayParentRef(gatewayNsName)},
				},
			},
			gw: newGateway(true, gatewayNsName),
			expAncestors: []PolicyAncestor{
				{Ancestor: getGatewayParentRef(gatewayNsName)},
				{Ancestor: getGatewayParentRef(gatewayNsName)},
			},
			expAttached: true,
		},
		{
			name: "not attached; gateway ignored",
			policy: &Policy{
				Source: &policiesfakes.FakePolicy{},
				TargetRefs: []PolicyTargetRef{
					{
						Nsname: ignoredGatewayNsName,
						Kind:   "Gateway",
					},
				},
			},
			gw: newGateway(true, gatewayNsName),
			expAncestors: []PolicyAncestor{
				{
					Ancestor:   getGatewayParentRef(ignoredGatewayNsName),
					Conditions: []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is ignored")},
				},
			},
			expAttached: false,
		},
		{
			name: "not attached; invalid gateway",
			policy: &Policy{
				Source: &policiesfakes.FakePolicy{},
				TargetRefs: []PolicyTargetRef{
					{
						Nsname: gatewayNsName,
						Kind:   "Gateway",
					},
				},
			},
			gw: newGateway(false, gatewayNsName),
			expAncestors: []PolicyAncestor{
				{
					Ancestor:   getGatewayParentRef(gatewayNsName),
					Conditions: []conditions.Condition{staticConds.NewPolicyTargetNotFound("TargetRef is invalid")},
				},
			},
			expAttached: false,
		},
		{
			name: "not attached; non-NGF gateway",
			policy: &Policy{
				Source: &policiesfakes.FakePolicy{},
				TargetRefs: []PolicyTargetRef{
					{
						Nsname: gateway2NsName,
						Kind:   "Gateway",
					},
				},
			},
			gw:           newGateway(true, gatewayNsName),
			expAncestors: nil,
			expAttached:  false,
		},
		{
			name: "not attached; max ancestors",
			policy: &Policy{
				Source: createTestPolicyWithAncestors(16),
				TargetRefs: []PolicyTargetRef{
					{
						Nsname: gatewayNsName,
						Kind:   "Gateway",
					},
				},
			},
			gw:           newGateway(true, gatewayNsName),
			expAncestors: nil,
			expAttached:  false,
		},
	}

	for _, test := range tests {
		ignoredGateways := map[types.NamespacedName]*v1.Gateway{
			ignoredGatewayNsName: nil,
		}

		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			attachPolicyToGateway(test.policy, test.policy.TargetRefs[0], test.gw, ignoredGateways, "nginx-gateway")

			if test.expAttached {
				g.Expect(test.gw.Policies).To(HaveLen(1))
			} else {
				g.Expect(test.gw.Policies).To(BeEmpty())
			}

			g.Expect(test.policy.Ancestors).To(BeEquivalentTo(test.expAncestors))
		})
	}
}

func TestProcessPolicies(t *testing.T) {
	policyGVK := schema.GroupVersionKind{Group: "Group", Version: "Version", Kind: "MyPolicy"}

	// These refs reference objects that belong to NGF.
	// Policies that contain these refs should be processed.
	hrRef := createTestRef(kinds.HTTPRoute, v1.GroupName, "hr")
	grpcRef := createTestRef(kinds.GRPCRoute, v1.GroupName, "grpc")
	gatewayRef := createTestRef(kinds.Gateway, v1.GroupName, "gw")
	ignoredGatewayRef := createTestRef(kinds.Gateway, v1.GroupName, "ignored")

	// These refs reference objects that do not belong to NGF.
	// Policies that contain these refs should NOT be processed.
	hrDoesNotExistRef := createTestRef(kinds.HTTPRoute, v1.GroupName, "dne")
	hrWrongGroup := createTestRef(kinds.HTTPRoute, "WrongGroup", "hr")
	gatewayWrongGroupRef := createTestRef(kinds.Gateway, "WrongGroup", "gw")
	nonNGFGatewayRef := createTestRef(kinds.Gateway, v1.GroupName, "not-ours")

	pol1, pol1Key := createTestPolicyAndKey(policyGVK, hrRef, "pol1")
	pol2, pol2Key := createTestPolicyAndKey(policyGVK, grpcRef, "pol2")
	pol3, pol3Key := createTestPolicyAndKey(policyGVK, gatewayRef, "pol3")
	pol4, pol4Key := createTestPolicyAndKey(policyGVK, ignoredGatewayRef, "pol4")
	pol5, pol5Key := createTestPolicyAndKey(policyGVK, hrDoesNotExistRef, "pol5")
	pol6, pol6Key := createTestPolicyAndKey(policyGVK, hrWrongGroup, "pol6")
	pol7, pol7Key := createTestPolicyAndKey(policyGVK, gatewayWrongGroupRef, "pol7")
	pol8, pol8Key := createTestPolicyAndKey(policyGVK, nonNGFGatewayRef, "pol8")

	pol1Conflict, pol1ConflictKey := createTestPolicyAndKey(policyGVK, hrRef, "pol1-conflict")

	allValidValidator := &policiesfakes.FakeValidator{}

	tests := []struct {
		validator            validation.PolicyValidator
		policies             map[PolicyKey]policies.Policy
		expProcessedPolicies map[PolicyKey]*Policy
		name                 string
	}{
		{
			name:                 "nil policies",
			expProcessedPolicies: nil,
		},
		{
			name:      "mix of relevant and irrelevant policies",
			validator: allValidValidator,
			policies: map[PolicyKey]policies.Policy{
				pol1Key: pol1,
				pol2Key: pol2,
				pol3Key: pol3,
				pol4Key: pol4,
				pol5Key: pol5,
				pol6Key: pol6,
				pol7Key: pol7,
				pol8Key: pol8,
			},
			expProcessedPolicies: map[PolicyKey]*Policy{
				pol1Key: {
					Source: pol1,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "hr"},
							Kind:   kinds.HTTPRoute,
							Group:  v1.GroupName,
						},
					},
					Ancestors: []PolicyAncestor{},
					Valid:     true,
				},
				pol2Key: {
					Source: pol2,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "grpc"},
							Kind:   kinds.GRPCRoute,
							Group:  v1.GroupName,
						},
					},
					Ancestors: []PolicyAncestor{},
					Valid:     true,
				},
				pol3Key: {
					Source: pol3,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "gw"},
							Kind:   kinds.Gateway,
							Group:  v1.GroupName,
						},
					},
					Ancestors: []PolicyAncestor{},
					Valid:     true,
				},
				pol4Key: {
					Source: pol4,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "ignored"},
							Kind:   kinds.Gateway,
							Group:  v1.GroupName,
						},
					},
					Ancestors: []PolicyAncestor{},
					Valid:     true,
				},
			},
		},
		{
			name: "invalid and valid policies",
			validator: &policiesfakes.FakeValidator{
				ValidateStub: func(
					policy policies.Policy,
					_ *policies.GlobalSettings,
				) []conditions.Condition {
					if policy.GetName() == "pol1" {
						return []conditions.Condition{staticConds.NewPolicyInvalid("invalid error")}
					}

					return nil
				},
			},
			policies: map[PolicyKey]policies.Policy{
				pol1Key: pol1,
				pol2Key: pol2,
			},
			expProcessedPolicies: map[PolicyKey]*Policy{
				pol1Key: {
					Source: pol1,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "hr"},
							Kind:   kinds.HTTPRoute,
							Group:  v1.GroupName,
						},
					},
					Conditions: []conditions.Condition{
						staticConds.NewPolicyInvalid("invalid error"),
					},
					Ancestors: []PolicyAncestor{},
					Valid:     false,
				},
				pol2Key: {
					Source: pol2,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "grpc"},
							Kind:   kinds.GRPCRoute,
							Group:  v1.GroupName,
						},
					},
					Ancestors: []PolicyAncestor{},
					Valid:     true,
				},
			},
		},
		{
			name: "conflicted policies",
			validator: &policiesfakes.FakeValidator{
				ConflictsStub: func(_ policies.Policy, _ policies.Policy) bool {
					return true
				},
			},
			policies: map[PolicyKey]policies.Policy{
				pol1Key:         pol1,
				pol1ConflictKey: pol1Conflict,
			},
			expProcessedPolicies: map[PolicyKey]*Policy{
				pol1Key: {
					Source: pol1,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "hr"},
							Kind:   kinds.HTTPRoute,
							Group:  v1.GroupName,
						},
					},
					Ancestors: []PolicyAncestor{},
					Valid:     true,
				},
				pol1ConflictKey: {
					Source: pol1Conflict,
					TargetRefs: []PolicyTargetRef{
						{
							Nsname: types.NamespacedName{Namespace: testNs, Name: "hr"},
							Kind:   kinds.HTTPRoute,
							Group:  v1.GroupName,
						},
					},
					Conditions: []conditions.Condition{
						staticConds.NewPolicyConflicted("Conflicts with another MyPolicy"),
					},
					Ancestors: []PolicyAncestor{},
					Valid:     false,
				},
			},
		},
	}

	gateways := processedGateways{
		Winner: &v1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "gw",
				Namespace: testNs,
			},
		},
		Ignored: map[types.NamespacedName]*v1.Gateway{
			{Namespace: testNs, Name: "ignored"}: {
				ObjectMeta: metav1.ObjectMeta{
					Name:      "gw",
					Namespace: testNs,
				},
			},
		},
	}

	routes := map[RouteKey]*L7Route{
		{RouteType: RouteTypeHTTP, NamespacedName: types.NamespacedName{Namespace: testNs, Name: "hr"}}:   {},
		{RouteType: RouteTypeGRPC, NamespacedName: types.NamespacedName{Namespace: testNs, Name: "grpc"}}: {},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			processed := processPolicies(test.policies, test.validator, gateways, routes, nil)
			g.Expect(processed).To(BeEquivalentTo(test.expProcessedPolicies))
		})
	}
}

func TestMarkConflictedPolicies(t *testing.T) {
	hrRef := createTestRef(kinds.HTTPRoute, v1.GroupName, "hr")
	hrTargetRef := PolicyTargetRef{
		Kind:   hrRef.Kind,
		Group:  hrRef.Group,
		Nsname: types.NamespacedName{Namespace: testNs, Name: string(hrRef.Name)},
	}

	grpcRef := createTestRef(kinds.GRPCRoute, v1.GroupName, "grpc")
	grpcTargetRef := PolicyTargetRef{
		Kind:   grpcRef.Kind,
		Group:  grpcRef.Group,
		Nsname: types.NamespacedName{Namespace: testNs, Name: string(grpcRef.Name)},
	}

	orangeGVK := schema.GroupVersionKind{Group: "Fruits", Version: "Fresh", Kind: "OrangePolicy"}
	appleGVK := schema.GroupVersionKind{Group: "Fruits", Version: "Fresh", Kind: "ApplePolicy"}

	tests := []struct {
		name                  string
		policies              map[PolicyKey]*Policy
		fakeValidator         *policiesfakes.FakeValidator
		conflictedNames       []string
		expConflictToBeCalled bool
	}{
		{
			name: "different policy types can not conflict",
			policies: map[PolicyKey]*Policy{
				createTestPolicyKey(orangeGVK, "orange"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(appleGVK, "apple"): {
					Source:     createTestPolicy(appleGVK, hrRef, "apple"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
			},
			fakeValidator:         &policiesfakes.FakeValidator{},
			expConflictToBeCalled: false,
		},
		{
			name: "policies of the same type but with different target refs can not conflict",
			policies: map[PolicyKey]*Policy{
				createTestPolicyKey(orangeGVK, "orange1"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange1"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(orangeGVK, "orange2"): {
					Source:     createTestPolicy(orangeGVK, grpcRef, "orange2"),
					TargetRefs: []PolicyTargetRef{grpcTargetRef},
					Valid:      true,
				},
			},
			fakeValidator:         &policiesfakes.FakeValidator{},
			expConflictToBeCalled: false,
		},
		{
			name: "invalid policies can not conflict",
			policies: map[PolicyKey]*Policy{
				createTestPolicyKey(orangeGVK, "valid"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "valid"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(orangeGVK, "invalid"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "invalid"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      false,
				},
			},
			fakeValidator:         &policiesfakes.FakeValidator{},
			expConflictToBeCalled: false,
		},
		{
			name: "when a policy conflicts with a policy that has greater precedence it's marked as invalid and a" +
				" condition is added",
			policies: map[PolicyKey]*Policy{
				createTestPolicyKey(orangeGVK, "orange1"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange1"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(orangeGVK, "orange2"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange2"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(orangeGVK, "orange3-conflicts-with-1"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange3-conflicts-with-1"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(orangeGVK, "orange4"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange4"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
				createTestPolicyKey(orangeGVK, "orange5-conflicts-with-4"): {
					Source:     createTestPolicy(orangeGVK, hrRef, "orange5-conflicts-with-4"),
					TargetRefs: []PolicyTargetRef{hrTargetRef},
					Valid:      true,
				},
			},
			fakeValidator: &policiesfakes.FakeValidator{
				ConflictsStub: func(policy policies.Policy, policy2 policies.Policy) bool {
					pol1Name := policy.GetName()
					pol2Name := policy2.GetName()

					if pol1Name == "orange1" && pol2Name == "orange3-conflicts-with-1" {
						return true
					}

					if pol1Name == "orange4" && pol2Name == "orange5-conflicts-with-4" {
						return true
					}

					return false
				},
			},
			conflictedNames:       []string{"orange3-conflicts-with-1", "orange5-conflicts-with-4"},
			expConflictToBeCalled: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			markConflictedPolicies(test.policies, test.fakeValidator)

			if !test.expConflictToBeCalled {
				g.Expect(test.fakeValidator.ConflictsCallCount()).To(BeZero())
			} else {
				g.Expect(test.fakeValidator.ConflictsCallCount()).To(Not(BeZero()))
				expConflictCond := staticConds.NewPolicyConflicted("Conflicts with another OrangePolicy")

				for key, policy := range test.policies {
					if slices.Contains(test.conflictedNames, key.NsName.Name) {
						g.Expect(policy.Valid).To(BeFalse())
						g.Expect(policy.Conditions).To(ConsistOf(expConflictCond))
					} else {
						g.Expect(policy.Valid).To(BeTrue())
						g.Expect(policy.Conditions).To(BeEmpty())
					}
				}
			}
		})
	}
}

func createTestPolicyWithAncestors(numAncestors int) policies.Policy {
	policy := &policiesfakes.FakePolicy{}

	ancestors := make([]v1alpha2.PolicyAncestorStatus, numAncestors)

	for i := range numAncestors {
		ancestors[i] = v1alpha2.PolicyAncestorStatus{ControllerName: "some-other-controller"}
	}

	policy.GetPolicyStatusReturns(v1alpha2.PolicyStatus{Ancestors: ancestors})
	return policy
}

func createTestPolicyAndKey(
	gvk schema.GroupVersionKind,
	ref v1alpha2.LocalPolicyTargetReference,
	name string,
) (policies.Policy, PolicyKey) {
	pol := createTestPolicy(gvk, ref, name)
	key := createTestPolicyKey(gvk, name)

	return pol, key
}

func createTestPolicy(
	gvk schema.GroupVersionKind,
	ref v1alpha2.LocalPolicyTargetReference,
	name string,
) policies.Policy {
	return &policiesfakes.FakePolicy{
		GetNameStub: func() string {
			return name
		},
		GetNamespaceStub: func() string {
			return testNs
		},
		GetTargetRefsStub: func() []v1alpha2.LocalPolicyTargetReference {
			return []v1alpha2.LocalPolicyTargetReference{ref}
		},
		GetObjectKindStub: func() schema.ObjectKind {
			return &policiesfakes.FakeObjectKind{
				GroupVersionKindStub: func() schema.GroupVersionKind {
					return gvk
				},
			}
		},
	}
}

func createTestPolicyKey(gvk schema.GroupVersionKind, name string) PolicyKey {
	return PolicyKey{
		NsName: types.NamespacedName{Namespace: testNs, Name: name},
		GVK:    gvk,
	}
}

func createTestRef(kind v1.Kind, group v1.Group, name string) v1alpha2.LocalPolicyTargetReference {
	return v1alpha2.LocalPolicyTargetReference{
		Group: group,
		Kind:  kind,
		Name:  v1.ObjectName(name),
	}
}
