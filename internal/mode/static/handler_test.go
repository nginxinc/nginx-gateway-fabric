package static

import (
	"context"
	"errors"

	ngxclient "github.com/nginxinc/nginx-plus-go-client/client"
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
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/graph"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/statefakes"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/staticfakes"
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
		zapLogLevelSetter   zapLogLevelSetter
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
		fakeProcessor.ProcessReturns(state.NoChange, &graph.Graph{})
		fakeGenerator = &configfakes.FakeGenerator{}
		fakeNginxFileMgr = &filefakes.FakeManager{}
		fakeNginxRuntimeMgr = &runtimefakes.FakeManager{}
		fakeStatusUpdater = &statusfakes.FakeUpdater{}
		fakeEventRecorder = record.NewFakeRecorder(1)
		zapLogLevelSetter = newZapLogLevelSetter(zap.NewAtomicLevel())

		handler = newEventHandlerImpl(eventHandlerConfig{
			k8sClient:                     fake.NewFakeClient(),
			processor:                     fakeProcessor,
			generator:                     fakeGenerator,
			logLevelSetter:                zapLogLevelSetter,
			nginxFileMgr:                  fakeNginxFileMgr,
			nginxRuntimeMgr:               fakeNginxRuntimeMgr,
			statusUpdater:                 fakeStatusUpdater,
			eventRecorder:                 fakeEventRecorder,
			nginxConfiguredOnStartChecker: newNginxConfiguredOnStartChecker(),
			controlConfigNSName:           types.NamespacedName{Namespace: namespace, Name: configName},
			gatewayPodConfig: config.GatewayPodConfig{
				ServiceName: "nginx-gateway",
				Namespace:   "nginx-gateway",
			},
			metricsCollector: collectors.NewControllerNoopCollector(),
		})
		Expect(handler.cfg.nginxConfiguredOnStartChecker.ready).To(BeFalse())
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
			fakeProcessor.ProcessReturns(state.ClusterStateChange /* changed */, &graph.Graph{})

			fakeGenerator.GenerateReturns(fakeCfgFiles)
		})

		AfterEach(func() {
			Expect(handler.cfg.nginxConfiguredOnStartChecker.ready).To(BeTrue())
		})

		When("a batch has one event", func() {
			It("should process Upsert", func() {
				e := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := &dataplane.Configuration{Version: 1}

				checkUpsertEventExpectations(e)
				expectReconfig(*dcfg, fakeCfgFiles)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), dcfg)).To(BeEmpty())
			})

			It("should process Delete", func() {
				e := &events.DeleteEvent{
					Type:           &gatewayv1.HTTPRoute{},
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route"},
				}
				batch := []interface{}{e}

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

				dcfg := &dataplane.Configuration{Version: 1}

				checkDeleteEventExpectations(e)
				expectReconfig(*dcfg, fakeCfgFiles)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), dcfg)).To(BeEmpty())
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
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dataplane.Configuration{Version: 2})).To(BeEmpty())
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

			Expect(handler.GetLatestConfiguration()).To(BeNil())

			Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
			_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
			Expect(statuses).To(Equal(expStatuses(staticConds.NewNginxGatewayValid())))
			Expect(zapLogLevelSetter.Enabled(zap.DebugLevel)).To(BeFalse())
			Expect(zapLogLevelSetter.Enabled(zap.ErrorLevel)).To(BeTrue())
		})

		It("handles an invalid config", func() {
			batch := []interface{}{&events.UpsertEvent{Resource: cfg(ngfAPI.ControllerLogLevel("invalid"))}}
			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

			Expect(handler.GetLatestConfiguration()).To(BeNil())

			Expect(fakeStatusUpdater.UpdateCallCount()).Should(Equal(1))
			_, statuses := fakeStatusUpdater.UpdateArgsForCall(0)
			cond := staticConds.NewNginxGatewayInvalid(
				"Failed to update control plane configuration: logging.level: Unsupported value: " +
					"\"invalid\": supported values: \"info\", \"debug\", \"error\"")
			Expect(statuses).To(Equal(expStatuses(cond)))
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

			Expect(fakeEventRecorder.Events).To(HaveLen(1))
			event := <-fakeEventRecorder.Events
			Expect(event).To(Equal("Warning ResourceDeleted NginxGateway configuration was deleted; using defaults"))
			Expect(zapLogLevelSetter.Enabled(zap.InfoLevel)).To(BeTrue())
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

			Expect(handler.GetLatestConfiguration()).To(BeNil())

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

			Expect(handler.GetLatestConfiguration()).To(BeNil())
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

			Expect(handler.GetLatestConfiguration()).To(BeNil())
			Expect(fakeStatusUpdater.UpdateAddressesCallCount()).ToNot(BeZero())
		})
	})

	When("receiving usage Secret updates", func() {
		var fakeSecretStore *staticfakes.FakeSecretStorer
		var usageSecret *v1.Secret

		BeforeEach(func() {
			fakeSecretStore = &staticfakes.FakeSecretStorer{}
			usageSecret = &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "usage",
					Namespace: "nginx-gateway",
				},
			}
		})

		It("should not set the N+ usage secret if not initialized", func() {
			handler.cfg.usageSecret = fakeSecretStore

			e := &events.UpsertEvent{
				Resource: usageSecret,
			}
			batch := []interface{}{e}

			handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
			Expect(fakeSecretStore.SetCallCount()).To(Equal(0))
			Expect(fakeProcessor.CaptureUpsertChangeCallCount()).To(Equal(1))
		})

		Context("usage secret is initialized", func() {
			var usageSecretHandler *eventHandlerImpl
			BeforeEach(func() {
				usageCfg := &config.UsageReportConfig{
					SecretNsName: client.ObjectKeyFromObject(usageSecret),
				}
				usageSecretHandler = newEventHandlerImpl(eventHandlerConfig{
					k8sClient:                     fake.NewFakeClient(),
					processor:                     fakeProcessor,
					nginxConfiguredOnStartChecker: newNginxConfiguredOnStartChecker(),
					controlConfigNSName:           types.NamespacedName{Namespace: namespace, Name: configName},
					usageReportConfig:             usageCfg,
					usageSecret:                   fakeSecretStore,
					gatewayPodConfig: config.GatewayPodConfig{
						ServiceName: "nginx-gateway",
						Namespace:   "nginx-gateway",
					},
					metricsCollector: collectors.NewControllerNoopCollector(),
				})
			})

			It("should not set the N+ usage secret if processing a normal secret", func() {
				e := &events.UpsertEvent{
					Resource: &v1.Secret{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "not-usage",
							Namespace: "nginx-gateway",
						},
					},
				}
				batch := []interface{}{e}

				usageSecretHandler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
				Expect(fakeSecretStore.SetCallCount()).To(Equal(0))
				Expect(fakeProcessor.CaptureUpsertChangeCallCount()).To(Equal(1))
			})

			It("should set the N+ usage secret when upserted", func() {
				e := &events.UpsertEvent{
					Resource: usageSecret,
				}
				batch := []interface{}{e}

				usageSecretHandler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
				Expect(fakeSecretStore.SetCallCount()).To(Equal(1))
				Expect(fakeProcessor.CaptureUpsertChangeCallCount()).To(Equal(1))
			})

			It("should remove the N+ usage secret when deleted", func() {
				e := &events.DeleteEvent{
					Type:           &v1.Secret{},
					NamespacedName: client.ObjectKeyFromObject(usageSecret),
				}
				batch := []interface{}{e}

				usageSecretHandler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
				Expect(fakeSecretStore.DeleteCallCount()).To(Equal(1))
				Expect(fakeProcessor.CaptureDeleteChangeCallCount()).To(Equal(1))
			})
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
			upstreams := ngxclient.Upstreams{
				"one": ngxclient.Upstream{
					Peers: []ngxclient.Peer{
						{Server: "server1"},
					},
				},
			}
			fakeNginxRuntimeMgr.GetUpstreamsReturns(upstreams, nil)
		})

		When("running NGINX Plus", func() {
			It("should call the NGINX Plus API", func() {
				fakeNginxRuntimeMgr.IsPlusReturns(true)

				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dataplane.Configuration{Version: 1})).To(BeEmpty())

				Expect(fakeGenerator.GenerateCallCount()).To(Equal(1))
				Expect(fakeNginxFileMgr.ReplaceFilesCallCount()).To(Equal(1))
				Expect(fakeNginxRuntimeMgr.GetUpstreamsCallCount()).To(Equal(1))
			})
		})

		When("not running NGINX Plus", func() {
			It("should not call the NGINX Plus API", func() {
				handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)
				Expect(helpers.Diff(handler.GetLatestConfiguration(), &dataplane.Configuration{Version: 1})).To(BeEmpty())

				Expect(fakeGenerator.GenerateCallCount()).To(Equal(1))
				Expect(fakeNginxFileMgr.ReplaceFilesCallCount()).To(Equal(1))
				Expect(fakeNginxRuntimeMgr.GetUpstreamsCallCount()).To(Equal(0))
				Expect(fakeNginxRuntimeMgr.ReloadCallCount()).To(Equal(1))
			})
		})
	})

	When("updating upstream servers", func() {
		conf := dataplane.Configuration{
			Upstreams: []dataplane.Upstream{
				{
					Name: "one",
				},
			},
		}

		type callCounts struct {
			generate int
			update   int
			reload   int
		}

		assertCallCounts := func(cc callCounts) {
			Expect(fakeGenerator.GenerateCallCount()).To(Equal(cc.generate))
			Expect(fakeNginxFileMgr.ReplaceFilesCallCount()).To(Equal(cc.generate))
			Expect(fakeNginxRuntimeMgr.UpdateHTTPServersCallCount()).To(Equal(cc.update))
			Expect(fakeNginxRuntimeMgr.ReloadCallCount()).To(Equal(cc.reload))
		}

		BeforeEach(func() {
			upstreams := ngxclient.Upstreams{
				"one": ngxclient.Upstream{
					Peers: []ngxclient.Peer{
						{Server: "server1"},
					},
				},
			}
			fakeNginxRuntimeMgr.GetUpstreamsReturns(upstreams, nil)
		})

		When("running NGINX Plus", func() {
			BeforeEach(func() {
				fakeNginxRuntimeMgr.IsPlusReturns(true)
			})

			It("should update servers using the NGINX Plus API", func() {
				Expect(handler.updateUpstreamServers(context.Background(), ctlrZap.New(), conf)).To(Succeed())

				assertCallCounts(callCounts{generate: 1, update: 1, reload: 0})
			})

			It("should reload when GET API returns an error", func() {
				fakeNginxRuntimeMgr.GetUpstreamsReturns(nil, errors.New("error"))
				Expect(handler.updateUpstreamServers(context.Background(), ctlrZap.New(), conf)).To(Succeed())

				assertCallCounts(callCounts{generate: 1, update: 0, reload: 1})
			})

			It("should reload when POST API returns an error", func() {
				fakeNginxRuntimeMgr.UpdateHTTPServersReturns(errors.New("error"))
				Expect(handler.updateUpstreamServers(context.Background(), ctlrZap.New(), conf)).To(Succeed())

				assertCallCounts(callCounts{generate: 1, update: 1, reload: 1})
			})
		})

		When("not running NGINX Plus", func() {
			It("should update servers by reloading", func() {
				Expect(handler.updateUpstreamServers(context.Background(), ctlrZap.New(), conf)).To(Succeed())

				assertCallCounts(callCounts{generate: 1, update: 0, reload: 1})
			})

			It("should return an error when reloading fails", func() {
				fakeNginxRuntimeMgr.ReloadReturns(errors.New("error"))
				Expect(handler.updateUpstreamServers(context.Background(), ctlrZap.New(), conf)).ToNot(Succeed())

				assertCallCounts(callCounts{generate: 1, update: 0, reload: 1})
			})
		})
	})

	It("should set the health checker status properly when there are changes", func() {
		e := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
		batch := []interface{}{e}
		readyChannel := handler.cfg.nginxConfiguredOnStartChecker.getReadyCh()

		fakeProcessor.ProcessReturns(state.ClusterStateChange, &graph.Graph{})

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).ToNot(Succeed())
		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(helpers.Diff(handler.GetLatestConfiguration(), &dataplane.Configuration{Version: 1})).To(BeEmpty())

		Expect(readyChannel).To(BeClosed())

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).To(Succeed())
	})

	It("should set the health checker status properly when there are no changes or errors", func() {
		e := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
		batch := []interface{}{e}
		readyChannel := handler.cfg.nginxConfiguredOnStartChecker.getReadyCh()

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).ToNot(Succeed())
		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(handler.GetLatestConfiguration()).To(BeNil())

		Expect(readyChannel).To(BeClosed())

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).To(Succeed())
	})

	It("should set the health checker status properly when there is an error", func() {
		e := &events.UpsertEvent{Resource: &gatewayv1.HTTPRoute{}}
		batch := []interface{}{e}
		readyChannel := handler.cfg.nginxConfiguredOnStartChecker.getReadyCh()

		fakeProcessor.ProcessReturns(state.ClusterStateChange, &graph.Graph{})
		fakeNginxRuntimeMgr.ReloadReturns(errors.New("reload error"))

		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).ToNot(Succeed())

		// now send an update with no changes; should still return an error
		fakeProcessor.ProcessReturns(state.NoChange, &graph.Graph{})

		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).ToNot(Succeed())

		// error goes away
		fakeProcessor.ProcessReturns(state.ClusterStateChange, &graph.Graph{})
		fakeNginxRuntimeMgr.ReloadReturns(nil)

		handler.HandleEventBatch(context.Background(), ctlrZap.New(), batch)

		Expect(helpers.Diff(handler.GetLatestConfiguration(), &dataplane.Configuration{Version: 2})).To(BeEmpty())

		Expect(readyChannel).To(BeClosed())

		Expect(handler.cfg.nginxConfiguredOnStartChecker.readyCheck(nil)).To(Succeed())
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

var _ = Describe("serversEqual", func() {
	DescribeTable("determines if server lists are equal",
		func(newServers []ngxclient.UpstreamServer, oldServers []ngxclient.Peer, equal bool) {
			Expect(serversEqual(newServers, oldServers)).To(Equal(equal))
		},
		Entry("different length",
			[]ngxclient.UpstreamServer{
				{Server: "server1"},
			},
			[]ngxclient.Peer{
				{Server: "server1"},
				{Server: "server2"},
			},
			false,
		),
		Entry("differing elements",
			[]ngxclient.UpstreamServer{
				{Server: "server1"},
				{Server: "server2"},
			},
			[]ngxclient.Peer{
				{Server: "server1"},
				{Server: "server3"},
			},
			false,
		),
		Entry("same elements",
			[]ngxclient.UpstreamServer{
				{Server: "server1"},
				{Server: "server2"},
			},
			[]ngxclient.Peer{
				{Server: "server1"},
				{Server: "server2"},
			},
			true,
		),
	)
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
