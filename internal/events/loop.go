package events

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
)

// EventLoop is the main event loop of the Gateway. It handles events coming through the event channel.
//
// When a new event comes, there are two cases:
// - If there is no event(s) currently being handled, the new event is handled immediately.
// - Otherwise, the new event will be saved for later handling. All saved events will be handled after the handling of
// the current event(s) finishes. Multiple saved events will be handled at once -- they will be batched.
//
// Batching is needed because, because typically handling an event (or multiple events at once) will result into
// reloading NGINX, which is the operation we want to minimize, for the following reasons:
// (1) A reload takes time - at least 200ms. The time depends on the size of the configuration including the number of
// TLS certs, available CPU cycles.
// (2) A reload can have side-effects for the data plane traffic.
// FIXME(pleshakov): better document the side effects and how to prevent and mitigate them.
// So when the EventLoop have 100 saved events, it is better to process them at once rather than one by one.
type EventLoop struct {
	eventCh <-chan interface{}
	logger  logr.Logger
	handler EventHandler
}

// NewEventLoop creates a new EventLoop.
func NewEventLoop(eventCh <-chan interface{}, logger logr.Logger, handler EventHandler) *EventLoop {
	return &EventLoop{
		eventCh: eventCh,
		logger:  logger,
		handler: handler,
	}
}

// Start starts the EventLoop.
// This method will block until the EventLoop stops, which will happen after the ctx is closed.
//
// FIXME(pleshakov). Ensure that when the Gateway starts, the first time it generates configuration for NGINX,
// it has a complete view of the cluster resources. For example, when the Gateway processes a Gateway resource
// with a listener with TLS termination enabled (the listener references a TLS Secret), the Gateway knows about the secret.
// This way the Gateway will not produce any incomplete transient configuration at the start.
func (el *EventLoop) Start(ctx context.Context) error {
	// The current batch.
	var batch EventBatch
	// handling tells if any batch is currently being handled.
	var handling bool
	// handlingDone is used to signal the completion of handling a batch.
	handlingDone := make(chan struct{})

	handle := func(ctx context.Context, batch EventBatch) {
		el.logger.Info("Handling events from the batch", "total", len(batch))

		el.handler.HandleEventBatch(ctx, batch)

		el.logger.Info("Finished handling the batch")

		handlingDone <- struct{}{}
	}

	// Note: at any point of time, no more than one batch is currently being handled.

	// The event loop
	for {
		select {
		case <-ctx.Done():
			// Wait for the completion if a batch is being handled.
			if handling {
				<-handlingDone
			}
			return nil
		case e := <-el.eventCh:
			// Add the event to the current batch.
			batch = append(batch, e)

			// FIXME(pleshakov): Log more details about the event like resource GVK and ns/name.
			el.logger.Info(
				"added an event to the current batch",
				"type", fmt.Sprintf("%T", e),
				"total", len(batch),
			)

			// Handle the current batch if no batch is being handled.
			if !handling {
				go handle(ctx, batch)
				batch = make([]interface{}, 0)
				handling = true
			}
		case <-handlingDone:
			handling = false

			// Handle the current batch if it has at least one event.
			if len(batch) > 0 {
				go handle(ctx, batch)
				batch = make([]interface{}, 0)
				handling = true
			}
		}
	}
}
