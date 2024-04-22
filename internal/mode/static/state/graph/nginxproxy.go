package graph

import (
	"k8s.io/apimachinery/pkg/types"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
)

// getNginxProxy returns the NginxProxy associated with the GatewayClass (if it exists).
func getNginxProxy(
	nps map[types.NamespacedName]*ngfAPI.NginxProxy,
	gc *v1.GatewayClass,
) *ngfAPI.NginxProxy {
	if gc != nil {
		ref := gc.Spec.ParametersRef
		if ref != nil && ref.Group == ngfAPI.GroupName && ref.Kind == v1.Kind("NginxProxy") {
			return nps[types.NamespacedName{Name: ref.Name}]
		}
	}

	return nil
}
