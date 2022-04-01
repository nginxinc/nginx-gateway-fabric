package events_test

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/config/configfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/file/filefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/nginx/runtime/runtimefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/statefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status/statusfakes"
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
	var fakeGenerator *configfakes.FakeGenerator
	var fakeNginxFimeMgr *filefakes.FakeManager
	var fakeNginxRuntimeMgr *runtimefakes.FakeManager
	var cancel context.CancelFunc
	var eventCh chan interface{}
	var errorCh chan error

	BeforeEach(func() {
		fakeConf = &statefakes.FakeConfiguration{}
		eventCh = make(chan interface{})
		fakeUpdater = &statusfakes.FakeUpdater{}
		fakeServiceStore = &statefakes.FakeServiceStore{}
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxFimeMgr = &filefakes.FakeManager{}
		fakeNginxRuntimeMgr = &runtimefakes.FakeManager{}
		ctrl = events.NewEventLoop(fakeConf, fakeServiceStore, fakeGenerator, eventCh, fakeUpdater, zap.New(), fakeNginxFimeMgr, fakeNginxRuntimeMgr)

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
			fakeChanges := []state.Change{
				{
					Op: state.Upsert,
					Host: state.Host{
						Value: "example.com",
					},
				},
			}
			fakeStatusUpdates := []state.StatusUpdate{
				{
					NamespacedName: types.NamespacedName{},
					Status:         nil,
				},
			}
			fakeConf.UpsertHTTPRouteReturns(fakeChanges, fakeStatusUpdates)

			fakeCfg := []byte("fake")
			fakeGenerator.GenerateForHostReturns(fakeCfg, config.Warnings{})

			hr := &v1alpha2.HTTPRoute{}

			eventCh <- &events.UpsertEvent{
				Resource: hr,
			}

			Eventually(fakeConf.UpsertHTTPRouteCallCount).Should(Equal(1))
			Eventually(func() *v1alpha2.HTTPRoute {
				return fakeConf.UpsertHTTPRouteArgsForCall(0)
			}).Should(Equal(hr))

			Eventually(fakeUpdater.ProcessStatusUpdatesCallCount).Should(Equal(1))
			Eventually(func() []state.StatusUpdate {
				_, updates := fakeUpdater.ProcessStatusUpdatesArgsForCall(0)
				return updates
			}).Should(Equal(fakeStatusUpdates))

			Eventually(fakeGenerator.GenerateForHostCallCount).Should(Equal(1))
			Expect(fakeGenerator.GenerateForHostArgsForCall(0)).Should(Equal(fakeChanges[0].Host))

			Eventually(fakeNginxFimeMgr.WriteServerConfigCallCount).Should(Equal(1))
			host, cfg := fakeNginxFimeMgr.WriteServerConfigArgsForCall(0)
			Expect(host).Should(Equal("example.com"))
			Expect(cfg).Should(Equal(fakeCfg))

			Eventually(fakeNginxRuntimeMgr.ReloadCallCount).Should(Equal(1))
		})

		It("should process delete event", func() {
			fakeChanges := []state.Change{
				{
					Op: state.Delete,
					Host: state.Host{
						Value: "example.com",
					},
				},
			}
			fakeStatusUpdates := []state.StatusUpdate{
				{
					NamespacedName: types.NamespacedName{},
					Status:         nil,
				},
			}
			fakeConf.DeleteHTTPRouteReturns(fakeChanges, fakeStatusUpdates)

			nsname := types.NamespacedName{Namespace: "test", Name: "route"}

			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &v1alpha2.HTTPRoute{},
			}

			Eventually(fakeConf.DeleteHTTPRouteCallCount).Should(Equal(1))
			Eventually(func() types.NamespacedName {
				return fakeConf.DeleteHTTPRouteArgsForCall(0)
			}).Should(Equal(nsname))

			Eventually(fakeUpdater.ProcessStatusUpdatesCallCount).Should(Equal(1))
			Eventually(func() []state.StatusUpdate {
				_, updates := fakeUpdater.ProcessStatusUpdatesArgsForCall(0)
				return updates
			}).Should(Equal(fakeStatusUpdates))

			Eventually(fakeNginxFimeMgr.DeleteServerConfigCallCount).Should(Equal(1))
			Expect(fakeNginxFimeMgr.DeleteServerConfigArgsForCall(0)).Should(Equal("example.com"))

			Eventually(fakeNginxRuntimeMgr.ReloadCallCount).Should(Equal(1))
		})
	})

	Describe("Process Service events", func() {
		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).To(BeNil())
		})

		It("should process upsert event", func() {
			svc := &apiv1.Service{}

			eventCh <- &events.UpsertEvent{
				Resource: svc,
			}

			Eventually(fakeServiceStore.UpsertCallCount).Should(Equal(1))
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

			Eventually(fakeServiceStore.DeleteCallCount).Should(Equal(1))
			Eventually(func() types.NamespacedName {
				return fakeServiceStore.DeleteArgsForCall(0)
			}).Should(Equal(nsname))
		})
	})

	Describe("Processing events common cases", func() {
		AfterEach(func() {
			cancel()

			var err error
			Eventually(errorCh).Should(Receive(&err))
			Expect(err).To(BeNil())
		})

		It("should reload once in case of multiple changes", func() {
			fakeChanges := []state.Change{
				{
					Op: state.Delete,
					Host: state.Host{
						Value: "one.example.com",
					},
				},
				{
					Op: state.Upsert,
					Host: state.Host{
						Value: "two.example.com",
					},
				},
			}
			fakeConf.DeleteHTTPRouteReturns(fakeChanges, nil)

			fakeCfg := []byte("fake")
			fakeGenerator.GenerateForHostReturns(fakeCfg, config.Warnings{})

			nsname := types.NamespacedName{Namespace: "test", Name: "route"}

			// the exact event doesn't matter. what matters is the generated changes return by DeleteHTTPRouteReturns
			eventCh <- &events.DeleteEvent{
				NamespacedName: nsname,
				Type:           &v1alpha2.HTTPRoute{},
			}

			Eventually(fakeConf.DeleteHTTPRouteCallCount).Should(Equal(1))
			Expect(fakeConf.DeleteHTTPRouteArgsForCall(0)).Should(Equal(nsname))

			Eventually(fakeNginxFimeMgr.WriteServerConfigCallCount).Should(Equal(1))
			host, cfg := fakeNginxFimeMgr.WriteServerConfigArgsForCall(0)
			Expect(host).Should(Equal("two.example.com"))
			Expect(cfg).Should(Equal(fakeCfg))

			Eventually(fakeNginxFimeMgr.DeleteServerConfigCallCount).Should(Equal(1))
			Expect(fakeNginxFimeMgr.DeleteServerConfigArgsForCall(0)).Should(Equal("one.example.com"))

			Eventually(fakeNginxRuntimeMgr.ReloadCallCount).Should(Equal(1))
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
