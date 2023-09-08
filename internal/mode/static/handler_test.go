package static

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	ctlrZap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	nkgAPI "github.com/nginxinc/nginx-kubernetes-gateway/apis/v1alpha1"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status/statusfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/config/configfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/file/filefakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/nginx/runtime/runtimefakes"
	staticConds "github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/mode/static/state/statefakes"
)

var _ = Describe("eventHandler", func() {
	var (
		handler             *eventHandlerImpl
		fakeProcessor       *statefakes.FakeChangeProcessor
		fakeGenerator       *configfakes.FakeGenerator
		fakeNginxFileMgr    *filefakes.FakeManager
		fakeNginxRuntimeMgr *runtimefakes.FakeManager
		fakeStatusUpdater   *statusfakes.FakeUpdater
		fakeEventRecorder   *record.FakeRecorder
		namespace           = "nginx-gateway"
		configName          = "nginx-gateway-config"
	)

	expectReconfig := func(expectedConf dataplane.Configuration, expectedFiles []file.File) {
		Expect(fakeProcessor.ProcessCallCount()).Should(Equal(1))

		Expect(fakeGenerator.GenerateCallCount()).Should(Equal(1))
		Expect(fakeGenerator.GenerateArgsForCall(0)).Should(Equal(expectedConf))

		Expect(fakeNginxFileMgr.ReplaceFilesCallCount()).Should(Equal(1))
		files := fakeNginxFileMgr.ReplaceFilesArgsForCall(0)
		Expect(files).Should(Equal(expectedFiles))

		Expect(fakeNginxRuntimeMgr.ReloadCallCount()).Should(Equal(1))

		Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
	}

	BeforeEach(func() {
		fakeProcessor = &statefakes.FakeChangeProcessor{}
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxFileMgr = &filefakes.FakeManager{}
		fakeNginxRuntimeMgr = &runtimefakes.FakeManager{}
		fakeStatusUpdater = &statusfakes.FakeUpdater{}
		fakeEventRecorder = record.NewFakeRecorder(1)

		handler = newEventHandlerImpl(eventHandlerConfig{
			processor:           fakeProcessor,
			generator:           fakeGenerator,
			logger:              ctlrZap.New(),
			logLevelSetter:      newZapLogLevelSetter(zap.NewAtomicLevel()),
			nginxFileMgr:        fakeNginxFileMgr,
			nginxRuntimeMgr:     fakeNginxRuntimeMgr,
			statusUpdater:       fakeStatusUpdater,
			eventRecorder:       fakeEventRecorder,
			healthChecker:       &healthChecker{},
			controlConfigNSName: types.NamespacedName{Namespace: namespace, Name: configName},
		})
		Expect(handler.cfg.healthChecker.ready).To(BeFalse())
	})

	Describe("Process the Gateway API resources events", func() {
		fakeCfgFiles := []file.File{
			{
				Type: file.TypeRegular,
				Path: "test.conf",
			},
		}

		checkUpsertEventExpectations := func(e *events.UpsertEvent) {
			Expect(fakeProcessor.CaptureUpsertChangeCallCount()).Should(Equal(1))
			Expect(fakeProcessor.CaptureUpsertChangeArgsForCall(0)).Should(Equal(e.Resource))
		}

		checkDeleteEventExpectations := func(e *events.DeleteEvent) {
			Expect(fakeProcessor.CaptureDeleteChangeCallCount()).Should(Equal(1))
			passedResourceType, passedNsName := fakeProcessor.CaptureDeleteChangeArgsForCall(0)
			Expect(passedResourceType).Should(Equal(e.Type))
			Expect(passedNsName).Should(Equal(e.NamespacedName))
		}

		BeforeEach(func() {
			fakeProcessor.ProcessReturns(true /* changed */, &graph.Graph{})

			fakeGenerator.GenerateReturns(fakeCfgFiles)
		})

		AfterEach(func() {
			Expect(handler.cfg.healthChecker.ready).To(BeTrue())
		})

		When("a batch has one event", func() {
			It("should process Upsert", func() {
				e := &events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), batch)

				checkUpsertEventExpectations(e)
				expectReconfig(dataplane.Configuration{Version: 1}, fakeCfgFiles)
			})

			It("should process Delete", func() {
				e := &events.DeleteEvent{
					Type:           &v1beta1.HTTPRoute{},
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"},
				}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), batch)

				checkDeleteEventExpectations(e)
				expectReconfig(dataplane.Configuration{Version: 1}, fakeCfgFiles)
			})
		})

		When("a batch has multiple events", func() {
			It("should process events", func() {
				upsertEvent := &events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}}
				deleteEvent := &events.DeleteEvent{
					Type:           &v1beta1.HTTPRoute{},
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"},
				}
				batch := []interface{}{upsertEvent, deleteEvent}

				handler.HandleEventBatch(context.Background(), batch)

				checkUpsertEventExpectations(upsertEvent)
				checkDeleteEventExpectations(deleteEvent)

				handler.HandleEventBatch(context.Background(), batch)
			})
		})
	})

	When("receiving control plane configuration updates", func() {
		cfg := func(level nkgAPI.ControllerLogLevel) *nkgAPI.NginxGateway {
			return &nkgAPI.NginxGateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      configName,
				},
				Spec: nkgAPI.NginxGatewaySpec{
					Logging: &nkgAPI.Logging{
						Level: helpers.GetPointer(level),
					},
				},
			}
		}

		expStatuses := func(cond conditions.Condition) status.Statuses {
			return status.Statuses{
				NginxGatewayStatus: &status.NginxGatewayStatus{
					NsName: types.NamespacedName{
						Namespace: namespace,
						Name:      configName,
					},
					Conditions:         []conditions.Condition{cond},
					ObservedGeneration: 0,
				},
			}
		}

		It("handles a valid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(nkgAPI.ControllerLogLevelError)}}
			handler.HandleEventBatch(context.Background(), batch)

			Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
			_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
			Expect(statuses).To(Equal(expStatuses(staticConds.NewNginxGatewayValid())))
			Expect(handler.cfg.logLevelSetter.Enabled(zap.DebugLevel)).To(BeFalse())
			Expect(handler.cfg.logLevelSetter.Enabled(zap.ErrorLevel)).To(BeTrue())
		})

		It("handles an invalid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(nkgAPI.ControllerLogLevel("invalid"))}}
			handler.HandleEventBatch(context.Background(), batch)

			Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
			_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
			cond := staticConds.NewNginxGatewayInvalid(
				"Failed to update control plane configuration: logging.level: Unsupported value: " +
					"\"invalid\": supported values: \"info\", \"debug\", \"error\"")
			Expect(statuses).To(Equal(expStatuses(cond)))
			Expect(len(fakeEventRecorder.Events)).To(Equal(1))
			event := <-fakeEventRecorder.Events
			Expect(event).To(Equal(
				"Warning UpdateFailed Failed to update control plane configuration: logging.level: Unsupported value: " +
					"\"invalid\": supported values: \"info\", \"debug\", \"error\""))
			Expect(handler.cfg.logLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
		})

		It("handles a deleted config", func() {
			batch := []interface{}{&events.DeleteEvent{Type: &nkgAPI.NginxGateway{}}}
			handler.HandleEventBatch(context.Background(), batch)
			Expect(len(fakeEventRecorder.Events)).To(Equal(1))
			event := <-fakeEventRecorder.Events
			Expect(event).To(Equal("Warning ResourceDeleted NginxGateway configuration was deleted; using defaults"))
			Expect(handler.cfg.logLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
		})
	})

	It("should panic for an unknown event type", func() {
		e := &struct{}{}

		handle := func() {
			batch := []interface{}{e}
			handler.HandleEventBatch(context.TODO(), batch)
		}

		Expect(handle).Should(Panic())
	})
})
