package events

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func TestEventLoop_SwapBatches(t *testing.T) {
	g := NewWithT(t)
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

	g.Expect(len(eventLoop.currentBatch)).To(Equal(len(nextBatch)))
	g.Expect(eventLoop.currentBatch).To(Equal(nextBatch))
	g.Expect(len(eventLoop.nextBatch)).To(Equal(0))
	g.Expect(cap(eventLoop.nextBatch)).To(Equal(3))
}
