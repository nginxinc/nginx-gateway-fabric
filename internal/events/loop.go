package events

import (
	"context"
	"fmt"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// EventLoop is the main event loop of the Gateway.
type EventLoop struct {
	conf    state.Configuration
	eventCh <-chan interface{}
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(conf state.Configuration, eventCh <-chan interface{}) *EventLoop {
	return &EventLoop{
		conf:    conf,
		eventCh: eventCh,
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
	// This code is temporary. We will remove it once we have a component that processes changes.
	for _, c := range changes {
		fmt.Println("Processing a change:")
		fmt.Printf("%+v\n", c)
	}

	// This code is temporary. We will remove it once we have a component that updates statuses.
	for _, u := range updates {
		fmt.Println("Processing a status update:")
		fmt.Printf("%+v\n", u)
	}
}
