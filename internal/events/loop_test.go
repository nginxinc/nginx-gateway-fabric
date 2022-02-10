package events_test

import (
	"context"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/events"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state/statefakes"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
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
	var cancel context.CancelFunc
	var eventCh chan interface{}
	var errorCh chan error

	BeforeEach(func() {
		fakeConf = &statefakes.FakeConfiguration{}
		eventCh = make(chan interface{})
		ctrl = events.NewEventLoop(fakeConf, eventCh)

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
			hr := &v1alpha2.HTTPRoute{}

			eventCh <- &events.UpsertEvent{
				Resource: hr,
			}

			Eventually(fakeConf.UpsertHTTPRouteCallCount()).Should(Equal(1))
			Eventually(fakeConf.UpsertHTTPRouteArgsForCall(0)).Should(Equal(hr))
		})

		It("should process delete event", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "route"}

			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &v1alpha2.HTTPRoute{},
			}

			Eventually(fakeConf.DeleteHTTPRouteCallCount()).Should(Equal(1))
			Eventually(fakeConf.DeleteHTTPRouteArgsForCall(0)).Should(Equal(nsname))
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
