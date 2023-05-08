package filter

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

// CreateFilterForGatewayClass creates a filter function that filters out all GatewayClass resources except the one
// with the given name.
func CreateFilterForGatewayClass(gcName string) reconciler.NamespacedNameFilterFunc {
	return func(nsname types.NamespacedName) (bool, string) {
		if nsname.Name != gcName {
			return false, fmt.Sprintf(
				"GatewayClass is ignored because this controller only supports the GatewayClass %s",
				gcName,
			)
		}
		return true, ""
	}
}

func CreateFilterForGateway(gwNsName *types.NamespacedName) reconciler.NamespacedNameFilterFunc {
	if gwNsName == nil {
		return nil
	}
	return func(nsname types.NamespacedName) (bool, string) {
		if nsname != *gwNsName {
			return false, fmt.Sprintf(
				"Gateway is ignored because this controller only supports the Gateway %s/%s",
				gwNsName.Namespace,
				gwNsName.Name,
			)
		}
		return true, ""
	}
}
