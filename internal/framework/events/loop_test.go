package events_test

import (
	"context"
	"errors"

	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events/eventsfakes"
)

var _ = Describe("EventLoop", func() {
	var (
		fakeHandler  *eventsfakes.FakeEventHandler
		eventCh      chan interface{}
		fakePreparer *eventsfakes.FakeFirstEventBatchPreparer
		eventLoop    *events.EventLoop
		ctx          context.Context
		cancel       context.CancelFunc
		errorCh      chan error
	)

	BeforeEach(func() {
		fakeHandler = &eventsfakes.FakeEventHandler{}
		eventCh = make(chan interface{})
		fakePreparer = &eventsfakes.FakeFirstEventBatchPreparer{}

		eventLoop = events.NewEventLoop(eventCh, zap.New(), fakeHandler, fakePreparer)

		ctx, cancel = context.WithCancel(context.Background())
		errorCh = make(chan error)
	})

	Describe("Normal processing", func() {
		BeforeEach(func() {
			batch := events.EventBatch{
				"event0",
			}
			fakePreparer.PrepareReturns(batch, nil)

			go func() {
				errorCh <- eventLoop.Start(ctx)
			}()

			// Ensure  the first batch is handled
			Eventually(fakeHandler.HandleEventBatchCallCount).Should(Equal(1))
			_, _, batch = fakeHandler.HandleEventBatchArgsForCall(0)

			var expectedBatch events.EventBatch = []interface{}{"event0"}
			Expect(batch).Should(Equal(expectedBatch))
		})

		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).ToNot(HaveOccurred())
		})

		// Because BeforeEach() creates the first batch and waits for it to be handled, in the tests below
		// HandleEventBatchCallCount() is already 1.

		It("should process a single event", func() {
			e := "event"

			eventCh <- e

			Eventually(fakeHandler.HandleEventBatchCallCount).Should(Equal(2))
			_, _, batch := fakeHandler.HandleEventBatchArgsForCall(1)

			var expectedBatch events.EventBatch = []interface{}{e}
			Expect(batch).Should(Equal(expectedBatch))
		})

		It("should batch multiple events", func() {
			firstHandleEventBatchCallInProgress := make(chan struct{})
			sentSecondAndThirdEvents := make(chan struct{})

			// The func below will pause the handler goroutine while it is processing the batch with e1 until
			// sentSecondAndThirdEvents is closed. This way we can add e2 and e3 to the current batch in the meantime.
			fakeHandler.HandleEventBatchCalls(func(_ context.Context, _ logr.Logger, _ events.EventBatch) {
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

			Eventually(fakeHandler.HandleEventBatchCallCount).Should(Equal(3))
			_, _, batch := fakeHandler.HandleEventBatchArgsForCall(1)

			var expectedBatch events.EventBatch = []interface{}{e1}

			// the first HandleEventBatch() call must have handled a batch with e1
			Expect(batch).Should(Equal(expectedBatch))

			_, _, batch = fakeHandler.HandleEventBatchArgsForCall(2)

			expectedBatch = []interface{}{e2, e3}
			// the second HandleEventBatch() call must have handled a batch with e2 and e3
			Expect(batch).Should(Equal(expectedBatch))
		})
	})

	Describe("Edge cases", func() {
		It("should return error when preparer returns error without blocking", func() {
			preparerError := errors.New("test")
			fakePreparer.PrepareReturns(events.EventBatch{}, preparerError)

			err := eventLoop.Start(ctx)

			Expect(err).Should(MatchError(preparerError))
		})

		It("should return nil when started with canceled context without blocking", func() {
			fakePreparer.PrepareReturns(events.EventBatch{}, nil)

			cancel()
			err := eventLoop.Start(ctx)

			Expect(err).ToNot(HaveOccurred())
		})
	})
})
