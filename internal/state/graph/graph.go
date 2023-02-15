package graph

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
)

// ClusterStore includes cluster resources necessary to build the Graph.
type ClusterStore struct {
	GatewayClass *v1beta1.GatewayClass
	Gateways     map[types.NamespacedName]*v1beta1.Gateway
	HTTPRoutes   map[types.NamespacedName]*v1beta1.HTTPRoute
	Services     map[types.NamespacedName]*v1.Service
}

// Graph is a Graph-like representation of Gateway API resources.
type Graph struct {
	// GatewayClass holds the GatewayClass resource.
	GatewayClass *GatewayClass
	// Gateway holds the winning Gateway resource.
	Gateway *Gateway
	// IgnoredGateways holds the ignored Gateway resources, which belong to the NGINX Gateway (based on the
	// GatewayClassName field of the resource) but ignored. It doesn't hold the Gateway resources that do not belong to
	// the NGINX Gateway.
	IgnoredGateways map[types.NamespacedName]*v1beta1.Gateway
	// Routes holds Route resources.
	Routes map[types.NamespacedName]*Route
}

// BuildGraph builds a Graph from a store.
func BuildGraph(
	store ClusterStore,
	controllerName string,
	gcName string,
	secretMemoryMgr secrets.SecretDiskMemoryManager,
) *Graph {
	gc := buildGatewayClass(store.GatewayClass, controllerName)

	gw, ignoredGws := processGateways(store.Gateways, gcName)

	listeners := buildListeners(gw, gcName, secretMemoryMgr)

	routes := make(map[types.NamespacedName]*Route)
	for _, ghr := range store.HTTPRoutes {
		ignored, r := bindHTTPRouteToListeners(ghr, gw, ignoredGws, listeners)
		if !ignored {
			routes[client.ObjectKeyFromObject(ghr)] = r
		}
	}

	addBackendGroupsToRoutes(routes, store.Services)

	g := &Graph{
		GatewayClass:    gc,
		Routes:          routes,
		IgnoredGateways: ignoredGws,
	}

	if gw != nil {
		g.Gateway = &Gateway{
			Source:    gw,
			Listeners: listeners,
		}
	}

	return g
}
