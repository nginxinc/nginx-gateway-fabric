package events_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events/eventsfakes"
)

var _ = Describe("EventLoop", func() {
	var (
		fakeHandler *eventsfakes.FakeEventHandler
		eventCh     chan interface{}
		eventLoop   *events.EventLoop
		cancel      context.CancelFunc
		errorCh     chan error
	)

	BeforeEach(func() {
		fakeHandler = &eventsfakes.FakeEventHandler{}
		eventCh = make(chan interface{})

		eventLoop = events.NewEventLoop(eventCh, zap.New(), fakeHandler)

		var ctx context.Context
		ctx, cancel = context.WithCancel(context.Background())
		errorCh = make(chan error)

		go func() {
			errorCh <- eventLoop.Start(ctx)
		}()
	})

	AfterEach(func() {
		cancel()

		var err error
		Eventually(errorCh).Should(Receive(&err))
		Expect(err).To(BeNil())
	})

	It("should process a single event", func() {
		e := "event"

		eventCh <- e

		Eventually(fakeHandler.HandleEventBatchCallCount).Should(Equal(1))
		_, batch := fakeHandler.HandleEventBatchArgsForCall(0)

		var expectedBatch events.EventBatch = []interface{}{e}
		Expect(batch).Should(Equal(expectedBatch))
	})

	It("should batch multiple events", func() {
		firstHandleEventBatchCallInProgress := make(chan struct{})
		sentSecondAndThirdEvents := make(chan struct{})

		// The func below will pause the handler goroutine while it is processing the batch with e1 until
		// sentSecondAndThirdEvents is closed. This way we can add e2 and e3 to the current batch in the meantime.
		fakeHandler.HandleEventBatchCalls(func(ctx context.Context, batch events.EventBatch) {
			close(firstHandleEventBatchCallInProgress)
			<-sentSecondAndThirdEvents
		})

		e1 := "event1"
		e2 := "event2"
		e3 := "event3"

		eventCh <- e1

		// Making sure the handler goroutine started handling the batch with e1.
		<-firstHandleEventBatchCallInProgress

		eventCh <- e2
		eventCh <- e3
		// The event loop will add the e2 and e3 event to current batch before starting another handler goroutine.

		fakeHandler.HandleEventBatchCalls(nil)

		// Unpause the handler goroutine so that it can handle the current batch.
		close(sentSecondAndThirdEvents)

		Eventually(fakeHandler.HandleEventBatchCallCount).Should(Equal(2))
		_, batch := fakeHandler.HandleEventBatchArgsForCall(0)

		var expectedBatch events.EventBatch = []interface{}{e1}

		// the first HandleEventBatch() call must have handled a batch with e1
		Expect(batch).Should(Equal(expectedBatch))

		_, batch = fakeHandler.HandleEventBatchArgsForCall(1)

		expectedBatch = []interface{}{e2, e3}
		// the second HandleEventBatch() call must have handled a batch with e2 and e3
		Expect(batch).Should(Equal(expectedBatch))
	})
})
