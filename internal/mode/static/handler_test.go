package static

import (
	"context"
	"errors"

	pb "github.com/nginx/agent/v3/api/grpc/mpi/v1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	ctlrZap "sigs.k8s.io/controller-runtime/pkg/log/zap"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/events"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status/statusfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/config"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/licensing/licensingfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/metrics/collectors"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/agent"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/agent/agentfakes"
	agentgrpcfakes "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/agent/grpc/grpcfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/configfakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/statefakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/status"
)

var _ = Describe("eventHandler", func() {
	var (
		handler           *eventHandlerImpl
		fakeProcessor     *statefakes.FakeChangeProcessor
		fakeGenerator     *configfakes.FakeGenerator
		fakeNginxUpdater  *agentfakes.FakeNginxUpdater
		fakeStatusUpdater *statusfakes.FakeGroupUpdater
		fakeEventRecorder *record.FakeRecorder
		fakeK8sClient     client.WithWatch
		queue             *status.Queue
		namespace         = "nginx-gateway"
		configName        = "nginx-gateway-config"
		zapLogLevelSetter zapLogLevelSetter
		ctx               context.Context
		cancel            context.CancelFunc
	)

	const nginxGatewayServiceName = "nginx-gateway"

	createService := func(name string) *v1.Service {
		return &v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: "nginx-gateway",
			},
		}
	}

	expectReconfig := func(expectedConf dataplane.Configuration, expectedFiles []agent.File) {
		Expect(fakeProcessor.ProcessCallCount()).Should(Equal(1))

		Expect(fakeGenerator.GenerateCallCount()).Should(Equal(1))
		Expect(fakeGenerator.GenerateArgsForCall(0)).Should(Equal(expectedConf))

		Expect(fakeNginxUpdater.UpdateConfigCallCount()).Should(Equal(1))
		_, files := fakeNginxUpdater.UpdateConfigArgsForCall(0)
		Expect(expectedFiles).To(Equal(files))

		Eventually(
			func() int {
				return fakeStatusUpdater.UpdateGroupCallCount()
			}).Should(Equal(2))
		_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
		Expect(name).To(Equal(groupAllExceptGateways))
		Expect(reqs).To(BeEmpty())

		_, name, reqs = fakeStatusUpdater.UpdateGroupArgsForCall(1)
		Expect(name).To(Equal(groupGateways))
		Expect(reqs).To(BeEmpty())
	}

	BeforeEach(func() {
		ctx, cancel = context.WithCancel(context.Background()) //nolint:fatcontext // ignore for test

		fakeProcessor = &statefakes.FakeChangeProcessor{}
		fakeProcessor.ProcessReturns(state.NoChange, &graph.Graph{})
		fakeProcessor.GetLatestGraphReturns(&graph.Graph{})
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxUpdater = &agentfakes.FakeNginxUpdater{}
		fakeNginxUpdater.UpdateConfigReturns(true)
		fakeStatusUpdater = &statusfakes.FakeGroupUpdater{}
		fakeEventRecorder = record.NewFakeRecorder(1)
		zapLogLevelSetter = newZapLogLevelSetter(zap.NewAtomicLevel())
		fakeK8sClient = fake.NewFakeClient()
		queue = status.NewQueue()

		// Needed because handler checks the service from the API on every HandleEventBatch
		Expect(fakeK8sClient.Create(context.Background(), createService(nginxGatewayServiceName))).To(Succeed())

		handler = newEventHandlerImpl(eventHandlerConfig{
			ctx:                     ctx,
			k8sClient:               fakeK8sClient,
			processor:               fakeProcessor,
			generator:               fakeGenerator,
			logLevelSetter:          zapLogLevelSetter,
			nginxUpdater:            fakeNginxUpdater,
			statusUpdater:           fakeStatusUpdater,
			eventRecorder:           fakeEventRecorder,
			deployCtxCollector:      &licensingfakes.FakeCollector{},
			graphBuiltHealthChecker: newGraphBuiltHealthChecker(),
			statusQueue:             queue,
			nginxDeployments:        agent.NewDeploymentStore(&agentgrpcfakes.FakeConnectionsTracker{}),
			controlConfigNSName:     types.NamespacedName{Namespace: namespace, Name: configName},
			gatewayPodConfig: config.GatewayPodConfig{
				ServiceName: "nginx-gateway",
				Namespace:   "nginx-gateway",
			},
			metricsCollector:         collectors.NewControllerNoopCollector(),
			updateGatewayClassStatus: true,
		})
		Expect(handler.cfg.graphBuiltHealthChecker.ready).To(BeFalse())
	})

	AfterEach(func() {
		cancel()
	})

	Describe("Process the Gateway API resources events", func() {
		fakeCfgFiles := []agent.File{
			{
				Meta: &pb.FileMeta{
					Name: "test.conf",
				},
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
			fakeProcessor.ProcessReturns(state.ClusterStateChange /* changed */, &graph.Graph{})

			fakeGenerator.GenerateReturns(fakeCfgFiles)
		})

		AfterEach(func() {
			Expect(handler.cfg.graphBuiltHealthChecker.ready).To(BeTrue())
		})

		When("a batch has one event", func() {
			It("should process Upsert", func() {
				e := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := dataplane.GetDefaultConfiguration(&graph.Graph{}, 1)

				checkUpsertEventExpectations(e)
				expectReconfig(dcfg, fakeCfgFiles)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dcfg)).To(BeEmpty())
			})

			It("should process Delete", func() {
				e := &events.DeleteEvent{
					Type:           &gatewayv1.HTTPRoute{},
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"},
				}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := dataplane.GetDefaultConfiguration(&graph.Graph{}, 1)

				checkDeleteEventExpectations(e)
				expectReconfig(dcfg, fakeCfgFiles)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dcfg)).To(BeEmpty())
			})
		})

		When("a batch has multiple events", func() {
			It("should process events", func() {
				upsertEvent := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
				deleteEvent := &events.DeleteEvent{
					Type:           &gatewayv1.HTTPRoute{},
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"},
				}
				batch := []interface{}{upsertEvent, deleteEvent}

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				checkUpsertEventExpectations(upsertEvent)
				checkDeleteEventExpectations(deleteEvent)

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := dataplane.GetDefaultConfiguration(&graph.Graph{}, 2)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dcfg)).To(BeEmpty())
			})
		})
	})

	DescribeTable(
		"updating statuses of GatewayClass conditionally based on handler configuration",
		func(updateGatewayClassStatus bool) {
			handler.cfg.updateGatewayClassStatus = updateGatewayClassStatus

			gc := &gatewayv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
			}
			ignoredGC := &gatewayv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "ignored",
				},
			}

			gr := &graph.Graph{
				GatewayClass: &graph.GatewayClass{
					Source: gc,
					Valid:  true,
				},
				IgnoredGatewayClasses: map[types.NamespacedName]*gatewayv1.GatewayClass{
					client.ObjectKeyFromObject(ignoredGC): ignoredGC,
				},
			}

			fakeProcessor.ProcessReturns(state.ClusterStateChange, gr)
			fakeProcessor.GetLatestGraphReturns(gr)

			e := &events.UpsertEvent{
				Resource: &gatewayv1.HTTPRoute{}, // any supported is OK
			}

			batch := []interface{}{e}

			var expectedReqsCount int
			if updateGatewayClassStatus {
				expectedReqsCount = 2
			}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Eventually(
				func() int {
					return fakeStatusUpdater.UpdateGroupCallCount()
				}).Should(Equal(2))

			_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
			Expect(name).To(Equal(groupAllExceptGateways))
			Expect(reqs).To(HaveLen(expectedReqsCount))
			for _, req := range reqs {
				Expect(req.NsName).To(BeElementOf(
					client.ObjectKeyFromObject(gc),
					client.ObjectKeyFromObject(ignoredGC),
				))
				Expect(req.ResourceType).To(Equal(&gatewayv1.GatewayClass{}))
			}
		},
		Entry("should update statuses of GatewayClass", true),
		Entry("should not update statuses of GatewayClass", false),
	)

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

		It("handles a valid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(ngfAPI.ControllerLogLevelError)}}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(handler.GetLatestConfiguration()).To(BeNil())

			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(Equal(1))
			_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
			Expect(name).To(Equal(groupControlPlane))
			Expect(reqs).To(HaveLen(1))

			Expect(zapLogLevelSetter.Enabled(zap.DebugLevel)).To(BeFalse())
			Expect(zapLogLevelSetter.Enabled(zap.ErrorLevel)).To(BeTrue())
		})

		It("handles an invalid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(ngfAPI.ControllerLogLevel("invalid"))}}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(handler.GetLatestConfiguration()).To(BeNil())

			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(Equal(1))
			_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
			Expect(name).To(Equal(groupControlPlane))
			Expect(reqs).To(HaveLen(1))

			Expect(fakeEventRecorder.Events).To(HaveLen(1))
			event := <-fakeEventRecorder.Events
			Expect(event).To(Equal(
				"Warning UpdateFailed Failed to update control plane configuration: logging.level: Unsupported value: " +
					"\"invalid\": supported values: \"info\", \"debug\", \"error\"",
			))
			Expect(zapLogLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
		})

		It("handles a deleted config", func() {
			batch := []interface{}{
				&events.DeleteEvent{
					Type: &ngfAPI.NginxGateway{},
					NamespacedName: types.NamespacedName{
						Namespace: namespace,
						Name:      configName,
					},
				},
			}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(handler.GetLatestConfiguration()).To(BeNil())

			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(Equal(1))
			_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
			Expect(name).To(Equal(groupControlPlane))
			Expect(reqs).To(BeEmpty())

			Expect(fakeEventRecorder.Events).To(HaveLen(1))
			event := <-fakeEventRecorder.Events
			Expect(event).To(Equal("Warning ResourceDeleted NginxGateway configuration was deleted; using defaults"))
			Expect(zapLogLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
		})
	})

	When("receiving Service updates", func() {
		const notNginxGatewayServiceName = "not-nginx-gateway"

		BeforeEach(func() {
			fakeProcessor.GetLatestGraphReturns(&graph.Graph{})

			Expect(fakeK8sClient.Create(context.Background(), createService(notNginxGatewayServiceName))).To(Succeed())
		})

		It("should not call UpdateAddresses if the Service is not for the Gateway Pod", func() {
			e := &events.UpsertEvent{Resource: &v1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Name: "not-nginx-gateway",
				},
			}}
			batch := []interface{}{e}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(BeZero())

			de := &events.DeleteEvent{Type: &v1.Service{}}
			batch = []interface{}{de}
			Expect(fakeK8sClient.Delete(context.Background(), createService(notNginxGatewayServiceName))).To(Succeed())

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(handler.GetLatestConfiguration()).To(BeNil())

			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(BeZero())
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

			Expect(handler.GetLatestConfiguration()).To(BeNil())
			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(Equal(1))
			_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
			Expect(name).To(Equal(groupGateways))
			Expect(reqs).To(BeEmpty())
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
			Expect(fakeK8sClient.Delete(context.Background(), createService(nginxGatewayServiceName))).To(Succeed())

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(handler.GetLatestConfiguration()).To(BeNil())
			Expect(fakeStatusUpdater.UpdateGroupCallCount()).To(Equal(1))
			_, name, reqs := fakeStatusUpdater.UpdateGroupArgsForCall(0)
			Expect(name).To(Equal(groupGateways))
			Expect(reqs).To(BeEmpty())
		})
	})

	When("receiving an EndpointsOnlyChange update", func() {
		e := &events.UpsertEvent{Resource: &discoveryV1.EndpointSlice{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-gateway",
				Namespace: "nginx-gateway",
			},
		}}
		batch := []interface{}{e}

		BeforeEach(func() {
			fakeProcessor.ProcessReturns(state.EndpointsOnlyChange, &graph.Graph{})
		})

		When("running NGINX Plus", func() {
			It("should call the NGINX Plus API", func() {
				handler.cfg.plus = true

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := dataplane.GetDefaultConfiguration(&graph.Graph{}, 1)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dcfg)).To(BeEmpty())

				Expect(fakeGenerator.GenerateCallCount()).To(Equal(0))
				Expect(fakeNginxUpdater.UpdateUpstreamServersCallCount()).To(Equal(1))
			})
		})

		When("not running NGINX Plus", func() {
			It("should not call the NGINX Plus API", func() {
				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := dataplane.GetDefaultConfiguration(&graph.Graph{}, 1)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dcfg)).To(BeEmpty())

				Expect(fakeGenerator.GenerateCallCount()).To(Equal(1))
				Expect(fakeNginxUpdater.UpdateConfigCallCount()).To(Equal(1))
				Expect(fakeNginxUpdater.UpdateUpstreamServersCallCount()).To(Equal(0))
			})
		})
	})

	It("should update status when receiving a queue event", func() {
		obj := &status.QueueObject{
			Deployment: types.NamespacedName{},
			Error:      errors.New("status error"),
		}
		queue.Enqueue(obj)

		Eventually(
			func() int {
				return fakeStatusUpdater.UpdateGroupCallCount()
			}).Should(Equal(2))

		gr := handler.cfg.processor.GetLatestGraph()
		Expect(gr.LatestReloadResult.Error.Error()).To(Equal("status error"))
	})

	It("should set the health checker status properly", func() {
		e := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
		batch := []interface{}{e}
		readyChannel := handler.cfg.graphBuiltHealthChecker.getReadyCh()

		fakeProcessor.ProcessReturns(state.ClusterStateChange, &graph.Graph{})

		Expect(handler.cfg.graphBuiltHealthChecker.readyCheck(nil)).ToNot(Succeed())
		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		dcfg := dataplane.GetDefaultConfiguration(&graph.Graph{}, 1)
		Expect(helpers.Diff(handler.GetLatestConfiguration(), &dcfg)).To(BeEmpty())

		Expect(readyChannel).To(BeClosed())

		Expect(handler.cfg.graphBuiltHealthChecker.readyCheck(nil)).To(Succeed())
	})

	It("should panic for an unknown event type", func() {
		e := &struct{}{}

		handle := func() {
			batch := []interface{}{e}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
		}

		Expect(handle).Should(Panic())

		Expect(handler.GetLatestConfiguration()).To(BeNil())
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

var _ = Describe("getDeploymentContext", func() {
	When("nginx plus is false", func() {
		It("doesn't set the deployment context", func() {
			handler := eventHandlerImpl{}

			depCtx, err := handler.getDeploymentContext(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(depCtx).To(Equal(dataplane.DeploymentContext{}))
		})
	})

	When("nginx plus is true", func() {
		var ctx context.Context
		var cancel context.CancelFunc

		BeforeEach(func() {
			ctx, cancel = context.WithCancel(context.Background()) //nolint:fatcontext
		})

		AfterEach(func() {
			cancel()
		})

		It("returns deployment context", func() {
			expDepCtx := dataplane.DeploymentContext{
				Integration:      "ngf",
				ClusterID:        helpers.GetPointer("cluster-id"),
				InstallationID:   helpers.GetPointer("installation-id"),
				ClusterNodeCount: helpers.GetPointer(1),
			}

			handler := newEventHandlerImpl(eventHandlerConfig{
				ctx:         ctx,
				statusQueue: status.NewQueue(),
				plus:        true,
				deployCtxCollector: &licensingfakes.FakeCollector{
					CollectStub: func(_ context.Context) (dataplane.DeploymentContext, error) {
						return expDepCtx, nil
					},
				},
			})

			dc, err := handler.getDeploymentContext(context.Background())
			Expect(err).ToNot(HaveOccurred())
			Expect(dc).To(Equal(expDepCtx))
		})
		It("returns error if it occurs", func() {
			expErr := errors.New("collect error")

			handler := newEventHandlerImpl(eventHandlerConfig{
				ctx:         ctx,
				statusQueue: status.NewQueue(),
				plus:        true,
				deployCtxCollector: &licensingfakes.FakeCollector{
					CollectStub: func(_ context.Context) (dataplane.DeploymentContext, error) {
						return dataplane.DeploymentContext{}, expErr
					},
				},
			})

			dc, err := handler.getDeploymentContext(context.Background())
			Expect(err).To(MatchError(expErr))
			Expect(dc).To(Equal(dataplane.DeploymentContext{}))
		})
	})
})
