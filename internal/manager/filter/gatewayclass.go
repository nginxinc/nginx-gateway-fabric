package filter

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

func CreateFilterForGatewayClass(gcName string) reconciler.NamespacedNameFilterFunc {
	return func(nsname types.NamespacedName) (bool, string) {
		if nsname.Name != gcName {
			return false, fmt.Sprintf("GatewayClass is ignored because this controller only supports the GatewayClass %s", gcName)
		}
		return true, ""
	}
}
