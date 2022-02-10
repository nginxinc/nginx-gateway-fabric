package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// EventLoop is the main event loop of the Gateway.
type EventLoop struct {
	conf    state.Configuration
	eventCh <-chan interface{}
	logger  logr.Logger
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(conf state.Configuration, eventCh <-chan interface{}, logger logr.Logger) *EventLoop {
	return &EventLoop{
		conf:    conf,
		eventCh: eventCh,
		logger:  logger.WithName("eventLoop"),
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
			err := el.handleEvent(e)
			if err != nil {
				return err
			}
		}
	}
}

func (el *EventLoop) handleEvent(event interface{}) error {
	var changes []state.Change
	var updates []state.StatusUpdate
	var err error

	switch e := event.(type) {
	case *UpsertEvent:
		changes, updates, err = el.propagateUpsert(e)
	case *DeleteEvent:
		changes, updates, err = el.propagateDelete(e)
	default:
		return fmt.Errorf("unknown event type %T", e)
	}

	if err != nil {
		return err
	}

	el.processChangesAndStatusUpdates(changes, updates)

	return nil
}

func (el *EventLoop) propagateUpsert(e *UpsertEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch r := e.Resource.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := el.conf.UpsertHTTPRoute(r)
		return changes, statusUpdates, nil
	}

	return nil, nil, fmt.Errorf("unknown resource type %T", e.Resource)
}

func (el *EventLoop) propagateDelete(e *DeleteEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch e.Type.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := el.conf.DeleteHTTPRoute(e.NamespacedName)
		return changes, statusUpdates, nil
	}

	return nil, nil, fmt.Errorf("unknown resource type %T", e.Type)
}

func (el *EventLoop) processChangesAndStatusUpdates(changes []state.Change, updates []state.StatusUpdate) {
	for _, c := range changes {
		el.logger.Info("Processing a change",
			"host", c.Host.Value)

		// TO-DO: This code is temporary. We will remove it once we have a component that processes changes.
		fmt.Printf("%+v\n", c)
	}

	for _, u := range updates {
		// TO-DO: in the next iteration, the update will include the namespace/name of the resource instead of
		// runtime.Object, so it will be easy to get the resource namespace/name and include it in the log output
		el.logger.Info("Processing a status update",
			"gvk", u.Object.GetObjectKind().GroupVersionKind().String())

		// TO-DO: This code is temporary. We will remove it once we have a component that updates statuses.
		fmt.Printf("%+v\n", u)
	}
}
