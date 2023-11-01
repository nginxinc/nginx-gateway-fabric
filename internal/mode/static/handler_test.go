package static

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctlrZap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status/statusfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics/collectors"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/configfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/file/filefakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/runtime/runtimefakes"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/statefakes"
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

		// One call for Gateway API resource statuses, one call for NginxProxy status
		Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(2))
	}

	BeforeEach(func() {
		fakeProcessor = &statefakes.FakeChangeProcessor{}
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxFileMgr = &filefakes.FakeManager{}
		fakeNginxRuntimeMgr = &runtimefakes.FakeManager{}
		fakeStatusUpdater = &statusfakes.FakeUpdater{}
		fakeEventRecorder = record.NewFakeRecorder(1)

		handler = newEventHandlerImpl(eventHandlerConfig{
			k8sClient:           fake.NewFakeClient(),
			processor:           fakeProcessor,
			generator:           fakeGenerator,
			logLevelSetter:      newZapLogLevelSetter(zap.NewAtomicLevel()),
			nginxFileMgr:        fakeNginxFileMgr,
			nginxRuntimeMgr:     fakeNginxRuntimeMgr,
			statusUpdater:       fakeStatusUpdater,
			eventRecorder:       fakeEventRecorder,
			healthChecker:       &healthChecker{},
			controlConfigNSName: types.NamespacedName{Namespace: namespace, Name: configName},
			gatewayPodConfig: config.GatewayPodConfig{
				ServiceName: "nginx-gateway",
				Namespace:   "nginx-gateway",
			},
			metricsCollector: collectors.NewControllerNoopCollector(),
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

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				checkUpsertEventExpectations(e)
				expectReconfig(dataplane.Configuration{Version: 1}, fakeCfgFiles)
			})

			It("should process Delete", func() {
				e := &events.DeleteEvent{
					Type:           &v1beta1.HTTPRoute{},
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"},
				}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

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

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				checkUpsertEventExpectations(upsertEvent)
				checkDeleteEventExpectations(deleteEvent)

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
			})
		})
	})

	When("receiving control plane configuration updates", func() {
		cfg := func(level ngfAPI.ControllerLogLevel) *ngfAPI.NginxGateway {
			return &ngfAPI.NginxGateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: namespace,
					Name:      configName,
				},
				Spec: ngfAPI.NginxGatewaySpec{
					Logging: &ngfAPI.Logging{
						Level: helpers.GetPointer(level),
					},
				},
			}
		}

		expStatuses := func(cond conditions.Condition) *status.NginxGatewayStatus {
			return &status.NginxGatewayStatus{
				NsName: types.NamespacedName{
					Namespace: namespace,
					Name:      configName,
				},
				Conditions:         []conditions.Condition{cond},
				ObservedGeneration: 0,
			}
		}

		It("handles a valid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(ngfAPI.ControllerLogLevelError)}}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
			_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
			Expect(statuses).To(Equal(expStatuses(staticConds.NewNginxGatewayValid())))
			Expect(handler.cfg.logLevelSetter.Enabled(zap.DebugLevel)).To(BeFalse())
			Expect(handler.cfg.logLevelSetter.Enabled(zap.ErrorLevel)).To(BeTrue())
		})

		It("handles an invalid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(ngfAPI.ControllerLogLevel("invalid"))}}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

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
					"\"invalid\": supported values: \"info\", \"debug\", \"error\"",
			))
			Expect(handler.cfg.logLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
		})

		It("handles a deleted config", func() {
			batch := []interface{}{&events.DeleteEvent{Type: &ngfAPI.NginxGateway{}}}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
			Expect(len(fakeEventRecorder.Events)).To(Equal(1))
			event := <-fakeEventRecorder.Events
			Expect(event).To(Equal("Warning ResourceDeleted NginxGateway configuration was deleted; using defaults"))
			Expect(handler.cfg.logLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
		})
	})

	When("receiving Service updates", func() {
		It("should not call UpdateAddresses if the Service is not for the Gateway Pod", func() {
			e := &events.UpsertEvent{Resource: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-nginx-gateway",
				},
			}}
			batch := []interface{}{e}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(fakeStatusUpdater.UpdateAddressesCallCount()).To(BeZero())

			de := &events.DeleteEvent{Type: &v1.Service{}}
			batch = []interface{}{de}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(fakeStatusUpdater.UpdateAddressesCallCount()).To(BeZero())
		})

		It("should update the addresses when the Gateway Service is upserted", func() {
			e := &events.UpsertEvent{Resource: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "nginx-gateway",
					Namespace: "nginx-gateway",
				},
			}}
			batch := []interface{}{e}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
			Expect(fakeStatusUpdater.UpdateAddressesCallCount()).ToNot(BeZero())
		})

		It("should update the addresses when the Gateway Service is deleted", func() {
			e := &events.DeleteEvent{
				Type: &v1.Service{},
				NamespacedName: types.NamespacedName{
					Name:      "nginx-gateway",
					Namespace: "nginx-gateway",
				},
			}
			batch := []interface{}{e}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
			Expect(fakeStatusUpdater.UpdateAddressesCallCount()).ToNot(BeZero())
		})
	})

	It("should set the health checker status properly when there are changes", func() {
		e := &events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}}
		batch := []interface{}{e}

		fakeProcessor.ProcessReturns(true, &graph.Graph{})

		Expect(handler.cfg.healthChecker.readyCheck(nil)).ToNot(Succeed())
		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
		Expect(handler.cfg.healthChecker.readyCheck(nil)).To(Succeed())
	})

	It("should set the health checker status properly when there are no changes or errors", func() {
		e := &events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}}
		batch := []interface{}{e}

		Expect(handler.cfg.healthChecker.readyCheck(nil)).ToNot(Succeed())
		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
		Expect(handler.cfg.healthChecker.readyCheck(nil)).To(Succeed())
	})

	It("should set the health checker status properly when there is an error", func() {
		e := &events.UpsertEvent{Resource: &v1beta1.HTTPRoute{}}
		batch := []interface{}{e}

		fakeProcessor.ProcessReturns(true, &graph.Graph{})
		fakeNginxRuntimeMgr.ReloadReturns(errors.New("reload error"))

		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(handler.cfg.healthChecker.readyCheck(nil)).ToNot(Succeed())

		// now send an update with no changes; should still return an error
		fakeProcessor.ProcessReturns(false, &graph.Graph{})

		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(handler.cfg.healthChecker.readyCheck(nil)).ToNot(Succeed())

		// error goes away
		fakeProcessor.ProcessReturns(true, &graph.Graph{})
		fakeNginxRuntimeMgr.ReloadReturns(nil)

		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(handler.cfg.healthChecker.readyCheck(nil)).To(Succeed())
	})

	It("should panic for an unknown event type", func() {
		e := &struct{}{}

		handle := func() {
			batch := []interface{}{e}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
		}

		Expect(handle).Should(Panic())
	})
})

