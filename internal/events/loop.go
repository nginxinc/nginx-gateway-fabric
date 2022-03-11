package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/status"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// EventLoop is the main event loop of the Gateway.
type EventLoop struct {
	conf          state.Configuration
	serviceStore  state.ServiceStore
	eventCh       <-chan interface{}
	logger        logr.Logger
	statusUpdater status.Updater
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(conf state.Configuration, serviceStore state.ServiceStore, eventCh <-chan interface{},
	statusUpdater status.Updater, logger logr.Logger) *EventLoop {
	return &EventLoop{
		conf:          conf,
		serviceStore:  serviceStore,
		eventCh:       eventCh,
		statusUpdater: statusUpdater,
		logger:        logger.WithName("eventLoop"),
	}
}

// Start starts the EventLoop.
// The method will block until the EventLoop stops:
// - if it stops because of an error, the Start will return the error.
// - if it stops normally, the Start will return nil.
func (el *EventLoop) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-el.eventCh:
			err := el.handleEvent(ctx, e)
			if err != nil {
				return err
			}
		}
	}
}

// TO-DO: think about how to avoid using an interface{} here
func (el *EventLoop) handleEvent(ctx context.Context, event interface{}) error {
	var changes []state.Change
	var updates []state.StatusUpdate
	var err error

	switch e := event.(type) {
	case *UpsertEvent:
		changes, updates, err = el.propagateUpsert(e)
	case *DeleteEvent:
		changes, updates, err = el.propagateDelete(e)
	default:
		// TO-DO: panic
		return fmt.Errorf("unknown event type %T", e)
	}

	if err != nil {
		return err
	}

	el.processChangesAndStatusUpdates(ctx, changes, updates)
	return nil
}

func (el *EventLoop) propagateUpsert(e *UpsertEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch r := e.Resource.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := el.conf.UpsertHTTPRoute(r)
		return changes, statusUpdates, nil
	case *apiv1.Service:
		el.serviceStore.Upsert(r)
		// TO-DO: make sure the affected hosts are updated
		return nil, nil, nil
	}

	// TO-DO: panic
	return nil, nil, fmt.Errorf("unknown resource type %T", e.Resource)
}

func (el *EventLoop) propagateDelete(e *DeleteEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch e.Type.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := el.conf.DeleteHTTPRoute(e.NamespacedName)
		return changes, statusUpdates, nil
	case *apiv1.Service:
		el.serviceStore.Delete(e.NamespacedName)
		// TO-DO: make sure the affected hosts are updated
	}

	// TO-DO: panic
	return nil, nil, fmt.Errorf("unknown resource type %T", e.Type)
}

func (el *EventLoop) processChangesAndStatusUpdates(ctx context.Context, changes []state.Change, updates []state.StatusUpdate) {
	for _, c := range changes {
		el.logger.Info("Processing a change",
			"host", c.Host.Value)

		// TO-DO: This code is temporary. We will remove it once we have a component that processes changes.
		fmt.Printf("%+v\n", c)

		if c.Op == state.Upsert {
			// The code below resolves service backend refs into their cluster IPs
			// TO-DO: this code will be removed once we have the component that generates NGINX config and
			// uses the ServiceStore to resolve services.
			for _, g := range c.Host.PathRouteGroups {
				for _, r := range g.Routes {
					for _, b := range r.Source.Spec.Rules[r.RuleIdx].BackendRefs {
						if b.BackendRef.Kind == nil || *b.BackendRef.Kind == "Service" {
							ns := r.Source.Namespace
							if b.BackendRef.Namespace != nil {
								ns = string(*b.BackendRef.Namespace)
							}

							address, err := el.serviceStore.Resolve(types.NamespacedName{
								Namespace: ns,
								Name:      string(b.BackendRef.Name),
							})

							if err != nil {
								fmt.Printf("Service %s/%s error: %v\n", ns, b.BackendRef.Name, err)
								continue
							}

							var port int32 = 80
							if b.BackendRef.Port != nil {
								port = int32(*b.BackendRef.Port)
							}

							address = fmt.Sprintf("%s:%d", address, port)

							fmt.Printf("Service %s/%s: %s\n", ns, b.BackendRef.Name, address)
						}
					}
				}
			}
		}
	}

	el.statusUpdater.ProcessStatusUpdates(ctx, updates)
}
