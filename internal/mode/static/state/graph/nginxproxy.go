package graph

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

func getNginxProxyConfig(
	nps map[types.NamespacedName]*ngfAPI.NginxProxy,
	gc *v1beta1.GatewayClass,
) *ngfAPI.NginxProxy {
	if gc != nil {
		ref := gc.Spec.ParametersRef
		if ref != nil && ref.Namespace != nil &&
			ref.Group == ngfAPI.GroupName && ref.Kind == v1beta1.Kind("NginxProxy") {
			nsName := types.NamespacedName{Name: ref.Name, Namespace: string(*ref.Namespace)}
			return nps[nsName]
		}
	}

	return nil
}
