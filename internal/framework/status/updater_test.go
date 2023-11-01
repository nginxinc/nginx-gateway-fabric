package status_test

import (
	"context"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/status/statusfakes"
	staticConds "github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
)

type unsupportedStatus struct{}

func (u unsupportedStatus) APIGroup() string {
	return "unsupported"
}

var _ = Describe("Updater", func() {
	const gcName = "my-class"

	var (
		client          client.Client
		fakeClockTime   metav1.Time
		fakeClock       *statusfakes.FakeClock
		gatewayCtrlName string
	)

	BeforeEach(OncePerOrdered, func() {
		scheme := runtime.NewScheme()

		Expect(v1beta1.AddToScheme(scheme)).Should(Succeed())
		Expect(ngfAPI.AddToScheme(scheme)).Should(Succeed())

		client = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(
				&v1beta1.GatewayClass{},
				&v1beta1.Gateway{},
				&v1beta1.HTTPRoute{},
				&ngfAPI.NginxGateway{},
				&ngfAPI.NginxProxy{},
			).
			Build()

		fakeClockTime = helpers.PrepareTimeForFakeClient(metav1.NewTime(time.Now()))
		fakeClock = &statusfakes.FakeClock{}
		fakeClock.NowReturns(fakeClockTime)

		gatewayCtrlName = "test.example.com"
	})

	Describe("Process status updates", Ordered, func() {
		type generations struct {
			gatewayClass int64
			gateways     int64
		}

		var (
			updater       *status.UpdaterImpl
			gc            *v1beta1.GatewayClass
			gw, ignoredGw *v1beta1.Gateway
			hr            *v1beta1.HTTPRoute
			ng            *ngfAPI.NginxGateway
			np            *ngfAPI.NginxProxy
			addr          = v1beta1.GatewayStatusAddress{
				Type:  helpers.GetPointer(v1beta1.IPAddressType),
				Value: "1.2.3.4",
			}

			createGwAPIStatuses = func(gens generations) status.GatewayAPIStatuses {
				return status.GatewayAPIStatuses{
					GatewayClassStatuses: status.GatewayClassStatuses{
						{Name: gcName}: {
							ObservedGeneration: gens.gatewayClass,
							Conditions:         status.CreateTestConditions("Test"),
						},
					},
					GatewayStatuses: status.GatewayStatuses{
						{Namespace: "test", Name: "gateway"}: {
							Conditions: status.CreateTestConditions("Test"),
							ListenerStatuses: map[string]status.ListenerStatus{
								"http": {
									AttachedRoutes: 1,
									Conditions:     status.CreateTestConditions("Test"),
									SupportedKinds: []v1beta1.RouteGroupKind{{Kind: "HTTPRoute"}},
								},
							},
							Addresses:          []v1beta1.GatewayStatusAddress{addr},
							ObservedGeneration: gens.gateways,
						},
						{Namespace: "test", Name: "ignored-gateway"}: {
							Conditions:         staticConds.NewGatewayConflict(),
							ObservedGeneration: 1,
							Ignored:            true,
						},
					},
					HTTPRouteStatuses: status.HTTPRouteStatuses{
						{Namespace: "test", Name: "route1"}: {
							ObservedGeneration: 5,
							ParentStatuses: []status.ParentStatus{
								{
									GatewayNsName: types.NamespacedName{Namespace: "test", Name: "gateway"},
									SectionName:   helpers.GetPointer[v1beta1.SectionName]("http"),
									Conditions:    status.CreateTestConditions("Test"),
								},
							},
						},
					},
				}
			}

			createNGStatus = func(gen int64) *status.NginxGatewayStatus {
				return &status.NginxGatewayStatus{
					NsName: types.NamespacedName{
						Namespace: "nginx-gateway",
						Name:      "nginx-gateway-config",
					},
					ObservedGeneration: gen,
					Conditions:         status.CreateTestConditions("Test"),
				}
			}

			createNPStatus = func(gen int64) *status.NginxProxyStatus {
				return &status.NginxProxyStatus{
					NsName: types.NamespacedName{
						Namespace: "nginx-gateway",
						Name:      "nginx-gateway-proxy-config",
					},
					ObservedGeneration: gen,
					Conditions:         status.CreateTestConditions("Test"),
				}
			}

			createExpectedGCWithGeneration = func(generation int64) *v1beta1.GatewayClass {
				return &v1beta1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: gcName,
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "GatewayClass",
						APIVersion: "gateway.networking.k8s.io/v1beta1",
					},
					Status: v1beta1.GatewayClassStatus{
						Conditions: status.CreateExpectedAPIConditions("Test", generation, fakeClockTime),
					},
				}
			}

			createExpectedGwWithGeneration = func(generation int64) *v1beta1.Gateway {
				return &v1beta1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "gateway",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.networking.k8s.io/v1beta1",
					},
					Status: v1beta1.GatewayStatus{
						Conditions: status.CreateExpectedAPIConditions("Test", generation, fakeClockTime),
						Listeners: []v1beta1.ListenerStatus{
							{
								Name:           "http",
								AttachedRoutes: 1,
								Conditions:     status.CreateExpectedAPIConditions("Test", generation, fakeClockTime),
								SupportedKinds: []v1beta1.RouteGroupKind{{Kind: "HTTPRoute"}},
							},
						},
						Addresses: []v1beta1.GatewayStatusAddress{addr},
					},
				}
			}

			createExpectedIgnoredGw = func() *v1beta1.Gateway {
				return &v1beta1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "ignored-gateway",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.networking.k8s.io/v1beta1",
					},
					Status: v1beta1.GatewayStatus{
						Conditions: []metav1.Condition{
							{
								Type:               string(v1beta1.GatewayConditionAccepted),
								Status:             metav1.ConditionFalse,
								ObservedGeneration: 1,
								LastTransitionTime: fakeClockTime,
								Reason:             string(staticConds.GatewayReasonGatewayConflict),
								Message:            staticConds.GatewayMessageGatewayConflict,
							},
							{
								Type:               string(v1beta1.GatewayConditionProgrammed),
								Status:             metav1.ConditionFalse,
								ObservedGeneration: 1,
								LastTransitionTime: fakeClockTime,
								Reason:             string(staticConds.GatewayReasonGatewayConflict),
								Message:            staticConds.GatewayMessageGatewayConflict,
							},
						},
					},
				}
			}

			createExpectedHR = func() *v1beta1.HTTPRoute {
				return &v1beta1.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "route1",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "HTTPRoute",
						APIVersion: "gateway.networking.k8s.io/v1beta1",
					},
					Status: v1beta1.HTTPRouteStatus{
						RouteStatus: v1beta1.RouteStatus{
							Parents: []v1beta1.RouteParentStatus{
								{
									ControllerName: v1beta1.GatewayController(gatewayCtrlName),
									ParentRef: v1beta1.ParentReference{
										Namespace:   (*v1beta1.Namespace)(helpers.GetPointer("test")),
										Name:        "gateway",
										SectionName: (*v1beta1.SectionName)(helpers.GetPointer("http")),
									},
									Conditions: status.CreateExpectedAPIConditions("Test", 5, fakeClockTime),
								},
							},
						},
					},
				}
			}

			createExpectedNGWithGeneration = func(gen int64) *ngfAPI.NginxGateway {
				return &ngfAPI.NginxGateway{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nginx-gateway",
						Name:      "nginx-gateway-config",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "NginxGateway",
						APIVersion: "gateway.nginx.org/v1alpha1",
					},
					Status: ngfAPI.NginxGatewayStatus{
						Conditions: status.CreateExpectedAPIConditions("Test", gen, fakeClockTime),
					},
				}
			}

			createExpectedNPWithGeneration = func(gen int64) *ngfAPI.NginxProxy {
				return &ngfAPI.NginxProxy{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nginx-gateway",
						Name:      "nginx-gateway-proxy-config",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "NginxProxy",
						APIVersion: "gateway.nginx.org/v1alpha1",
					},
					Status: ngfAPI.NginxProxyStatus{
						Conditions: status.CreateExpectedAPIConditions("Test", gen, fakeClockTime),
					},
				}
			}
		)

		BeforeAll(func() {
			updater = status.NewUpdater(status.UpdaterConfig{
				GatewayCtlrName:          gatewayCtrlName,
				GatewayClassName:         gcName,
				Client:                   client,
				Logger:                   zap.New(),
				Clock:                    fakeClock,
				UpdateGatewayClassStatus: true,
			})

			gc = &v1beta1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: gcName,
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "GatewayClass",
					APIVersion: "gateway.networking.k8s.io/v1beta1",
				},
			}
			gw = &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "gateway",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: "gateway.networking.k8s.io/v1beta1",
				},
			}
			ignoredGw = &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "ignored-gateway",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: "gateway.networking.k8s.io/v1beta1",
				},
			}
			hr = &v1beta1.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "route1",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "HTTPRoute",
					APIVersion: "gateway.networking.k8s.io/v1beta1",
				},
			}
			ng = &ngfAPI.NginxGateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx-gateway-config",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "NginxGateway",
					APIVersion: "gateway.nginx.org/v1alpha1",
				},
			}
			np = &ngfAPI.NginxProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "nginx-gateway",
					Name:      "nginx-gateway-proxy-config",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "NginxProxy",
					APIVersion: "gateway.nginx.org/v1alpha1",
				},
			}
		})

		It("should create resources in the API server", func() {
			Expect(client.Create(context.Background(), gc)).Should(Succeed())
			Expect(client.Create(context.Background(), gw)).Should(Succeed())
			Expect(client.Create(context.Background(), ignoredGw)).Should(Succeed())
			Expect(client.Create(context.Background(), hr)).Should(Succeed())
			Expect(client.Create(context.Background(), ng)).Should(Succeed())
			Expect(client.Create(context.Background(), np)).Should(Succeed())
		})

		It("should update gateway API statuses", func() {
			updater.Update(context.Background(), createGwAPIStatuses(generations{
				gatewayClass: 1,
				gateways:     1,
			}))
		})

		It("should have the updated status of GatewayClass in the API server", func() {
			latestGc := &v1beta1.GatewayClass{}
			expectedGc := createExpectedGCWithGeneration(1)

			err := client.Get(context.Background(), types.NamespacedName{Name: gcName}, latestGc)
			Expect(err).ToNot(HaveOccurred())

			expectedGc.ResourceVersion = latestGc.ResourceVersion // updating the status changes the ResourceVersion

			Expect(helpers.Diff(expectedGc, latestGc)).To(BeEmpty())
		})

		It("should have the updated status of Gateway in the API server", func() {
			latestGw := &v1beta1.Gateway{}
			expectedGw := createExpectedGwWithGeneration(1)

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "gateway"}, latestGw)
			Expect(err).ToNot(HaveOccurred())

			expectedGw.ResourceVersion = latestGw.ResourceVersion

			Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
		})

		It("should have the updated status of ignored Gateway in the API server", func() {
			latestGw := &v1beta1.Gateway{}
			expectedGw := createExpectedIgnoredGw()

			err := client.Get(
				context.Background(),
				types.NamespacedName{Namespace: "test", Name: "ignored-gateway"},
				latestGw,
			)
			Expect(err).ToNot(HaveOccurred())

			expectedGw.ResourceVersion = latestGw.ResourceVersion

			Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
		})

		It("should have the updated status of HTTPRoute in the API server", func() {
			latestHR := &v1beta1.HTTPRoute{}
			expectedHR := createExpectedHR()

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "route1"}, latestHR)
			Expect(err).ToNot(HaveOccurred())

			expectedHR.ResourceVersion = latestHR.ResourceVersion

			Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
		})

		It("should update nginx gateway status", func() {
			updater.Update(context.Background(), createNGStatus(1))
		})

		It("should have the updated status of NginxGateway in the API server", func() {
			latestNG := &ngfAPI.NginxGateway{}
			expectedNG := createExpectedNGWithGeneration(1)

			err := client.Get(
				context.Background(),
				types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx-gateway-config"},
				latestNG,
			)
			Expect(err).ToNot(HaveOccurred())

			expectedNG.ResourceVersion = latestNG.ResourceVersion

			Expect(helpers.Diff(expectedNG, latestNG)).To(BeEmpty())
		})

		It("should update nginx proxy status", func() {
			updater.Update(context.Background(), createNPStatus(1))
		})

		It("should have the updated status of NginxProxy in the API server", func() {
			latestNP := &ngfAPI.NginxProxy{}
			expectedNP := createExpectedNPWithGeneration(1)

			err := client.Get(
				context.Background(),
				types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx-gateway-proxy-config"},
				latestNP,
			)
			Expect(err).ToNot(HaveOccurred())

			expectedNP.ResourceVersion = latestNP.ResourceVersion

			Expect(helpers.Diff(expectedNP, latestNP)).To(BeEmpty())
		})

		When("the Gateway Service is updated with a new address", func() {
			AfterEach(func() {
				// reset the IP for the remaining tests
				updater.UpdateAddresses(context.Background(), []v1beta1.GatewayStatusAddress{
					{
						Type:  helpers.GetPointer(v1beta1.IPAddressType),
						Value: "1.2.3.4",
					},
				})
			})

			It("should update the previous Gateway statuses with new address", func() {
				latestGw := &v1beta1.Gateway{}
				expectedGw := createExpectedGwWithGeneration(1)
				expectedGw.Status.Addresses[0].Value = "5.6.7.8"

				updater.UpdateAddresses(context.Background(), []v1beta1.GatewayStatusAddress{
					{
						Type:  helpers.GetPointer(v1beta1.IPAddressType),
						Value: "5.6.7.8",
					},
				})

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedGw.ResourceVersion = latestGw.ResourceVersion

				Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
			})
		})

		It("should not update Gateway API statuses with canceled context - function normally returns", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			updater.Update(ctx, createGwAPIStatuses(generations{
				gatewayClass: 2,
				gateways:     2,
			}))
		})

		When("updating with canceled context", func() {
			It("should not have the updated status of GatewayClass in the API server", func() {
				latestGc := &v1beta1.GatewayClass{}
				expectedGc := createExpectedGCWithGeneration(1)

				err := client.Get(context.Background(), types.NamespacedName{Name: gcName}, latestGc)
				Expect(err).ToNot(HaveOccurred())

				expectedGc.ResourceVersion = latestGc.ResourceVersion

				Expect(helpers.Diff(expectedGc, latestGc)).To(BeEmpty())
			})

			It("should not have the updated status of Gateway in the API server", func() {
				latestGw := &v1beta1.Gateway{}
				expectedGw := createExpectedGwWithGeneration(1)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedGw.ResourceVersion = latestGw.ResourceVersion

				Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
			})

			It("should not have the updated status of ignored Gateway in the API server", func() {
				latestGw := &v1beta1.Gateway{}
				expectedGw := createExpectedIgnoredGw()

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "ignored-gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedGw.ResourceVersion = latestGw.ResourceVersion

				// if the status was updated, we would see a different ObservedGeneration
				Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
			})

			It("should not have the updated status of HTTPRoute in the API server", func() {
				latestHR := &v1beta1.HTTPRoute{}
				expectedHR := createExpectedHR()

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "route1"},
					latestHR,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedHR.ResourceVersion = latestHR.ResourceVersion

				// if the status was updated, we would see the route rejected (Accepted = false)
				Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
			})
		})

		It("should not update NginxGateway status with canceled context - function normally returns", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			updater.Update(ctx, createNGStatus(2))
		})

		When("updating with canceled context", func() {
			It("should not have the updated status of the NginxGateway in the API server", func() {
				latestNG := &ngfAPI.NginxGateway{}
				expectedNG := createExpectedNGWithGeneration(1)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx-gateway-config"},
					latestNG,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedNG.ResourceVersion = latestNG.ResourceVersion

				Expect(helpers.Diff(expectedNG, latestNG)).To(BeEmpty())
			})
		})

		When("the Pod is not the current leader", func() {
			It("should not update any statuses", func() {
				updater.Disable()
				updater.Update(context.Background(), createGwAPIStatuses(generations{
					gateways: 3,
				}))
				updater.Update(context.Background(), createNGStatus(2))
			})

			It("should not have the updated status of Gateway in the API server", func() {
				latestGw := &v1beta1.Gateway{}
				// testing that the generation has not changed from 1 to 3
				expectedGw := createExpectedGwWithGeneration(1)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedGw.ResourceVersion = latestGw.ResourceVersion

				Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
			})

			It("should not have the updated status of the Nginx Gateway resource in the API server", func() {
				latestNG := &ngfAPI.NginxGateway{}
				expectedNG := createExpectedNGWithGeneration(1)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx-gateway-config"},
					latestNG,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedNG.ResourceVersion = latestNG.ResourceVersion

				Expect(helpers.Diff(expectedNG, latestNG)).To(BeEmpty())
			})
		})
		When("the Pod starts leading", func() {
			It("writes the last statuses", func() {
				updater.Enable(context.Background())
			})

			It("should have the updated status of Gateway in the API server", func() {
				latestGw := &v1beta1.Gateway{}
				expectedGw := createExpectedGwWithGeneration(3)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedGw.ResourceVersion = latestGw.ResourceVersion

				Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
			})

			It("should have the updated status of the Nginx Gateway resource in the API server", func() {
				latestNG := &ngfAPI.NginxGateway{}
				expectedNG := createExpectedNGWithGeneration(2)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx-gateway-config"},
					latestNG,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedNG.ResourceVersion = latestNG.ResourceVersion

				Expect(helpers.Diff(expectedNG, latestNG)).To(BeEmpty())
			})
		})

		When("the Pod is the current leader", func() {
			It("should update Gateway API statuses", func() {
				updater.Update(context.Background(), createGwAPIStatuses(generations{
					gateways: 4,
				}))
			})

			It("should have the updated status of Gateway in the API server", func() {
				latestGw := &v1beta1.Gateway{}
				expectedGw := createExpectedGwWithGeneration(4)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedGw.ResourceVersion = latestGw.ResourceVersion

				Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
			})

			It("should update Nginx Gateway status", func() {
				updater.Update(context.Background(), createNGStatus(3))
			})
			It("should have the updated status of Nginx Gateway in the API server", func() {
				latestNG := &ngfAPI.NginxGateway{}
				expectedNG := createExpectedNGWithGeneration(3)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "nginx-gateway", Name: "nginx-gateway-config"},
					latestNG,
				)
				Expect(err).ToNot(HaveOccurred())

				expectedNG.ResourceVersion = latestNG.ResourceVersion

				Expect(helpers.Diff(expectedNG, latestNG)).To(BeEmpty())
			})
			It("updates and writes last statuses synchronously", func() {
				wg := &sync.WaitGroup{}
				ctx := context.Background()

				// Spin up 10 goroutines that Update and 10 that call Enable which writes the last statuses.
				// Since we only write statuses when they've changed, we will only update the status 10 times.
				// The purpose of this test is to exercise the locking mechanism embedded in the updater.
				// If there is a data race, this test combined with the -race flag that we run tests with,
				// should catch it.
				for i := 0; i < 10; i++ {
					wg.Add(2)
					gen := 5 + i
					go func() {
						updater.Update(ctx, createGwAPIStatuses(generations{gateways: int64(gen)}))
						wg.Done()
					}()

					go func() {
						updater.Enable(ctx)
						wg.Done()
					}()
				}

				wg.Wait()

				latestGw := &v1beta1.Gateway{}

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).ToNot(HaveOccurred())

				// Before this test there were 6 updates to the Gateway resource.
				// So now the resource version should equal 16.
				Expect(latestGw.ResourceVersion).To(Equal("16"))
			})
		})
	})

	Describe("Skip GatewayClass updates", Ordered, func() {
		var (
			updater status.Updater
			gc      *v1beta1.GatewayClass
		)

		BeforeAll(func() {
			updater = status.NewUpdater(status.UpdaterConfig{
				GatewayCtlrName:          gatewayCtrlName,
				GatewayClassName:         gcName,
				Client:                   client,
				Logger:                   zap.New(),
				Clock:                    fakeClock,
				UpdateGatewayClassStatus: false,
			})

			gc = &v1beta1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: gcName,
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "GatewayClass",
					APIVersion: "gateway.networking.k8s.io/v1beta1",
				},
			}
		})

		It("should create resources in the API server", func() {
			Expect(client.Create(context.Background(), gc)).Should(Succeed())
		})

		It("should not update GatewayClass status", func() {
			updater.Update(
				context.Background(),
				status.GatewayAPIStatuses{
					GatewayClassStatuses: status.GatewayClassStatuses{
						{Name: gcName}: {
							ObservedGeneration: 1,
							Conditions:         status.CreateTestConditions("Test"),
						},
					},
				},
			)

			latestGc := &v1beta1.GatewayClass{}

			err := client.Get(context.Background(), types.NamespacedName{Name: gcName}, latestGc)
			Expect(err).ToNot(HaveOccurred())

			Expect(latestGc.Status).To(BeZero())
		})
	})

	Describe("Edge cases", func() {
		It("panics on update if status type is unknown", func() {
			updater := status.NewUpdater(status.UpdaterConfig{
				GatewayCtlrName:          gatewayCtrlName,
				GatewayClassName:         gcName,
				Client:                   client,
				Logger:                   zap.New(),
				Clock:                    fakeClock,
				UpdateGatewayClassStatus: true,
			})

			update := func() {
				updater.Update(context.Background(), unsupportedStatus{})
			}

			Expect(update).Should(Panic())
		})
	})
})
