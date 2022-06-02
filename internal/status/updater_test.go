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
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/newstate"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
)

var _ = Describe("Updater", func() {
	var (
		updater         status.Updater
		client          client.Client
		fakeClockTime   metav1.Time
		fakeClock       *status.FakeClock
		gatewayCtrlName string
		gwNsName        types.NamespacedName
	)

	BeforeEach(OncePerOrdered, func() {
		scheme := runtime.NewScheme()

		Expect(gatewayv1alpha2.AddToScheme(scheme)).Should(Succeed())

		client = fake.NewClientBuilder().
			WithScheme(scheme).
			Build()

		// Rfc3339Copy() removes the monotonic clock reading
		// We need to remove it, because updating the status in the FakeClient and then getting the resource back
		// involves encoding and decoding the resource to/from JSON, which removes the monotonic clock reading.
		fakeClockTime = metav1.NewTime(time.Now()).Rfc3339Copy()
		fakeClock = status.NewFakeClock(fakeClockTime.Time)

		gatewayCtrlName = "test.example.com"
		gwNsName = types.NamespacedName{Namespace: "test", Name: "gateway"}

		updater = status.NewUpdater(gatewayCtrlName, gwNsName, client, zap.New(), fakeClock)
	})

	Describe("Process status updates", Ordered, func() {
		var (
			hr *v1alpha2.HTTPRoute
			gw *v1alpha2.Gateway

			createStatuses   func(bool, bool) newstate.Statuses
			createExpectedHR func() *v1alpha2.HTTPRoute
			createExpectedGw func(metav1.ConditionStatus, string) *v1alpha2.Gateway
		)

		BeforeAll(func() {
			hr = &v1alpha2.HTTPRoute{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "route1",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "HTTPRoute",
					APIVersion: "gateway.networking.k8s.io/v1alpha2",
				},
			}
			gw = &v1alpha2.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "gateway",
				},
				TypeMeta: metav1.TypeMeta{
					Kind:       "Gateway",
					APIVersion: "gateway.networking.k8s.io/v1alpha2",
				},
			}

			createStatuses = func(listenerValid, routeAttached bool) newstate.Statuses {
				return newstate.Statuses{
					ListenerStatuses: map[string]newstate.ListenerStatus{
						"http": {
							Valid:          listenerValid,
							AttachedRoutes: 1,
						},
					},
					HTTPRouteStatuses: map[types.NamespacedName]newstate.HTTPRouteStatus{
						{Namespace: "test", Name: "route1"}: {
							ParentStatuses: map[string]newstate.ParentStatus{
								"http": {
									Attached: routeAttached,
								},
							},
						},
					},
				}
			}

			createExpectedHR = func() *v1alpha2.HTTPRoute {
				return &v1alpha2.HTTPRoute{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "route1",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "HTTPRoute",
						APIVersion: "gateway.networking.k8s.io/v1alpha2",
					},
					Status: gatewayv1alpha2.HTTPRouteStatus{
						RouteStatus: gatewayv1alpha2.RouteStatus{
							Parents: []gatewayv1alpha2.RouteParentStatus{
								{
									ControllerName: gatewayv1alpha2.GatewayController(gatewayCtrlName),
									ParentRef: gatewayv1alpha2.ParentRef{
										Namespace:   (*v1alpha2.Namespace)(helpers.GetStringPointer("test")),
										Name:        "gateway",
										SectionName: (*v1alpha2.SectionName)(helpers.GetStringPointer("http")),
									},
									Conditions: []metav1.Condition{
										{
											Type:               string(gatewayv1alpha2.ConditionRouteAccepted),
											Status:             "True",
											ObservedGeneration: 123,
											LastTransitionTime: fakeClockTime,
											Reason:             "Attached",
										},
									},
								},
							},
						},
					},
				}
			}

			createExpectedGw = func(status metav1.ConditionStatus, reason string) *v1alpha2.Gateway {
				return &v1alpha2.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "gateway",
					},
					TypeMeta: metav1.TypeMeta{
						Kind:       "Gateway",
						APIVersion: "gateway.networking.k8s.io/v1alpha2",
					},
					Status: v1alpha2.GatewayStatus{
						Listeners: []gatewayv1alpha2.ListenerStatus{
							{
								Name: "http",
								SupportedKinds: []v1alpha2.RouteGroupKind{
									{
										Kind: "HTTPRoute",
									},
								},
								AttachedRoutes: 1,
								Conditions: []metav1.Condition{
									{
										Type:               string(v1alpha2.ListenerConditionReady),
										Status:             status,
										ObservedGeneration: 123,
										LastTransitionTime: fakeClockTime,
										Reason:             reason,
									},
								},
							},
						},
					},
				}
			}
		})

		It("should create resources in the API server", func() {
			Expect(client.Create(context.Background(), hr)).Should(Succeed())
			Expect(client.Create(context.Background(), gw)).Should(Succeed())
		})

		It("should update statuses", func() {
			updater.Update(context.Background(), createStatuses(true, true))
		})

		It("should have the updated status of HTTPRoute in the API server", func() {
			latestHR := &v1alpha2.HTTPRoute{}
			expectedHR := createExpectedHR()

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "route1"}, latestHR)
			Expect(err).Should(Not(HaveOccurred()))

			expectedHR.ResourceVersion = latestHR.ResourceVersion // updating the status changes the ResourceVersion

			Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
		})

		It("should have the updated status of Gateway in the API server", func() {
			latestGw := &v1alpha2.Gateway{}
			expectedGw := createExpectedGw("True", string(v1alpha2.ListenerReasonReady))

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "gateway"}, latestGw)
			Expect(err).Should(Not(HaveOccurred()))

			expectedGw.ResourceVersion = latestGw.ResourceVersion

			Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
		})

		It("should update statuses with canceled context - function normally returns", func() {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			updater.Update(ctx, createStatuses(false, false))
		})

		It("should not have the updated status of HTTPRoute in the API server after updating with canceled context", func() {
			latestHR := &v1alpha2.HTTPRoute{}
			expectedHR := createExpectedHR()

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "route1"}, latestHR)
			Expect(err).Should(Not(HaveOccurred()))

			expectedHR.ResourceVersion = latestHR.ResourceVersion

			Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
		})

		It("should have the updated status of Gateway in the API server after updating with canceled context", func() {
			latestGw := &v1alpha2.Gateway{}
			expectedGw := createExpectedGw("False", string(v1alpha2.ListenerReasonInvalid))

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "gateway"}, latestGw)
			Expect(err).Should(Not(HaveOccurred()))

			expectedGw.ResourceVersion = latestGw.ResourceVersion

			Expect(helpers.Diff(expectedGw, latestGw)).To(BeEmpty())
		})
	})
})
