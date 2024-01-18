package config

import (
	"time"

	"github.com/go-logr/logr"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

type Config struct {
	// Version is the running NGF version.
	Version string
	// AtomicLevel is an atomically changeable, dynamic logging level.
	AtomicLevel zap.AtomicLevel
	// GatewayNsName is the namespaced name of a Gateway resource that the Gateway will use.
	// The Gateway will ignore all other Gateway resources.
	GatewayNsName *types.NamespacedName
	// GatewayPodConfig contains information about this Pod.
	GatewayPodConfig GatewayPodConfig
	// Logger is the Zap Logger used by all components.
	Logger logr.Logger
	// GatewayCtlrName is the name of this controller.
	GatewayCtlrName string
	// ConfigName is the name of the NginxGateway resource for this controller.
	ConfigName string
	// GatewayClassName is the name of the GatewayClass resource that the Gateway will use.
	GatewayClassName string
	// LeaderElection contains the configuration for leader election.
	LeaderElection LeaderElection
	// MetricsConfig specifies the metrics config.
	MetricsConfig MetricsConfig
	// HealthConfig specifies the health probe config.
	HealthConfig HealthConfig
	// TelemetryReportPeriod is the period at which telemetry reports are sent.
	TelemetryReportPeriod time.Duration
	// UpdateGatewayClassStatus enables updating the status of the GatewayClass resource.
	UpdateGatewayClassStatus bool
	// Plus indicates whether NGINX Plus is being used.
	Plus bool
	// EnableExperimentalFeatures enables experimental features.
	EnableExperimentalFeatures bool
}

// GatewayPodConfig contains information about this Pod.
type GatewayPodConfig struct {
	// PodIP is the IP address of this Pod.
	PodIP string
	// ServiceName is the name of the Service that fronts this Pod.
	ServiceName string
	// Namespace is the namespace of this Pod.
	Namespace string
	// Name is the name of the Pod.
	Name string
}

// MetricsConfig specifies the metrics config.
type MetricsConfig struct {
	// Port is the port the metrics should be exposed on.
	Port int
	// Enabled is the flag for toggling metrics on or off.
	Enabled bool
	// Secure is the flag for toggling the metrics endpoint to https.
	Secure bool
}

// HealthConfig specifies the health probe config.
type HealthConfig struct {
	// Port is the port that the health probe server listens on.
	Port int
	// Enabled is the flag for toggling the health probe server on or off.
	Enabled bool
}

// LeaderElection contains the configuration for leader election.
type LeaderElection struct {
	// LockName holds the name of the leader election lock.
	LockName string
	// Identity is the unique name of the controller used for identifying the leader.
	Identity string
	// Enabled indicates whether leader election is enabled.
	Enabled bool
}
