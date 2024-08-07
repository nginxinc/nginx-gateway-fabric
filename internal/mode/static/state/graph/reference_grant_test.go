package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/kinds"
)

func TestReferenceGrantResolver(t *testing.T) {
	gwNs := "gw-ns"
	secretNsName := types.NamespacedName{Namespace: "test", Name: "certificate"}

	refGrants := map[types.NamespacedName]*v1beta1.ReferenceGrant{
		{Namespace: "test", Name: "valid"}: {
			Spec: v1beta1.ReferenceGrantSpec{
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.Gateway,
						Namespace: v1beta1.Namespace(gwNs),
					},
				},
				To: []v1beta1.ReferenceGrantTo{
					{
						Kind: "Secret",
						Name: helpers.GetPointer(v1beta1.ObjectName("wrong-name1")),
					},
					{
						Kind: "Secret",
						Name: helpers.GetPointer(v1beta1.ObjectName("wrong-name2")),
					},
					{
						Kind: "Secret",
						Name: helpers.GetPointer(v1beta1.ObjectName(secretNsName.Name)), // matches
					},
				},
			},
		},
		{Namespace: "explicit-core-group", Name: "valid"}: {
			Spec: v1beta1.ReferenceGrantSpec{
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.Gateway,
						Namespace: v1beta1.Namespace(gwNs),
					},
				},
				To: []v1beta1.ReferenceGrantTo{
					{
						Group: "core",
						Kind:  "Secret",
						Name:  helpers.GetPointer(v1beta1.ObjectName(secretNsName.Name)),
					},
				},
			},
		},
		{Namespace: "all-in-namespace", Name: "valid"}: {
			Spec: v1beta1.ReferenceGrantSpec{
				To: []v1beta1.ReferenceGrantTo{
					{
						Kind: "Secret",
					},
				},
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.Gateway,
						Namespace: "wrong-ns1",
					},
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.Gateway,
						Namespace: "wrong-ns2",
					},
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.Gateway,
						Namespace: v1beta1.Namespace(gwNs),
					},
				},
			},
		},
	}

	normalTo := toResource{kind: "Secret", name: secretNsName.Name, namespace: secretNsName.Namespace}
	normalFrom := fromResource{group: v1beta1.GroupName, kind: kinds.Gateway, namespace: gwNs}

	tests := []struct {
		to      toResource
		from    fromResource
		msg     string
		allowed bool
	}{
		{
			msg:     "wrong 'to' kind",
			to:      toResource{kind: "WrongKind", name: secretNsName.Name, namespace: secretNsName.Namespace},
			from:    normalFrom,
			allowed: false,
		},
		{
			msg: "wrong 'to' group",
			to: toResource{
				group:     "wrong.group",
				kind:      "Secret",
				name:      secretNsName.Name,
				namespace: secretNsName.Namespace,
			},
			from:    normalFrom,
			allowed: false,
		},
		{
			msg:     "wrong 'to' name",
			to:      toResource{kind: "Secret", name: "wrong-name", namespace: secretNsName.Namespace},
			from:    normalFrom,
			allowed: false,
		},
		{
			msg:     "wrong 'from' kind",
			to:      normalTo,
			from:    fromResource{group: v1beta1.GroupName, kind: "WrongKind", namespace: gwNs},
			allowed: false,
		},
		{
			msg:     "wrong 'from' group",
			to:      normalTo,
			from:    fromResource{group: "wrong.group", kind: kinds.Gateway, namespace: gwNs},
			allowed: false,
		},
		{
			msg:     "wrong 'from' namespace",
			to:      normalTo,
			from:    fromResource{group: v1beta1.GroupName, kind: kinds.Gateway, namespace: "wrong-ns"},
			allowed: false,
		},
		{
			msg:     "allowed; matches specific reference grant",
			to:      normalTo,
			from:    normalFrom,
			allowed: true,
		},
		{
			msg:     "allowed; matches all-in-namespace reference grant",
			to:      toResource{kind: "Secret", name: secretNsName.Name, namespace: "all-in-namespace"},
			from:    normalFrom,
			allowed: true,
		},
		{
			msg:     "allowed; matches specific reference grant with explicit 'core' group name",
			to:      toResource{kind: "Secret", name: secretNsName.Name, namespace: "explicit-core-group"},
			from:    normalFrom,
			allowed: true,
		},
	}

	resolver := newReferenceGrantResolver(refGrants)

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(resolver.refAllowed(test.to, test.from)).To(Equal(test.allowed))
		})
	}
}

