package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

// EventLoop is the main event loop of the Gateway.
type EventLoop struct {
	conf            state.Configuration
	serviceStore    state.ServiceStore
	generator       config.Generator
	eventCh         <-chan interface{}
	logger          logr.Logger
	statusUpdater   status.Updater
	nginxFileMgr    file.Manager
	nginxRuntimeMgr runtime.Manager
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(
	conf state.Configuration,
	serviceStore state.ServiceStore,
	generator config.Generator,
	eventCh <-chan interface{},
	statusUpdater status.Updater,
	logger logr.Logger,
	nginxFileMgr file.Manager,
	nginxRuntimeMgr runtime.Manager,
) *EventLoop {
	return &EventLoop{
		conf:            conf,
		serviceStore:    serviceStore,
		generator:       generator,
		eventCh:         eventCh,
		statusUpdater:   statusUpdater,
		logger:          logger.WithName("eventLoop"),
		nginxFileMgr:    nginxFileMgr,
		nginxRuntimeMgr: nginxRuntimeMgr,
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

// FIXME(pleshakov): think about how to avoid using an interface{} here
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
		// FIXME(pleshakov): panic because it is a coding error
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
		// FIXME(pleshakov): make sure the affected hosts are updated
		return nil, nil, nil
	}

	// FIXME(pleshakov): panic because it is a coding error
	return nil, nil, fmt.Errorf("unknown resource type %T", e.Resource)
}

func (el *EventLoop) propagateDelete(e *DeleteEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch e.Type.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := el.conf.DeleteHTTPRoute(e.NamespacedName)
		return changes, statusUpdates, nil
	case *apiv1.Service:
		el.serviceStore.Delete(e.NamespacedName)
		// FIXME(pleshakov): make sure the affected hosts are updated
		return nil, nil, nil
	}

	// FIXME(pleshakov): panic because it is a coding error
	return nil, nil, fmt.Errorf("unknown resource type %T", e.Type)
}

func (el *EventLoop) processChangesAndStatusUpdates(ctx context.Context, changes []state.Change, updates []state.StatusUpdate) {
	for _, c := range changes {
		el.logger.Info("Processing a change",
			"host", c.Host.Value)

		if c.Op == state.Upsert {
			cfg, warnings := el.generator.GenerateForHost(c.Host)

			for obj, objWarnings := range warnings {
				for _, w := range objWarnings {
					// FIXME(pleshakov): report warnings via Object status
					el.logger.Info("got warning while generating config",
						"kind", obj.GetObjectKind().GroupVersionKind().Kind,
						"namespace", obj.GetNamespace(),
						"name", obj.GetName(),
						"warning", w)
				}
			}

			el.logger.Info("Writing configuration",
				"host", c.Host.Value)

			err := el.nginxFileMgr.WriteServerConfig(c.Host.Value, cfg)
			if err != nil {
				el.logger.Error(err, "Failed to write configuration",
					"host", c.Host.Value)
			}
		} else {
			err := el.nginxFileMgr.DeleteServerConfig(c.Host.Value)
			if err != nil {
				el.logger.Error(err, "Failed to delete configuration",
					"host", c.Host.Value)
			}
		}
	}

	if len(changes) > 0 {
		err := el.nginxRuntimeMgr.Reload(ctx)
		if err != nil {
			el.logger.Error(err, "Failed to reload NGINX")
		}
	}

	el.statusUpdater.ProcessStatusUpdates(ctx, updates)
}
