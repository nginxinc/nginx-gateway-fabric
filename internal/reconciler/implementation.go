package reconciler

import (
	"context"
	"fmt"
	"reflect"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
)

// NamespacedNameFilterFunc is a function that returns true if the resource should be processed by the reconciler.
// If the function returns false, the reconciler will log the returned string.
type NamespacedNameFilterFunc func(nsname types.NamespacedName) (bool, string)

// Config contains the configuration for the Implementation.
type Config struct {
	// Getter gets a resource from the k8s API.
	Getter Getter
	// ObjectType is the type of the resource that the reconciler will reconcile.
	ObjectType client.Object
	// EventCh is the channel where the reconciler will send events.
	EventCh chan<- interface{}
	// NamespacedNameFilter filters resources the controller will process. Can be nil.
	NamespacedNameFilter NamespacedNameFilterFunc
}

// Implementation is a reconciler for Kubernetes resources.
// It implements the reconcile.Reconciler interface.
// A successful reconciliation of a resource has the two possible outcomes:
// (1) If the resource is deleted, the Implementation will send a DeleteEvent to the event channel.
// (2) If the resource is upserted (created or updated), the Implementation will send an UpsertEvent
// to the event channel.
type Implementation struct {
	cfg Config
}

var _ reconcile.Reconciler = &Implementation{}

// NewImplementation creates a new Implementation.
func NewImplementation(cfg Config) *Implementation {
	return &Implementation{
		cfg: cfg,
	}
}

func newObject(objectType client.Object) client.Object {
	// without Elem(), t will be a pointer to the type. For example, *v1beta1.Gateway, not v1beta1.Gateway
	t := reflect.TypeOf(objectType).Elem()

	return reflect.New(t).Interface().(client.Object)
}

// Reconcile implements the reconcile.Reconciler Reconcile method.
func (r *Implementation) Reconcile(ctx context.Context, req reconcile.Request) (reconcile.Result, error) {
	logger := log.FromContext(ctx)
	// The controller runtime has set the logger with the group, kind, namespace and name of the resource,
	// and a few other key/value pairs. So we don't need to set them here.

	logger.Info("Reconciling the resource")

	found := true
	obj := newObject(r.cfg.ObjectType)
	err := r.cfg.Getter.Get(ctx, req.NamespacedName, obj)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			logger.Error(err, "Failed to get the resource")
			return reconcile.Result{}, err
		}
		found = false
	}

	if r.cfg.NamespacedNameFilter != nil {
		if allow, msg := r.cfg.NamespacedNameFilter(req.NamespacedName); !allow {
			logger.Info(msg)
			return reconcile.Result{}, nil
		}
	}

	var e interface{}
	var operation string

	if !found {
		e = &events.DeleteEvent{
			Type:           r.cfg.ObjectType,
			NamespacedName: req.NamespacedName,
		}
		operation = "deleted"
	} else {
		e = &events.UpsertEvent{
			Resource: obj,
		}
		operation = "upserted"
	}

	select {
	case <-ctx.Done():
		logger.Info(fmt.Sprintf("The resource was not %s because the context was canceled", operation))
	case r.cfg.EventCh <- e:
		logger.Info(fmt.Sprintf("The resource was %s", operation))
	}

	return reconcile.Result{}, nil
}