func TestToSecret(t *testing.T) {
	ref := toSecret(types.NamespacedName{Namespace: "ns", Name: "secret"})

	exp := toResource{
		kind:      "Secret",
		namespace: "ns",
		name:      "secret",
	}

	g := NewWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestToService(t *testing.T) {
	ref := toService(types.NamespacedName{Namespace: "ns", Name: "service"})

	exp := toResource{
		kind:      "Service",
		namespace: "ns",
		name:      "service",
	}

	g := NewWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestFromGateway(t *testing.T) {
	ref := fromGateway("ns")

	exp := fromResource{
		group:     v1beta1.GroupName,
		kind:      kinds.Gateway,
		namespace: "ns",
	}

	g := NewWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestFromHTTPRoute(t *testing.T) {
	ref := fromHTTPRoute("ns")

	exp := fromResource{
		group:     v1beta1.GroupName,
		kind:      kinds.HTTPRoute,
		namespace: "ns",
	}

	g := NewWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestFromGRPCRoute(t *testing.T) {
	ref := fromGRPCRoute("ns")

	exp := fromResource{
		group:     v1beta1.GroupName,
		kind:      kinds.GRPCRoute,
		namespace: "ns",
	}

	g := NewWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestRefAllowedFrom(t *testing.T) {
	gwNs := "gw-ns"
	hrNs := "hr-ns"
	grNs := "gr-ns"

	allowedHTTPRouteNs := "hr-allowed-ns"
	allowedHTTPRouteNsName := types.NamespacedName{Namespace: allowedHTTPRouteNs, Name: "all-allowed-in-ns"}

	allowedGRPCRouteNs := "gr-allowed-ns"
	allowedGRPCRouteNsName := types.NamespacedName{Namespace: allowedGRPCRouteNs, Name: "all-allowed-in-ns"}

	allowedGatewayNs := "gw-allowed-ns"
	allowedGatewayNsName := types.NamespacedName{Namespace: allowedGatewayNs, Name: "all-allowed-in-ns"}

	notAllowedNsName := types.NamespacedName{Namespace: "not-allowed-ns", Name: "not-allowed-in-ns"}

	refGrants := map[types.NamespacedName]*v1beta1.ReferenceGrant{
		{Namespace: allowedGatewayNs, Name: "gw-2-secret"}: {
			Spec: v1beta1.ReferenceGrantSpec{
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.Gateway,
						Namespace: v1beta1.Namespace(gwNs),
					},
				},
				To: []v1beta1.ReferenceGrantTo{
					{
						Kind: "Secret",
					},
				},
			},
		},
		{Namespace: allowedHTTPRouteNs, Name: "hr-2-svc"}: {
			Spec: v1beta1.ReferenceGrantSpec{
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.HTTPRoute,
						Namespace: v1beta1.Namespace(hrNs),
					},
				},
				To: []v1beta1.ReferenceGrantTo{
					{
						Kind: "Service",
					},
				},
			},
		},
		{Namespace: allowedGRPCRouteNs, Name: "gr-2-svc"}: {
			Spec: v1beta1.ReferenceGrantSpec{
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      kinds.GRPCRoute,
						Namespace: v1beta1.Namespace(grNs),
					},
				},
				To: []v1beta1.ReferenceGrantTo{
					{
						Kind: "Service",
					},
				},
			},
		},
	}

	resolver := newReferenceGrantResolver(refGrants)
	refAllowedFromGRPCRoute := resolver.refAllowedFrom(fromGRPCRoute(grNs))
	refAllowedFromHTTPRoute := resolver.refAllowedFrom(fromHTTPRoute(hrNs))
	refAllowedFromGateway := resolver.refAllowedFrom(fromGateway(gwNs))

	tests := []struct {
		name           string
		refAllowedFrom func(resource toResource) bool
		toResource     toResource
		expAllowed     bool
	}{
		{
			name:           "ref allowed from gateway to secret",
			refAllowedFrom: refAllowedFromGateway,
			toResource:     toSecret(allowedGatewayNsName),
			expAllowed:     true,
		},
		{
			name:           "ref not allowed from gateway to secret",
			refAllowedFrom: refAllowedFromGateway,
			toResource:     toSecret(notAllowedNsName),
			expAllowed:     false,
		},
		{
			name:           "ref allowed from httproute to service",
			refAllowedFrom: refAllowedFromHTTPRoute,
			toResource:     toService(allowedHTTPRouteNsName),
			expAllowed:     true,
		},
		{
			name:           "ref not allowed from httproute to service",
			refAllowedFrom: refAllowedFromHTTPRoute,
			toResource:     toService(notAllowedNsName),
			expAllowed:     false,
		},
		{
			name:           "ref allowed from grpcroute to service",
			refAllowedFrom: refAllowedFromGRPCRoute,
			toResource:     toService(allowedGRPCRouteNsName),
			expAllowed:     true,
		},
		{
			name:           "ref not allowed from grpcroute to service",
			refAllowedFrom: refAllowedFromGRPCRoute,
			toResource:     toService(notAllowedNsName),
			expAllowed:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)
			g.Expect(test.refAllowedFrom(test.toResource)).To(Equal(test.expAllowed))
		})
	}
}
