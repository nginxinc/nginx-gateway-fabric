package events

import (
	"context"
	"fmt"
	"sync"

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

	// lock protects the current batch
	lock sync.Mutex
	// the current batch
	batch EventBatch
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
	// We will use this cond to signal the handler goroutine to handle the current batch.
	cond := sync.NewCond(&el.lock)

	isShuttingDown := func() bool {
		select {
		case <-ctx.Done():
			return true
		default:
			return false
		}
	}

	// The handler goroutine will use this channel to signal its termination.
	handlerTerminated := make(chan struct{})

	// This is the handler goroutine.
	// It will handle the current batch once it has at least one event.
	// It will run until the eventLoop is shutting down.
	go func() {
		for {
			cond.L.Lock()

			// If the current batch has zero events and the eventLoop isn't shutting down, sleep until awaken.
			for len(el.batch) == 0 && !isShuttingDown() {
				cond.Wait()
			}

			if isShuttingDown() {
				// Terminate the goroutine if the eventLoop is shutting down.
				cond.L.Unlock()

				close(handlerTerminated)
				return
			}

			oldBatch := el.batch
			// replace the current batch with a new empty one
			el.batch = make([]interface{}, 0)

			cond.L.Unlock()

			el.handler.HandleEventBatch(ctx, oldBatch)
		}
	}()

	// The event loop
	for {
		select {
		case <-ctx.Done():
			// Locking here is needed to make sure the signal is only sent either when the handler goroutine is
			// sleeping or when it is handling the previous batch before acquiring the lock again.
			// This way we know the signal will not get lost.
			cond.L.Lock()

			// Signal the handler goroutine so that it terminates. Two possible outcomes:
			// - If the handler is sleeping, it will wake up and terminate.
			// - If the handler is busy handling the previous batch, it will continue handling it. Once it is done, it
			// will terminate.
			cond.Signal()
			cond.L.Unlock()

			// Wait until the handler goroutine terminates.
			<-handlerTerminated

			// although Start() always returns nil as an error, we cannot remove the error from the returned parameters,
			// because "sigs.k8s.io/controller-runtime/pkg/manager".Runnable requires it, which Start() must satisfy.
			return nil
		case e := <-el.eventCh:
			// Add the new event to the current batch.
			cond.L.Lock()

			el.batch = append(el.batch, e)

			// FIXME(pleshakov): Log more details about the event like resource GVK and ns/name.
			el.logger.Info(
				"added an event to the current batch",
				"type", fmt.Sprintf("%T", e),
				"total", len(el.batch),
			)

			// Signal the handler goroutine to handle the current batch. Two possible outcomes:
			// - If the handler is sleeping, it will wake up and handle the current batch.
			// - If the handler is busy handling the previous batch, it will continue handling it. Once it is done, it
			// will handle the current batch without sleeping.
			cond.Signal()
			cond.L.Unlock()
		}
	}
}

// CurrentEventBatchLen returns the length of the current event batch.
// It was added to simplify testing of the EventLoop.
func (el *EventLoop) CurrentEventBatchLen() int {
	el.lock.Lock()
	defer el.lock.Unlock()

	return len(el.batch)
}
