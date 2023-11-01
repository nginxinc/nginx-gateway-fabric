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
	var cfg *ngfAPI.NginxProxy
	for _, np := range nps {
		ref := gc.Spec.ParametersRef
		if ref != nil {
			if ref.Group == ngfAPI.GroupName &&
				ref.Kind == v1beta1.Kind("NginxProxy") &&
				ref.Name == np.Name &&
				ref.Namespace != nil &&
				*ref.Namespace == v1beta1.Namespace(np.Namespace) {
				cfg = np
				break
			}
		}
	}

	return cfg
}