var _ = Describe("getGatewayAddresses", func() {
	It("gets gateway addresses from a Service", func() {
		fakeClient := fake.NewFakeClient()
		podConfig := config.GatewayPodConfig{
			PodIP:       "1.2.3.4",
			ServiceName: "my-service",
			Namespace:   "nginx-gateway",
		}

		// no Service exists yet, should get error and Pod Address
		addrs, err := getGatewayAddresses(context.Background(), fakeClient, nil, podConfig)
		Expect(err).To(HaveOccurred())
		Expect(addrs).To(HaveLen(1))
		Expect(addrs[0].Value).To(Equal("1.2.3.4"))

		// Create LoadBalancer Service
		svc := v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "my-service",
				Namespace: "nginx-gateway",
			},
			Spec: v1.ServiceSpec{
				Type: v1.ServiceTypeLoadBalancer,
			},
			Status: v1.ServiceStatus{
				LoadBalancer: v1.LoadBalancerStatus{
					Ingress: []v1.LoadBalancerIngress{
						{
							IP: "34.35.36.37",
						},
						{
							Hostname: "myhost",
						},
					},
				},
			},
		}

		Expect(fakeClient.Create(context.Background(), &svc)).To(Succeed())

		addrs, err = getGatewayAddresses(context.Background(), fakeClient, &svc, podConfig)
		Expect(err).ToNot(HaveOccurred())
		Expect(addrs).To(HaveLen(2))
		Expect(addrs[0].Value).To(Equal("34.35.36.37"))
		Expect(addrs[1].Value).To(Equal("myhost"))
	})
})
