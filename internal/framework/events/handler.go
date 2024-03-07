package events

import (
	"context"

	"github.com/go-logr/logr"
)

//counterfeiter:generate . EventHandler

// EventHandler handles events.
type EventHandler interface {
	// HandleEventBatch handles a batch of events.
	// EventBatch can include duplicated events.
	HandleEventBatch(ctx context.Context, logger logr.Logger, batch EventBatch)
}
