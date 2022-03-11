package status_test

import (
	"context"

	"github.com/nginxinc/nginx-gateway-kubernetes/internal/helpers"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/state"
	"github.com/nginxinc/nginx-gateway-kubernetes/internal/status"
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
)

var _ = Describe("Updater", func() {
	var updater status.Updater
	var client client.Client

	BeforeEach(OncePerOrdered, func() {
		scheme := runtime.NewScheme()

		Expect(gatewayv1alpha2.AddToScheme(scheme)).Should(Succeed())

		client = fake.NewClientBuilder().
			WithScheme(scheme).
			Build()
		updater = status.NewUpdater(client, zap.New())
	})

	Describe("Process status update of HTTPRoute", Ordered, func() {
		var hr *v1alpha2.HTTPRoute
		var testTime metav1.Time

		BeforeAll(func() {
			// Rfc3339Copy() removes the monotonic clock reading
			// it is important, because updating the status in the FakeClient and then getting the resource back
			// involves encoding and decoding the resource to/from JSON, which removes the monotonic clock reading.
			testTime = metav1.Now().Rfc3339Copy()

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

			Expect(client.Create(context.Background(), hr)).Should(Succeed())
		})

		It("should process status update", func() {
			updates := []state.StatusUpdate{
				{
					NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
					Status: &v1alpha2.HTTPRouteStatus{
						RouteStatus: v1alpha2.RouteStatus{
							Parents: []v1alpha2.RouteParentStatus{
								{
									ControllerName: "test",
									ParentRef: v1alpha2.ParentRef{
										Name: "fake",
									},
									Conditions: []metav1.Condition{
										{
											Type:               string(v1alpha2.ConditionRouteAccepted),
											Status:             "True",
											ObservedGeneration: hr.Generation,
											LastTransitionTime: testTime,
											Reason:             string(v1alpha2.ConditionRouteAccepted),
											Message:            "",
										},
									},
								},
							},
						},
					},
				},
			}

			updater.ProcessStatusUpdates(context.Background(), updates)
		})

		It("should have the updated status in the API server", func() {
			expectedHR := &v1alpha2.HTTPRoute{
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
								ControllerName: "test",
								ParentRef: gatewayv1alpha2.ParentRef{
									Name: "fake",
								},
								Conditions: []metav1.Condition{
									{
										Type:               string(gatewayv1alpha2.ConditionRouteAccepted),
										Status:             "True",
										ObservedGeneration: 0,
										LastTransitionTime: testTime,
										Reason:             string(gatewayv1alpha2.ConditionRouteAccepted),
										Message:            "",
									},
								},
							},
						},
					},
				},
			}

			latestHR := &v1alpha2.HTTPRoute{}

			err := client.Get(context.Background(), types.NamespacedName{Namespace: "test", Name: "route1"}, latestHR)
			Expect(err).Should(Not(HaveOccurred()))

			expectedHR.ResourceVersion = latestHR.ResourceVersion // updating the status changes the ResourceVersion

			Expect(helpers.Diff(expectedHR, latestHR)).To(BeEmpty())
		})
	})

	It("should panic for unknown status type", func() {
		updates := []state.StatusUpdate{
			{
				NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
				Status:         "unsupported",
			},
		}

		process := func() {
			updater.ProcessStatusUpdates(context.Background(), updates)
		}
		Expect(process).Should(Panic())
	})

	It("should not process updates with canceled context", func() {
		updates := []state.StatusUpdate{
			{
				NamespacedName: types.NamespacedName{Namespace: "test", Name: "route1"},
				Status:         "unsupported",
			},
		}

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// because the ctx is canceled, ProcessStatusUpdates should return immediately without panicking
		// because of the unsupported status type
		updater.ProcessStatusUpdates(ctx, updates)
	})
})
