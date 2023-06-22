package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func TestRefGrantAllowsGatewayToSecret(t *testing.T) {
	gwNs := "gw-ns"
	secretNsName := types.NamespacedName{Namespace: "test", Name: "certificate"}

	getNormalRefGrant := func() *v1beta1.ReferenceGrant {
		return &v1beta1.ReferenceGrant{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "rg",
				Namespace: "test",
			},
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
						Group: "core",
						Kind:  "Secret",
						Name:  helpers.GetPointer(v1beta1.ObjectName(secretNsName.Name)),
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

	tests := []struct {
		refGrants map[types.NamespacedName]*v1beta1.ReferenceGrant
		msg       string
		allowed   bool
	}{
		{
			msg: "allowed; specific ref grant exists",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "wrong-ns", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Namespace = "wrong-ns"
				}),
				{Namespace: "test", Name: "wrong-to-kind"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To[0].Kind = "WrongKind"
					rg.Name = "wrong-to-kind"
				}),
				{Namespace: "test", Name: "rg"}: getNormalRefGrant(),
			},
			allowed: true,
		},
		{
			msg: "allowed; all-namespace ref grant exists",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To[0].Name = nil
				}),
			},
			allowed: true,
		},
		{
			msg: "allowed; implicit to Group",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To[0].Group = ""
				}),
			},
			allowed: true,
		},
		{
			msg: "allowed; one matching 'to' ref",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To = []v1beta1.ReferenceGrantTo{
						{
							Group: "wrong.group",
						},
						{
							Kind: "WrongKind",
						},
						{
							Group: "core",
							Kind:  "Secret",
							Name:  helpers.GetPointer(v1beta1.ObjectName(secretNsName.Name)),
						},
					}
				}),
			},
			allowed: true,
		},
		{
			msg: "allowed; one matching 'from' ref",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.From = []v1beta1.ReferenceGrantFrom{
						{
							Group: "wrong.group",
						},
						{
							Kind: "WrongKind",
						},
						{
							Group:     "gateway.networking.k8s.io",
							Kind:      "Gateway",
							Namespace: v1beta1.Namespace(gwNs),
						},
					}
				}),
			},
			allowed: true,
		},
		{
			msg: "not allowed; no ref group in secret namespace",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "wrong-ns", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Namespace = "wrong-ns"
				}),
			},
			allowed: false,
		},
		{
			msg: "not allowed; no ref group with the right from Group",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.From[0].Group = "wrong.group"
				}),
			},
			allowed: false,
		},
		{
			msg: "not allowed; no ref group with the right from Kind",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.From[0].Kind = "WrongKind"
				}),
			},
			allowed: false,
		},
		{
			msg: "not allowed; no ref group with the right from Namespace",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.From[0].Namespace = "wrong-ns"
				}),
			},
			allowed: false,
		},
		{
			msg: "not allowed; no ref group with the right to Group",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To[0].Group = "wrong.group"
				}),
			},
			allowed: false,
		},
		{
			msg: "not allowed; no ref group with the right to Kind",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To[0].Kind = "WrongKind"
				}),
			},
			allowed: false,
		},
		{
			msg: "not allowed; no ref group with the right to Name",
			refGrants: map[types.NamespacedName]*v1beta1.ReferenceGrant{
				{Namespace: "test", Name: "rg"}: createModifiedRefGrant(func(rg *v1beta1.ReferenceGrant) {
					rg.Spec.To[0].Name = helpers.GetPointer(v1beta1.ObjectName("wrong-name"))
				}),
			},
			allowed: false,
		},
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewGomegaWithT(t)

			allowed := refGrantAllowsGatewayToSecret(test.refGrants, gwNs, secretNsName)
			g.Expect(allowed).To(Equal(test.allowed))
		})
	}
}
