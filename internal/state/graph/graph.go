package graph

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/secrets"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
)

// ClusterState includes cluster resources necessary to build the Graph.
type ClusterState struct {
	GatewayClasses  map[types.NamespacedName]*v1beta1.GatewayClass
	Gateways        map[types.NamespacedName]*v1beta1.Gateway
	HTTPRoutes      map[types.NamespacedName]*v1beta1.HTTPRoute
	Services        map[types.NamespacedName]*v1.Service
	Namespaces      map[types.NamespacedName]*v1.Namespace
	ReferenceGrants map[types.NamespacedName]*v1beta1.ReferenceGrant
}

// Graph is a Graph-like representation of Gateway API resources.
type Graph struct {
	// GatewayClass holds the GatewayClass resource.
	GatewayClass *GatewayClass
	// Gateway holds the winning Gateway resource.
	Gateway *Gateway
	// IgnoredGatewayClasses holds the ignored GatewayClass resources, which reference NGINX Gateway in the
	// controllerName, but are not configured via the NGINX Gateway CLI argument. It doesn't hold the GatewayClass
	// resources that do not belong to the NGINX Gateway.
	IgnoredGatewayClasses map[types.NamespacedName]*v1beta1.GatewayClass
	// IgnoredGateways holds the ignored Gateway resources, which belong to the NGINX Gateway (based on the
	// GatewayClassName field of the resource) but ignored. It doesn't hold the Gateway resources that do not belong to
	// the NGINX Gateway.
	IgnoredGateways map[types.NamespacedName]*v1beta1.Gateway
	// Routes holds Route resources.
	Routes map[types.NamespacedName]*Route
}

// BuildGraph builds a Graph from a state.
func BuildGraph(
	state ClusterState,
	controllerName string,
	gcName string,
	secretMemoryMgr secrets.SecretDiskMemoryManager,
	validators validation.Validators,
) *Graph {
	processedGwClasses, gcExists := processGatewayClasses(state.GatewayClasses, gcName, controllerName)
	if gcExists && processedGwClasses.Winner == nil {
		// configured GatewayClass does not reference this controller
		return &Graph{}
	}
	gc := buildGatewayClass(processedGwClasses.Winner)

	processedGws := processGateways(state.Gateways, gcName)

	refGrantResolver := newReferenceGrantResolver(state.ReferenceGrants)
	gw := buildGateway(processedGws.Winner, secretMemoryMgr, gc, refGrantResolver)

	routes := buildRoutesForGateways(validators.HTTPFieldsValidator, state.HTTPRoutes, processedGws.GetAllNsNames())
	bindRoutesToListeners(routes, gw, state.Namespaces)
	addBackendRefsToRouteRules(routes, refGrantResolver, state.Services)

	g := &Graph{
		GatewayClass:          gc,
		Gateway:               gw,
		Routes:                routes,
		IgnoredGatewayClasses: processedGwClasses.Ignored,
		IgnoredGateways:       processedGws.Ignored,
	}

	return g
}
