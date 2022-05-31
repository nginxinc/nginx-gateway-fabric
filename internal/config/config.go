package config

import (
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
)

type Config struct {
	GatewayCtlrName string
	Logger          logr.Logger
	// GatewayNsName is the namespaced name of a Gateway resource that the Gateway will use.
	// The Gateway will ignore all other Gateway resources.
	GatewayNsName types.NamespacedName
}
