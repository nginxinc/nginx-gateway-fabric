package main

import (
	"context"
	"fmt"

	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

type newReconcilerFunc func(cfg reconciler.Config) *reconciler.Implementation

type controllerConfig struct {
	namespacedNameFilter reconciler.NamespacedNameFilterFunc
	k8sPredicate         predicate.Predicate
	newReconciler        newReconcilerFunc
}

type controllerOption func(*controllerConfig)

func withNamespacedNameFilter(filter reconciler.NamespacedNameFilterFunc) controllerOption {
	return func(cfg *controllerConfig) {
		cfg.namespacedNameFilter = filter
	}
}

func withK8sPredicate(p predicate.Predicate) controllerOption {
	return func(cfg *controllerConfig) {
		cfg.k8sPredicate = p
	}
}

// withNewReconciler allows us to mock reconciler creation in the unit tests.
func withNewReconciler(newReconciler newReconcilerFunc) controllerOption {
	return func(cfg *controllerConfig) {
		cfg.newReconciler = newReconciler
	}
}

func defaultControllerConfig() controllerConfig {
	return controllerConfig{
		newReconciler: reconciler.NewImplementation,
	}
}

func registerController(
	ctx context.Context,
	objectType client.Object,
	mgr manager.Manager,
	eventCh chan<- interface{},
	options ...controllerOption,
) error {
	cfg := defaultControllerConfig()

	for _, opt := range options {
		opt(&cfg)
	}

	builder := ctlr.NewControllerManagedBy(mgr).For(objectType)

	if cfg.k8sPredicate != nil {
		builder = builder.WithEventFilter(cfg.k8sPredicate)
	}

	recCfg := reconciler.Config{
		Getter:               mgr.GetClient(),
		ObjectType:           objectType,
		EventCh:              eventCh,
		NamespacedNameFilter: cfg.namespacedNameFilter,
	}

	err := builder.Complete(cfg.newReconciler(recCfg))
	if err != nil {
		return fmt.Errorf("cannot build a controller for %T: %w", objectType, err)
	}

	return nil
}
