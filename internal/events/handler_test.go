package events_test

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
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

var _ = Describe("EventHandler", func() {
	var (
		handler                 *events.EventHandlerImpl
		fakeProcessor           *statefakes.FakeChangeProcessor
		fakeSecretStore         *statefakes.FakeSecretStore
		fakeSecretMemoryManager *statefakes.FakeSecretDiskMemoryManager
		fakeGenerator           *configfakes.FakeGenerator
		fakeNginxFileMgr        *filefakes.FakeManager
		fakeNginxRuntimeMgr     *runtimefakes.FakeManager
		fakeStatusUpdater       *statusfakes.FakeUpdater
	)

	expectReconfig := func(expectedConf state.Configuration, expectedCfg []byte, expectedStatuses state.Statuses) {
		Expect(fakeProcessor.ProcessCallCount()).Should(Equal(1))

		Expect(fakeGenerator.GenerateCallCount()).Should(Equal(1))
		Expect(fakeGenerator.GenerateArgsForCall(0)).Should(Equal(expectedConf))

		Expect(fakeNginxFileMgr.WriteHTTPConfigCallCount()).Should(Equal(1))
		name, cfg := fakeNginxFileMgr.WriteHTTPConfigArgsForCall(0)
		Expect(name).Should(Equal("http"))
		Expect(cfg).Should(Equal(expectedCfg))

		Expect(fakeNginxRuntimeMgr.ReloadCallCount()).Should(Equal(1))

		Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
		_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
		Expect(statuses).Should(Equal(expectedStatuses))
	}

	BeforeEach(func() {
		fakeProcessor = &statefakes.FakeChangeProcessor{}
		fakeSecretMemoryManager = &statefakes.FakeSecretDiskMemoryManager{}
		fakeSecretStore = &statefakes.FakeSecretStore{}
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxFileMgr = &filefakes.FakeManager{}
		fakeNginxRuntimeMgr = &runtimefakes.FakeManager{}
		fakeStatusUpdater = &statusfakes.FakeUpdater{}

		handler = events.NewEventHandlerImpl(events.EventHandlerConfig{
			Processor:           fakeProcessor,
			SecretStore:         fakeSecretStore,
			SecretMemoryManager: fakeSecretMemoryManager,
			Generator:           fakeGenerator,
			Logger:              zap.New(),
			NginxFileMgr:        fakeNginxFileMgr,
			NginxRuntimeMgr:     fakeNginxRuntimeMgr,
			StatusUpdater:       fakeStatusUpdater,
		})
	})

	Describe("Process the Gateway API resources events", func() {
		DescribeTable("A batch with one event",
			func(e interface{}) {
				fakeConf := state.Configuration{}
				fakeStatuses := state.Statuses{}
				changed := true
				fakeProcessor.ProcessReturns(changed, fakeConf, fakeStatuses)

				fakeCfg := []byte("fake")
				fakeGenerator.GenerateReturns(fakeCfg)

				batch := []interface{}{e}

				handler.HandleEventBatch(context.TODO(), batch)

				// Check that the events were captured
				switch typedEvent := e.(type) {
				case *events.UpsertEvent:
					Expect(fakeProcessor.CaptureUpsertChangeCallCount()).Should(Equal(1))
					Expect(fakeProcessor.CaptureUpsertChangeArgsForCall(0)).Should(Equal(typedEvent.Resource))
				case *events.DeleteEvent:
					Expect(fakeProcessor.CaptureDeleteChangeCallCount()).Should(Equal(1))
					passedObj, passedNsName := fakeProcessor.CaptureDeleteChangeArgsForCall(0)
					Expect(passedObj).Should(Equal(typedEvent.Type))
					Expect(passedNsName).Should(Equal(typedEvent.NamespacedName))
				default:
					Fail(fmt.Sprintf("unsupported event type %T", e))
				}

				// Check that a reconfig happened
				expectReconfig(fakeConf, fakeCfg, fakeStatuses)
			},
			Entry("HTTPRoute upsert", &events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}}),
			Entry("Gateway upsert", &events.UpsertEvent{Resource: &v1beta1.Gateway{}}),
			Entry("GatewayClass upsert", &events.UpsertEvent{Resource: &v1beta1.GatewayClass{}}),
			Entry("Service upsert", &events.UpsertEvent{Resource: &apiv1.Service{}}),
			Entry("EndpointSlice upsert", &events.UpsertEvent{Resource: &discoveryV1.EndpointSlice{}}),

			Entry("HTTPRoute delete", &events.DeleteEvent{Type: &v1beta1.HTTPRoute{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"}}),
			Entry("Gateway delete", &events.DeleteEvent{Type: &v1beta1.Gateway{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "gateway"}}),
			Entry("GatewayClass delete", &events.DeleteEvent{Type: &v1beta1.GatewayClass{}, NamespacedName: types.NamespacedName{Name: "class"}}),
			Entry("Service delete", &events.DeleteEvent{Type: &apiv1.Service{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "service"}}),
			Entry("EndpointSlice deleted", &events.DeleteEvent{Type: &discoveryV1.EndpointSlice{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "endpointslice"}}),
		)
	})

	Describe("Process Secret events", func() {
		expectNoReconfig := func() {
			Expect(fakeProcessor.ProcessCallCount()).Should(Equal(1))
			Expect(fakeGenerator.GenerateCallCount()).Should(Equal(0))
			Expect(fakeNginxFileMgr.WriteHTTPConfigCallCount()).Should(Equal(0))
			Expect(fakeNginxRuntimeMgr.ReloadCallCount()).Should(Equal(0))
			Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(0))
		}
		It("should process upsert event", func() {
			secret := &apiv1.Secret{}

			batch := []interface{}{
				&events.UpsertEvent{
					Resource: secret,
				},
			}

			handler.HandleEventBatch(context.TODO(), batch)

			Expect(fakeSecretStore.UpsertCallCount()).Should(Equal(1))
			Expect(fakeSecretStore.UpsertArgsForCall(0)).Should(Equal(secret))

			expectNoReconfig()
		})

		It("should process delete event", func() {
			nsname := types.NamespacedName{Namespace: "test", Name: "secret"}

			batch := []interface{}{
				&events.DeleteEvent{
					NamespacedName: nsname,
					Type:           &apiv1.Secret{},
				},
			}

			handler.HandleEventBatch(context.TODO(), batch)

			Expect(fakeSecretStore.DeleteCallCount()).Should(Equal(1))
			Expect(fakeSecretStore.DeleteArgsForCall(0)).Should(Equal(nsname))

			expectNoReconfig()
		})
	})

	It("should process a batch with upsert and delete events for every supported resource", func() {
		svc := &apiv1.Service{}
		svcNsName := types.NamespacedName{Namespace: "test", Name: "service"}
		secret := &apiv1.Secret{}
		secretNsName := types.NamespacedName{Namespace: "test", Name: "secret"}

		upserts := []interface{}{
			&events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}},
			&events.UpsertEvent{Resource: &v1beta1.Gateway{}},
			&events.UpsertEvent{Resource: &v1beta1.GatewayClass{}},
			&events.UpsertEvent{Resource: svc},
			&events.UpsertEvent{Resource: &discoveryV1.EndpointSlice{}},
			&events.UpsertEvent{Resource: secret},
		}
		deletes := []interface{}{
			&events.DeleteEvent{Type: &v1beta1.HTTPRoute{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"}},
			&events.DeleteEvent{Type: &v1beta1.Gateway{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "gateway"}},
			&events.DeleteEvent{Type: &v1beta1.GatewayClass{}, NamespacedName: types.NamespacedName{Name: "class"}},
			&events.DeleteEvent{Type: &apiv1.Service{}, NamespacedName: svcNsName},
			&events.DeleteEvent{Type: &discoveryV1.EndpointSlice{}, NamespacedName: types.NamespacedName{Namespace: "test", Name: "endpointslice"}},
			&events.DeleteEvent{Type: &apiv1.Secret{}, NamespacedName: secretNsName},
		}

		batch := make([]interface{}, 0, len(upserts)+len(deletes))
		batch = append(batch, upserts...)
		batch = append(batch, deletes...)

		fakeConf := state.Configuration{}
		changed := true
		fakeStatuses := state.Statuses{}
		fakeProcessor.ProcessReturns(changed, fakeConf, fakeStatuses)

		fakeCfg := []byte("fake")
		fakeGenerator.GenerateReturns(fakeCfg)

		handler.HandleEventBatch(context.TODO(), batch)

		// Check that the events for Gateway API resources were captured

		// 5, not 6, because secret upsert events do not result into CaptureUpsertChange() call
		Expect(fakeProcessor.CaptureUpsertChangeCallCount()).Should(Equal(5))
		for i := 0; i < 5; i++ {
			Expect(fakeProcessor.CaptureUpsertChangeArgsForCall(i)).Should(Equal(upserts[i].(*events.UpsertEvent).Resource))
		}

		// 5, not 6, because secret delete events do not result into CaptureDeleteChange() call
		Expect(fakeProcessor.CaptureDeleteChangeCallCount()).Should(Equal(5))
		for i := 0; i < 5; i++ {
			d := deletes[i].(*events.DeleteEvent)
			passedObj, passedNsName := fakeProcessor.CaptureDeleteChangeArgsForCall(i)
			Expect(passedObj).Should(Equal(d.Type))
			Expect(passedNsName).Should(Equal(d.NamespacedName))
		}

		// Check Secret-related expectations
		Expect(fakeSecretStore.UpsertCallCount()).Should(Equal(1))
		Expect(fakeSecretStore.UpsertArgsForCall(0)).Should(Equal(secret))

		Expect(fakeSecretStore.DeleteCallCount()).Should(Equal(1))
		Expect(fakeSecretStore.DeleteArgsForCall(0)).Should(Equal(secretNsName))

		// Check that a reconfig happened
		expectReconfig(fakeConf, fakeCfg, fakeStatuses)
	})

	Describe("Edge cases", func() {
		DescribeTable("Edge cases for events",
			func(e interface{}) {
				handle := func() {
					batch := []interface{}{e}
					handler.HandleEventBatch(context.TODO(), batch)
				}

				Expect(handle).Should(Panic())
			},
			Entry("should panic for an unknown event type",
				&struct{}{}),
			Entry("should panic for an unknown type of resource in upsert event",
				&events.UpsertEvent{
					Resource: &unsupportedResource{},
				}),
			Entry("should panic for an unknown type of resource in delete event",
				&events.DeleteEvent{
					Type: &unsupportedResource{},
				}),
		)
	})
})
