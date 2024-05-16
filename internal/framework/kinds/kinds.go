package kinds

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Gateway API Kinds
const (
	// Gateway is the Gateway Kind
	Gateway = "Gateway"
	// GatewayClass is the GatewayClass Kind.
	GatewayClass = "GatewayClass"
	// HTTPRoute is the HTTPRoute kind.
	HTTPRoute = "HTTPRoute"
	// GRPCRoute is the GRPCRoute kind.
	GRPCRoute = "GRPCRoute"
)

// NGINX Gateway Fabric kinds.
const (
	// ClientSettingsPolicy is the ClientSettingsPolicy kind.
	ClientSettingsPolicy = "ClientSettingsPolicy"
	// NginxProxy is the NginxProxy kind.
	NginxProxy = "NginxProxy"
)

// MustExtractGVK is a function that extracts the GroupVersionKind (GVK) of a client.object.
// It will panic if the GKV cannot be extracted.
type MustExtractGVK func(object client.Object) schema.GroupVersionKind
