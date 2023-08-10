package config

import (
	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

type Config struct {
	// GatewayCtlrName is the name of this controller.
	GatewayCtlrName string
	// ConfigName is the name of the NginxGateway resource for this controller.
	ConfigName string
	// Logger is the Zap Logger used by all components.
	Logger logr.Logger
	// AtomicLevel is an atomically changeable, dynamic logging level.
	AtomicLevel zap.AtomicLevel
	// GatewayNsName is the namespaced name of a Gateway resource that the Gateway will use.
	// The Gateway will ignore all other Gateway resources.
	GatewayNsName *types.NamespacedName
	// GatewayClassName is the name of the GatewayClass resource that the Gateway will use.
	GatewayClassName string
	// PodIP is the IP address of this Pod.
	PodIP string
	// Namespace is the Namespace of this Pod.
	Namespace string
	// UpdateGatewayClassStatus enables updating the status of the GatewayClass resource.
	UpdateGatewayClassStatus bool
}
