package static

import (
	"testing"

	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestPrepareFirstEventBatchPreparerArgs(t *testing.T) {
	const gcName = "nginx"

	tests := []struct {
		name                string
		gwNsName            *types.NamespacedName
		expectedObjects     []client.Object
		expectedObjectLists []client.ObjectList
	}{
		{
			name:     "gwNsName is nil",
			gwNsName: nil,
			expectedObjects: []client.Object{
				&gatewayv1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1beta1.HTTPRouteList{},
				&gatewayv1beta1.GatewayList{},
				&gatewayv1beta1.ReferenceGrantList{},
			},
		},
		{
			name: "gwNsName is not nil",
			gwNsName: &types.NamespacedName{
				Namespace: "test",
				Name:      "my-gateway",
			},
			expectedObjects: []client.Object{
				&gatewayv1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: "nginx"}},
				&gatewayv1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: "my-gateway", Namespace: "test"}},
			},
			expectedObjectLists: []client.ObjectList{
				&apiv1.ServiceList{},
				&apiv1.SecretList{},
				&apiv1.NamespaceList{},
				&discoveryV1.EndpointSliceList{},
				&gatewayv1beta1.HTTPRouteList{},
				&gatewayv1beta1.ReferenceGrantList{},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			g := NewGomegaWithT(t)

			objects, objectLists := prepareFirstEventBatchPreparerArgs(gcName, test.gwNsName)

			g.Expect(objects).To(ConsistOf(test.expectedObjects))
			g.Expect(objectLists).To(ConsistOf(test.expectedObjectLists))
		})
	}
}
