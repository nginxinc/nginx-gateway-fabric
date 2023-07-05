package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestReferenceGrantResolver(t *testing.T) {
	gwNs := "gw-ns"
	secretNsName := types.NamespacedName{Namespace: "test", Name: "certificate"}

	getNormalRefGrant := func() *v1beta1.ReferenceGrant {
		return &v1beta1.ReferenceGrant{
			Spec: v1beta1.ReferenceGrantSpec{
				From: []v1beta1.ReferenceGrantFrom{
					{
						Group:     v1beta1.GroupName,
						Kind:      "Gateway",
						Namespace: v1beta1.Namespace(gwNs),
					},
				},
				To: []v1beta1.ReferenceGrantTo{
					{
						Kind: "Secret",
						Name: helpers.GetPointer(v1beta1.ObjectName(secretNsName.Name)),
					},
				},
			},
		}
	}

	createModifiedRefGrant := func(mod func(rg *v1beta1.ReferenceGrant)) *v1beta1.ReferenceGrant {
		rg := getNormalRefGrant()
		mod(rg)
		return rg
	}

	refGrants := map[types.NamespacedName]*v1beta1.ReferenceGrant{
		{Namespace: "test", Name: "valid"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
			rg.Spec.To = []v1beta1.ReferenceGrantTo{
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
			}
		}),
		{Namespace: "explicit-core-group", Name: "valid"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
			rg.Spec.To[0].Group = "core"
		}),
		{Namespace: "all-in-namespace", Name: "valid"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
			rg.Spec.To[0].Name = nil
			rg.Spec.From = []v1beta1.ReferenceGrantFrom{
				{
					Group:     v1beta1.GroupName,
					Kind:      "Gateway",
					Namespace: "wrong-ns1",
				},
				{
					Group:     v1beta1.GroupName,
					Kind:      "Gateway",
					Namespace: "wrong-ns2",
				},
				{
					Group:     v1beta1.GroupName,
					Kind:      "Gateway",
					Namespace: v1beta1.Namespace(gwNs), // matches
				},
			}
		}),
	}

	tests := []struct {
		overrideTo   *toResource
		overrideFrom *fromResource
		msg          string
		allowed      bool
	}{
		{
			msg:        "wrong 'to' kind",
			overrideTo: &toResource{kind: "WrongKind", name: secretNsName.Name, namespace: secretNsName.Namespace},
			allowed:    false,
		},
		{
			msg: "wrong 'to' group",
			overrideTo: &toResource{
				group:     "wrong.group",
				kind:      "Secret",
				name:      secretNsName.Name,
				namespace: secretNsName.Namespace,
			},
			allowed: false,
		},
		{
			msg:        "wrong 'to' name",
			overrideTo: &toResource{kind: "Secret", name: "wrong-name", namespace: secretNsName.Namespace},
			allowed:    false,
		},
		{
			msg:          "wrong 'from' kind",
			overrideFrom: &fromResource{group: v1beta1.GroupName, kind: "WrongKind", namespace: gwNs},
			allowed:      false,
		},
		{
			msg:          "wrong 'from' group",
			overrideFrom: &fromResource{group: "wrong.group", kind: "Gateway", namespace: gwNs},
			allowed:      false,
		},
		{
			msg:          "wrong 'from' namespace",
			overrideFrom: &fromResource{group: v1beta1.GroupName, kind: "Gateway", namespace: "wrong-ns"},
			allowed:      false,
		},
		{
			msg:     "allowed; matches specific reference grant",
			allowed: true,
		},
		{
			msg:        "allowed; matches all-in-namespace reference grant",
			overrideTo: &toResource{kind: "Secret", name: secretNsName.Name, namespace: "all-in-namespace"},
			allowed:    true,
		},
	}

	resolver := newReferenceGrantResolver(refGrants)

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewGomegaWithT(t)

			to := toResource{kind: "Secret", name: secretNsName.Name, namespace: secretNsName.Namespace}
			if test.overrideTo != nil {
				to = *test.overrideTo
			}

			from := fromResource{group: v1beta1.GroupName, kind: "Gateway", namespace: gwNs}
			if test.overrideFrom != nil {
				from = *test.overrideFrom
			}

			g.Expect(resolver.refAllowed(to, from)).To(Equal(test.allowed))
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

	g := NewGomegaWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestToService(t *testing.T) {
	ref := toService(types.NamespacedName{Namespace: "ns", Name: "service"})

	exp := toResource{
		kind:      "Service",
		namespace: "ns",
		name:      "service",
	}

	g := NewGomegaWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestFromGateway(t *testing.T) {
	ref := fromGateway("ns")

	exp := fromResource{
		group:     v1beta1.GroupName,
		kind:      "Gateway",
		namespace: "ns",
	}

	g := NewGomegaWithT(t)
	g.Expect(ref).To(Equal(exp))
}

func TestFromHTTPRoute(t *testing.T) {
	ref := fromHTTPRoute("ns")

	exp := fromResource{
		group:     v1beta1.GroupName,
		kind:      "HTTPRoute",
		namespace: "ns",
	}

	g := NewGomegaWithT(t)
	g.Expect(ref).To(Equal(exp))
}
