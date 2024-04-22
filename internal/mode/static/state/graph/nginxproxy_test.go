package graph

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

func TestGetNginxProxy(t *testing.T) {
	tests := []struct {
		nps   map[types.NamespacedName]*ngfAPI.NginxProxy
		gc    *v1.GatewayClass
		expNP *ngfAPI.NginxProxy
		name  string
	}{
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {},
			},
			gc:    nil,
			expNP: nil,
			name:  "nil gatewayclass",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {},
			},
			gc:    &v1.GatewayClass{},
			expNP: nil,
			name:  "nil paramsRef",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{},
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind("NginxProxy"),
						Name:  "np1",
					},
				},
			},
			expNP: nil,
			name:  "no nginxproxy resources",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: "np1",
					},
				},
			},
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: v1.Group("wrong-group"),
						Kind:  v1.Kind("NginxProxy"),
						Name:  "wrong-group",
					},
				},
			},
			expNP: nil,
			name:  "wrong group",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: "np1",
					},
				},
			},
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind("WrongKind"),
						Name:  "wrong-kind",
					},
				},
			},
			expNP: nil,
			name:  "wrong kind",
		},
		{
			nps: map[types.NamespacedName]*ngfAPI.NginxProxy{
				{Name: "np1"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: "np1",
					},
				},
				{Name: "np2"}: {
					ObjectMeta: metav1.ObjectMeta{
						Name: "np2",
					},
				},
			},
			gc: &v1.GatewayClass{
				Spec: v1.GatewayClassSpec{
					ParametersRef: &v1.ParametersReference{
						Group: ngfAPI.GroupName,
						Kind:  v1.Kind("NginxProxy"),
						Name:  "np2",
					},
				},
			},
			expNP: &ngfAPI.NginxProxy{
				ObjectMeta: metav1.ObjectMeta{
					Name: "np2",
				},
			},
			name: "returns correct resource",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewWithT(t)

			g.Expect(getNginxProxy(test.nps, test.gc)).To(Equal(test.expNP))
		})
	}
}
