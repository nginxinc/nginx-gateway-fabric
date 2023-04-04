package agent

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/go-logr/logr"
	"github.com/nginx/agent/sdk/v2/proto"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/observer"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/dataplane"
)

// NginxConfig is an intermediate object that contains nginx configuration in a form that agent expects.
// We convert the dataplane configuration to NginxConfig in the config store, so we only need to do it once
// per configuration change. The NginxConfig is then used by the agent to generate the nginx configuration payload.
type NginxConfig struct {
	ID          string
	Config      *proto.ZippedFile
	Aux         *proto.ZippedFile
	Directories []*proto.Directory
}

// ConfigStore stores accepts the latest dataplane configuration and stores is as NginxConfig.
// ConfigStore implements the observer.Subject interface,
// so that it can notify the agent observers when the configuration changes.
// ConfigStore is thread-safe.
type ConfigStore struct {
	latestConfig  atomic.Value
	configBuilder *NginxConfigBuilder
	logger        logr.Logger
	observers     []observer.Observer
	observerLock  sync.Mutex
}

// NewConfigStore creates a new ConfigStore.
func NewConfigStore(configBuilder *NginxConfigBuilder, logger logr.Logger) *ConfigStore {
	return &ConfigStore{
		observers:     make([]observer.Observer, 0),
		configBuilder: configBuilder,
		logger:        logger,
	}
}

// Register registers an observer.
func (a *ConfigStore) Register(observer observer.Observer) {
	a.observerLock.Lock()
	defer a.observerLock.Unlock()

	a.observers = append(a.observers, observer)
	a.logger.Info("Registering observer", "number of registered observers", len(a.observers))
}

// Notify notifies all registered observers.
func (a *ConfigStore) Notify() {
	a.observerLock.Lock()
	defer a.observerLock.Unlock()

	a.logger.Info("Notifying observers", "number of registered observers", len(a.observers))
	for _, o := range a.observers {
		o.Update()
	}
}

// Remove removes an observer.
func (a *ConfigStore) Remove(observer observer.Observer) {
	a.observerLock.Lock()
	defer a.observerLock.Unlock()

	for i, o := range a.observers {
		if o == observer {
			a.observers = append(a.observers[:i], a.observers[i+1:]...)
			a.logger.Info("Removed observer", "number of registered observers", len(a.observers))
			return
		}
	}
}

// Store accepts the latest dataplane configuration, builds the NginxConfig from it, and stores it.
// It's possible for an error to occur when building the NginxConfig,
// in which case the error is returned, and the configuration is not stored.
// If the configuration is successfully stored, the observers are notified.
func (a *ConfigStore) Store(configuration dataplane.Configuration) error {
	agentConf, err := a.configBuilder.Build(configuration)
	if err != nil {
		return fmt.Errorf("error building nginx agent configuration: %w", err)
	}

	a.logger.Info("Storing configuration", "config generation", configuration.Generation)

	a.latestConfig.Store(agentConf)
	a.Notify()
	return nil
}

// GetLatestConfig returns the latest NginxConfig.
func (a *ConfigStore) GetLatestConfig() *NginxConfig {
	return a.latestConfig.Load().(*NginxConfig)
}
