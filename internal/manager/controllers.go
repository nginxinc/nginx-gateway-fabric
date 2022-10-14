package manager

import (
	"context"
	"fmt"
	"time"

	ctlr "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

const (
	// addIndexFieldTimeout is the timeout used for adding an Index Field to a cache.
	addIndexFieldTimeout = 2 * time.Minute
)

type newReconcilerFunc func(cfg reconciler.Config) *reconciler.Implementation

type controllerConfig struct {
	namespacedNameFilter reconciler.NamespacedNameFilterFunc
	k8sPredicate         predicate.Predicate
	fieldIndices         index.FieldIndices
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

func withFieldIndices(fieldIndices index.FieldIndices) controllerOption {
	return func(cfg *controllerConfig) {
		cfg.fieldIndices = fieldIndices
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
	eventCh chan interface{},
	options ...controllerOption,
) error {
	cfg := defaultControllerConfig()

	for _, opt := range options {
		opt(&cfg)
	}

	for field, indexerFunc := range cfg.fieldIndices {
		err := addIndex(ctx, mgr.GetFieldIndexer(), objectType, field, indexerFunc)
		if err != nil {
			return err
		}
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

func addIndex(
	ctx context.Context,
	indexer client.FieldIndexer,
	objectType client.Object,
	field string,
	indexerFunc client.IndexerFunc,
) error {
	c, cancel := context.WithTimeout(ctx, addIndexFieldTimeout)
	defer cancel()

	err := indexer.IndexField(c, objectType, field, indexerFunc)
	if err != nil {
		return fmt.Errorf("failed to add index for %T for field %s: %w", objectType, field, err)
	}

	return nil
}
