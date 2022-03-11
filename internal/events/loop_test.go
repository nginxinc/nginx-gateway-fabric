package events_test

import (
	"context"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/events"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state/statefakes"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/status/statusfakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
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
	var fakeServiceStore *statefakes.FakeServiceStore
	var cancel context.CancelFunc
	var eventCh chan interface{}
	var errorCh chan error

	BeforeEach(func() {
		fakeConf = &statefakes.FakeConfiguration{}
		eventCh = make(chan interface{})
		fakeUpdater = &statusfakes.FakeUpdater{}
		fakeServiceStore = &statefakes.FakeServiceStore{}
		ctrl = events.NewEventLoop(fakeConf, fakeServiceStore, eventCh, fakeUpdater, zap.New())

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

	Describe("Process Service events", func() {
		It("should process upsert event", func() {
			svc := &apiv1.Service{}

			eventCh <- &events.UpsertEvent{
				Resource: svc,
			}

			Eventually(fakeServiceStore.UpsertCallCount()).Should(Equal(1))
			Eventually(func() *apiv1.Service {
				return fakeServiceStore.UpsertArgsForCall(0)
			}).Should(Equal(svc))
		})

		It("should process delete event", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "service"}

			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &apiv1.Service{},
			}

			Eventually(fakeServiceStore.DeleteCallCount()).Should(Equal(1))
			Eventually(func() types.NamespacedName {
				return fakeServiceStore.DeleteArgsForCall(0)
			}).Should(Equal(nsname))
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
	})
})
