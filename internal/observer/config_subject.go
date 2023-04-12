package observer

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/go-logr/logr"
)

type VersionedConfig interface {
	GetVersion() string
}

// ConfigSubject stores the latest VersionedConfig.
// It implements the Subject interface and can be observed by Observers.
// When a new VersionedConfig is stored, all registered Observers are notified.
type ConfigSubject[T VersionedConfig] struct {
	latestConfig atomic.Value
	logger       logr.Logger
	observers    map[string]Observer
	observerLock sync.Mutex
}

// NewConfigSubject creates a new ConfigSubject.
func NewConfigSubject[T VersionedConfig](logger logr.Logger) *ConfigSubject[T] {
	return &ConfigSubject[T]{
		observers: make(map[string]Observer),
		logger:    logger,
	}
}

// Register registers an observer.
func (a *ConfigSubject[T]) Register(observer Observer) {
	a.observerLock.Lock()
	defer a.observerLock.Unlock()

	a.observers[observer.ID()] = observer
	a.logger.Info(
		fmt.Sprintf("Registering observer %s", observer.ID()),
		"number of registered observers",
		len(a.observers),
	)
}

// Notify notifies all registered observers.
func (a *ConfigSubject[T]) notify() {
	a.observerLock.Lock()
	defer a.observerLock.Unlock()

	a.logger.Info("Notifying observers", "number of registered observers", len(a.observers))
	for _, o := range a.observers {
		o.Update()
	}
}

// Remove removes an observer.
func (a *ConfigSubject[T]) Remove(observer Observer) {
	a.observerLock.Lock()
	defer a.observerLock.Unlock()

	delete(a.observers, observer.ID())
	a.logger.Info(
		fmt.Sprintf("Removing observer %s", observer.ID()),
		"number of registered observers",
		len(a.observers),
	)
}

func (a *ConfigSubject[T]) Update(cfg VersionedConfig) {
	a.logger.Info("Storing configuration", "config version", cfg.GetVersion())

	a.latestConfig.Store(cfg)
	a.notify()
}

// GetLatestConfig returns the current stored config.
func (a *ConfigSubject[T]) GetLatestConfig() T {
	return a.latestConfig.Load().(T)
}
