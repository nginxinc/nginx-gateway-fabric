package events

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEventLoop_SwapBatches(t *testing.T) {
	eventLoop := NewEventLoop(nil, zap.New(), nil, nil)

	eventLoop.currentBatch = EventBatch{
		"event0",
		"event1",
		"event2",
	}

	nextBatch := EventBatch{
		"event3",
		"event4",
		"event5",
		"event6",
	}

	eventLoop.nextBatch = nextBatch

	eventLoop.swapBatches()

	if l := len(eventLoop.currentBatch); l != 4 {
		t.Errorf("EventLoop.swapBatches() mismatch. Expected 4 events in the current batch, got %d", l)
	}

	if diff := cmp.Diff(eventLoop.currentBatch, nextBatch); diff != "" {
		t.Errorf("EventLoop.swapBatches() mismatch on current batch events (-want +got):\n%s", diff)
	}

	if l := len(eventLoop.nextBatch); l != 0 {
		t.Errorf("EventLoop.swapBatches() mismatch. Expected 0 events in the next batch, got %d", l)
	}

	if c := cap(eventLoop.nextBatch); c != 3 {
		t.Errorf("EventLoop.swapBatches() mismatch. Expected capacity of 3 in the next batch, got %d", c)
	}
}
