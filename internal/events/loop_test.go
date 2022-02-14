package events_test

import (
	"context"
	"errors"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/events"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state/statefakes"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/status/statusfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
)

type unsupportedResource struct {
	metav1.ObjectMeta
}

func (r *unsupportedResource) GetObjectKind() schema.ObjectKind {
	return nil
}

func (r *unsupportedResource) DeepCopyObject() runtime.Object {
	return nil
}

var _ = Describe("EventLoop", func() {
	var ctrl *events.EventLoop
	var fakeConf *statefakes.FakeConfiguration
	var fakeUpdater *statusfakes.FakeUpdater
	var cancel context.CancelFunc
	var eventCh chan interface{}
	var errorCh chan error

	BeforeEach(func() {
		fakeConf = &statefakes.FakeConfiguration{}
		eventCh = make(chan interface{})
		fakeUpdater = &statusfakes.FakeUpdater{}
		ctrl = events.NewEventLoop(fakeConf, eventCh, fakeUpdater, zap.New())

		var ctx context.Context

		ctx, cancel = context.WithCancel(context.Background())
		errorCh = make(chan error)

		go func() {
			errorCh <- ctrl.Start(ctx)
		}()
	})

	Describe("Process HTTPRoute events", func() {
		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).To(BeNil())
		})

		It("should process upsert event", func() {
			fakeStatusUpdates := []state.StatusUpdate{
				{
					NamespacedName: types.NamespacedName{},
					Status:         nil,
				},
			}
			// for now, we pass nil, because we don't need to test how EventLoop processes changes yet. We will start
			// testing once we have NGINX Configuration Manager component.
			fakeConf.UpsertHTTPRouteReturns(nil, fakeStatusUpdates)

			hr := &v1alpha2.HTTPRoute{}

			eventCh <- &events.UpsertEvent{
				Resource: hr,
			}

			Eventually(fakeConf.UpsertHTTPRouteCallCount).Should(Equal(1))
			Eventually(func() *v1alpha2.HTTPRoute {
				return fakeConf.UpsertHTTPRouteArgsForCall(0)
			}).Should(Equal(hr))

			Eventually(fakeUpdater.ProcessStatusUpdatesCallCount()).Should(Equal(1))
			Eventually(func() []state.StatusUpdate {
				_, updates := fakeUpdater.ProcessStatusUpdatesArgsForCall(0)
				return updates
			}).Should(Equal(fakeStatusUpdates))
		})

		It("should process delete event", func() {
			fakeStatusUpdates := []state.StatusUpdate{
				{
					NamespacedName: types.NamespacedName{},
					Status:         nil,
				},
			}
			// for now, we pass nil, because we don't need to test how EventLoop processes changes yet. We will start
			// testing once we have NGINX Configuration Manager component.
			fakeConf.DeleteHTTPRouteReturns(nil, fakeStatusUpdates)

			nsname := types.NamespacedName{Namespace: "test", Name: "route"}

			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &v1alpha2.HTTPRoute{},
			}

			Eventually(fakeConf.DeleteHTTPRouteCallCount).Should(Equal(1))
			Eventually(func() types.NamespacedName {
				return fakeConf.DeleteHTTPRouteArgsForCall(0)
			}).Should(Equal(nsname))

			Eventually(fakeUpdater.ProcessStatusUpdatesCallCount()).Should(Equal(1))
			Eventually(func() []state.StatusUpdate {
				_, updates := fakeUpdater.ProcessStatusUpdatesArgsForCall(0)
				return updates
			}).Should(Equal(fakeStatusUpdates))
		})
	})

	Describe("Edge cases", func() {
		AfterEach(func() {
			cancel()
		})

		DescribeTable("Edge cases for events",
			func(e interface{}) {
				eventCh <- e

				var err error
				Eventually(errorCh).Should(Receive(&err))
				Expect(err).Should(HaveOccurred())
			},
			Entry("should return error for an unknown event type",
				&struct{}{}),
			Entry("should return error for an unknown type of resource in upsert event",
				&events.UpsertEvent{
					Resource: &unsupportedResource{},
				}),
			Entry("should return error for an unknown type of resource in delete event",
				&events.DeleteEvent{
					Type: &unsupportedResource{},
				}),
		)

		It("should return error if status updater returns error", func() {
			testError := errors.New("test")
			fakeUpdater.ProcessStatusUpdatesReturns(testError)

			eventCh <- &events.UpsertEvent{
				Resource: &v1alpha2.HTTPRoute{},
			}

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).Should(Equal(testError))
		})
	})
})
