package controller

import (
	"context"
	"fmt"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

// MainController is the main control loop of the Gateway.
type MainController struct {
	conf    state.Configuration
	eventCh <-chan interface{}
}

// NewMainController creates a new MainController.
func NewMainController(conf state.Configuration, eventCh <-chan interface{}) *MainController {
	return &MainController{
		conf:    conf,
		eventCh: eventCh,
	}
}

// Start starts the MainController.
// The method will block until the MainController stops:
// - if it stops because of an error, the Start will return the error.
// - if it stops normally, the Start will return nil.
func (c *MainController) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-c.eventCh:
			err := c.handleEvent(e)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *MainController) handleEvent(event interface{}) error {
	var changes []state.Change
	var updates []state.StatusUpdate
	var err error

	switch e := event.(type) {
	case *UpsertEvent:
		changes, updates, err = c.propagateUpsert(e)
	case *DeleteEvent:
		changes, updates, err = c.propagateDelete(e)
	default:
		return fmt.Errorf("unknown event type %T", e)
	}

	if err != nil {
		return err
	}

	c.processChangesAndStatusUpdates(changes, updates)

	return nil
}

func (c *MainController) propagateUpsert(e *UpsertEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch r := e.Resource.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := c.conf.UpsertHTTPRoute(r)
		return changes, statusUpdates, nil
	}

	return nil, nil, fmt.Errorf("unknown resource type %T", e.Resource)
}

func (c *MainController) propagateDelete(e *DeleteEvent) ([]state.Change, []state.StatusUpdate, error) {
	switch e.Type.(type) {
	case *v1alpha2.HTTPRoute:
		changes, statusUpdates := c.conf.DeleteHTTPRoute(e.NamespacedName)
		return changes, statusUpdates, nil
	}

	return nil, nil, fmt.Errorf("unknown resource type %T", e.Type)
}

func (c *MainController) processChangesAndStatusUpdates(changes []state.Change, updates []state.StatusUpdate) {
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
