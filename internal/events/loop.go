package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	apiv1 "k8s.io/api/core/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
)

// EventLoop is the main event loop of the Gateway.
type EventLoop struct {
	processor       newstate.ChangeProcessor
	serviceStore    state.ServiceStore
	generator       config.Generator
	eventCh         <-chan interface{}
	logger          logr.Logger
	nginxFileMgr    file.Manager
	nginxRuntimeMgr runtime.Manager
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(
	processor newstate.ChangeProcessor,
	serviceStore state.ServiceStore,
	generator config.Generator,
	eventCh <-chan interface{},
	logger logr.Logger,
	nginxFileMgr file.Manager,
	nginxRuntimeMgr runtime.Manager,
) *EventLoop {
	return &EventLoop{
		processor:       processor,
		serviceStore:    serviceStore,
		generator:       generator,
		eventCh:         eventCh,
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
			// although we always return nil, Start must return it to satisfy
			// "sigs.k8s.io/controller-runtime/pkg/manager".Runnable
			return nil
		case e := <-el.eventCh:
			el.handleEvent(ctx, e)
		}
	}
}

// FIXME(pleshakov): think about how to avoid using an interface{} here
func (el *EventLoop) handleEvent(ctx context.Context, event interface{}) {
	switch e := event.(type) {
	case *UpsertEvent:
		el.propagateUpsert(e)
	case *DeleteEvent:
		el.propagateDelete(e)
	default:
		panic(fmt.Errorf("unknown event type %T", e))
	}

	changed, conf, statuses := el.processor.Process()
	if !changed {
		return
	}

	err := el.updateNginx(ctx, conf)
	if err != nil {
		el.logger.Error(err, "Failed to update NGINX configuration")
	}

	// FIXME(pleshakov) Update resource statuses instead of printing to stdout
	for name, s := range statuses.ListenerStatuses {
		fmt.Printf("Listener %q, Statuses: %v\n", name, s)
	}
	for nsname, s := range statuses.HTTPRouteStatuses {
		fmt.Printf("HTTPRoute %q, Statuses: %v\n", nsname, s)
	}
}

func (el *EventLoop) updateNginx(ctx context.Context, conf newstate.Configuration) error {
	cfg, warnings := el.generator.Generate(conf)

	// For now, we keep all http servers in one config
	// We might rethink that. For example, we can write each server to its file
	// or group servers in some way.
	err := el.nginxFileMgr.WriteHTTPServersConfig("http-servers", cfg)
	if err != nil {
		return err
	}

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

	return el.nginxRuntimeMgr.Reload(ctx)
}

func (el *EventLoop) propagateUpsert(e *UpsertEvent) {
	switch r := e.Resource.(type) {
	case *v1alpha2.Gateway:
		el.processor.CaptureUpsertChange(r)
	case *v1alpha2.HTTPRoute:
		el.processor.CaptureUpsertChange(r)
	case *apiv1.Service:
		// FIXME(pleshakov): make sure the affected hosts are updated
		el.serviceStore.Upsert(r)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Resource))
	}
}

func (el *EventLoop) propagateDelete(e *DeleteEvent) {
	switch e.Type.(type) {
	case *v1alpha2.Gateway:
		el.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *v1alpha2.HTTPRoute:
		el.processor.CaptureDeleteChange(e.Type, e.NamespacedName)
	case *apiv1.Service:
		// FIXME(pleshakov): make sure the affected hosts are updated
		el.serviceStore.Delete(e.NamespacedName)
	default:
		panic(fmt.Errorf("unknown resource type %T", e.Type))
	}
}
