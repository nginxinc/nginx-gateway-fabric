package status_test

import (
	"context"
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
	gatewayv1beta1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status/statusfakes"
)

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

		Expect(gatewayv1beta1.AddToScheme(scheme)).Should(Succeed())

		client = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(
				&v1beta1.GatewayClass{},
				&v1beta1.Gateway{},
				&v1beta1.HTTPRoute{},
			).
			Build()

		// Rfc3339Copy() removes the monotonic clock reading and leaves only second-level precision.
		// We use it because updating the status in the FakeClient and then getting the resource back
		// involves encoding and decoding the resource to/from JSON, which uses RFC 3339 for metav1.Time.
		fakeClockTime = metav1.NewTime(time.Now()).Rfc3339Copy()
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
			updater       status.Updater
			gc            *v1beta1.GatewayClass
			gw, ignoredGw *v1beta1.Gateway
			hr            *v1beta1.HTTPRoute
			ipAddrType    = v1beta1.IPAddressType
			addr          = v1beta1.GatewayAddress{
				Type:  &ipAddrType,
				Value: "1.2.3.4",
			}

			createStatuses = func(gens generations) status.Statuses {
				return status.Statuses{
					GatewayClassStatus: &status.GatewayClassStatus{
						ObservedGeneration: gens.gatewayClass,
						Conditions:         status.CreateTestConditions("Test"),
					},
					GatewayStatuses: status.GatewayStatuses{
						{Namespace: "test", Name: "gateway"}: {
							Conditions: status.CreateTestConditions("Test"),
							ListenerStatuses: map[string]status.ListenerStatus{
								"http": {
									AttachedRoutes: 1,
									Conditions:     status.CreateTestConditions("Test"),
								},
							},
							ObservedGeneration: gens.gateways,
						},
						{Namespace: "test", Name: "ignored-gateway"}: {
							Conditions:         conditions.NewGatewayConflict(),
							ObservedGeneration: 1,
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
						Listeners: []gatewayv1beta1.ListenerStatus{
							{
								Name: "http",
								SupportedKinds: []v1beta1.RouteGroupKind{
									{
										Kind: "HTTPRoute",
									},
								},
								AttachedRoutes: 1,
								Conditions:     status.CreateExpectedAPIConditions("Test", generation, fakeClockTime),
							},
						},
						Addresses: []v1beta1.GatewayAddress{addr},
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
								Reason:             string(conditions.GatewayReasonGatewayConflict),
								Message:            conditions.GatewayMessageGatewayConflict,
							},
							{
								Type:               string(v1beta1.GatewayConditionProgrammed),
								Status:             metav1.ConditionFalse,
								ObservedGeneration: 1,
								LastTransitionTime: fakeClockTime,
								Reason:             string(conditions.GatewayReasonGatewayConflict),
								Message:            conditions.GatewayMessageGatewayConflict,
							},
						},
						Addresses: []v1beta1.GatewayAddress{addr},
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
					Status: gatewayv1beta1.HTTPRouteStatus{
						RouteStatus: gatewayv1beta1.RouteStatus{
							Parents: []gatewayv1beta1.RouteParentStatus{
								{
									ControllerName: gatewayv1beta1.GatewayController(gatewayCtrlName),
									ParentRef: gatewayv1beta1.ParentReference{
										Namespace:   (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
										Name:        "gateway",
										SectionName: (*v1beta1.SectionName)(helpers.GetStringPointer("http")),
									},
									Conditions: status.CreateExpectedAPIConditions("Test", 5, fakeClockTime),
								},
							},
						},
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
				PodIP:                    "1.2.3.4",
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
		})

		It("should create resources in the API server", func() {
			Expect(client.Create(context.Background(), gc)).Should(Succeed())
			Expect(client.Create(context.Background(), gw)).Should(Succeed())
			Expect(client.Create(context.Background(), ignoredGw)).Should(Succeed())
			Expect(client.Create(context.Background(), hr)).Should(Succeed())
		})

		It("should update statuses", func() {
			updater.Update(context.Background(), createStatuses(generations{
				gatewayClass: 1,
				gateways:     1,
			}))
		})

		It("should have the updated status of GatewayClass in the API server", func() {
			latestGc := &v1beta1.GatewayClass{}
			expectedGc := createExpectedGCWithGeneration(1)

			err := client.Get(context.Background(), types.NamespacedName{Name: gcName}, latestGc)
			Expect(err).Should(Not(HaveOccurred()))

			expectedGc.ResourceVersion = latestGc.ResourceVersion // updating the status changes the ResourceVersion

			Expect(helpers.Diff(expectedGc, latestGc)).To(BeEmpty())
		})

		It("should have the updated status of Gateway in the API server", func() {
			latestGw := &v1beta1.Gateway{}
			expectedGw := createExpectedGwWithGeneration(1)

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "gateway"}, latestGw)
			Expect(err).Should(Not(HaveOccurred()))

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
			Expect(err).Should(Not(HaveOccurred()))

			expectedGw.ResourceVersion = latestGw.ResourceVersion

			Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
		})

		It("should have the updated status of HTTPRoute in the API server", func() {
			latestHR := &v1beta1.HTTPRoute{}
			expectedHR := createExpectedHR()

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "route1"}, latestHR)
			Expect(err).Should(Not(HaveOccurred()))

			expectedHR.ResourceVersion = latestHR.ResourceVersion

			Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
		})

		It("should update statuses with canceled context - function normally returns", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			updater.Update(ctx, createStatuses(generations{
				gatewayClass: 2,
				gateways:     2,
			}))
		})

		When("updating with canceled context", func() {
			It("should have the updated status of GatewayClass in the API server", func() {
				latestGc := &v1beta1.GatewayClass{}
				expectedGc := createExpectedGCWithGeneration(2)

				err := client.Get(context.Background(), types.NamespacedName{Name: gcName}, latestGc)
				Expect(err).Should(Not(HaveOccurred()))

				expectedGc.ResourceVersion = latestGc.ResourceVersion

				Expect(helpers.Diff(expectedGc, latestGc)).To(BeEmpty())
			})

			It("should have the updated status of Gateway in the API server", func() {
				latestGw := &v1beta1.Gateway{}
				expectedGw := createExpectedGwWithGeneration(2)

				err := client.Get(
					context.Background(),
					types.NamespacedName{Namespace: "test", Name: "gateway"},
					latestGw,
				)
				Expect(err).Should(Not(HaveOccurred()))

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
				Expect(err).Should(Not(HaveOccurred()))

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
				Expect(err).Should(Not(HaveOccurred()))

				expectedHR.ResourceVersion = latestHR.ResourceVersion

				// if the status was updated, we would see the route rejected (Accepted = false)
				Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
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
				PodIP:                    "1.2.3.4",
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
				status.Statuses{
					GatewayClassStatus: &status.GatewayClassStatus{
						ObservedGeneration: 1,
						Conditions:         status.CreateTestConditions("Test"),
					},
				},
			)

			latestGc := &v1beta1.GatewayClass{}

			err := client.Get(context.Background(), types.NamespacedName{Name: gcName}, latestGc)
			Expect(err).Should(Not(HaveOccurred()))

			Expect(latestGc.Status).To(BeZero())
		})
	})
})
