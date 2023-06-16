package graph

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/validation"
)

// ClusterState includes cluster resources necessary to build the Graph.
type ClusterState struct {
	GatewayClasses map[types.NamespacedName]*v1beta1.GatewayClass
	Gateways       map[types.NamespacedName]*v1beta1.Gateway
	HTTPRoutes     map[types.NamespacedName]*v1beta1.HTTPRoute
	Services       map[types.NamespacedName]*v1.Service
	Namespaces     map[types.NamespacedName]*v1.Namespace
	Secrets        map[types.NamespacedName]*v1.Secret
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
	// Secrets includes Secrets used by the Gateway. The Secret might not exist in the API, but it will be
	// present in the Secrets if a Gateway listener references it.
	Secrets map[types.NamespacedName]*Secret
}

// IncludesResource returns true if the Graph includes the resource.
func (g *Graph) IncludesResource(objType client.Object, nsname types.NamespacedName) bool {
	// only works for Secrets for now, but we can add more types later to reduce the number of reloads.
	_, ok := objType.(*v1.Secret)
	if !ok {
		return false
	}

	_, exists := g.Secrets[nsname]
	return exists
}

// BuildGraph builds a Graph from a state.
func BuildGraph(
	state ClusterState,
	controllerName string,
	gcName string,
	validators validation.Validators,
) *Graph {
	gatewayClass := state.GatewayClasses[types.NamespacedName{Name: gcName}]

	if !gatewayClassBelongsToController(gatewayClass, controllerName) {
		return &Graph{}
	}

	gc := buildGatewayClass(gatewayClass)

	processedGws := processGateways(state.Gateways, gcName)

	secretResolver := newSecretResolver(state.Secrets)

	gw := buildGateway(processedGws.Winner, secretResolver, gc)

	routes := buildRoutesForGateways(validators.HTTPFieldsValidator, state.HTTPRoutes, processedGws.GetAllNsNames())
	bindRoutesToListeners(routes, gw, state.Namespaces)
	addBackendRefsToRouteRules(routes, state.Services)

	g := &Graph{
		GatewayClass:    gc,
		Gateway:         gw,
		Routes:          routes,
		IgnoredGateways: processedGws.Ignored,
		Secrets:         secretResolver.ResolvedSecrets(),
	}

	return g
}
