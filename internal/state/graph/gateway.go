package graph

import (
	"sort"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	nkgsort "github.com/nginxinc/nginx-kubernetes-gateway/internal/sort"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

// Gateway represents the winning Gateway resource.
type Gateway struct {
	// Source is the corresponding Gateway resource.
	Source *v1beta1.Gateway
	// Listeners include the listeners of the Gateway.
	Listeners map[string]*Listener
}

// processedGateways holds the resources that belong to NKG.
type processedGateways struct {
	Winner  *v1beta1.Gateway
	Ignored map[types.NamespacedName]*v1beta1.Gateway
}

// GetAllNsNames returns all the NamespacedNames of the Gateway resources that belong to NKG
func (gws processedGateways) GetAllNsNames() []types.NamespacedName {
	winnerCnt := 0
	if gws.Winner != nil {
		winnerCnt = 1
	}

	length := winnerCnt + len(gws.Ignored)
	if length == 0 {
		return nil
	}

	allNsNames := make([]types.NamespacedName, 0, length)

	if gws.Winner != nil {
		allNsNames = append(allNsNames, client.ObjectKeyFromObject(gws.Winner))
	}
	for nsName := range gws.Ignored {
		allNsNames = append(allNsNames, nsName)
	}

	return allNsNames
}

// processGateways determines which Gateway resource belong to NKG (determined by the Gateway GatewayClassName field).
func processGateways(
	gws map[types.NamespacedName]*v1beta1.Gateway,
	gcName string,
) processedGateways {
	referencedGws := make([]*v1beta1.Gateway, 0, len(gws))

	for _, gw := range gws {
		if string(gw.Spec.GatewayClassName) != gcName {
			continue
		}

		referencedGws = append(referencedGws, gw)
	}

	if len(referencedGws) == 0 {
		return processedGateways{}
	}

	sort.Slice(referencedGws, func(i, j int) bool {
		return nkgsort.LessObjectMeta(&referencedGws[i].ObjectMeta, &referencedGws[j].ObjectMeta)
	})

	ignoredGws := make(map[types.NamespacedName]*v1beta1.Gateway)

	for _, gw := range referencedGws[1:] {
		ignoredGws[client.ObjectKeyFromObject(gw)] = gw
	}

	return processedGateways{
		Winner:  referencedGws[0],
		Ignored: ignoredGws,
	}
}

func buildGateway(gw *v1beta1.Gateway, secretMemoryMgr secrets.SecretDiskMemoryManager, gc *GatewayClass) *Gateway {
	if gw == nil {
		return nil
	}

	return &Gateway{
		Source:    gw,
		Listeners: buildListeners(gw, secretMemoryMgr, gc),
	}
}
