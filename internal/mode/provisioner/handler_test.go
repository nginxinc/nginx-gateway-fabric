package provisioner

import (
	"context"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	. "github.com/onsi/gomega"

	embeddedfiles "github.com/nginxinc/nginx-kubernetes-gateway"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/conditions"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/framework/status/statusfakes"
)

var _ = Describe("handler", func() {
	const (
		gcName = "test-gc"
	)
	var (
		handler       *eventHandler
		fakeClockTime metav1.Time

		statusUpdater status.Updater
		k8sclient     client.Client
	)

	BeforeEach(OncePerOrdered, func() {
		scheme := runtime.NewScheme()

		Expect(v1beta1.AddToScheme(scheme)).Should(Succeed())
		Expect(v1.AddToScheme(scheme)).Should(Succeed())

		k8sclient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithStatusSubresource(
				&v1beta1.Gateway{},
				&v1beta1.GatewayClass{},
			).
			Build()

		fakeClockTime = helpers.PrepareTimeForFakeClient(metav1.Now())
		fakeClock := &statusfakes.FakeClock{}
		fakeClock.NowReturns(fakeClockTime)

		statusUpdater = status.NewUpdater(status.UpdaterConfig{
			Client:                   k8sclient,
			Clock:                    fakeClock,
			Logger:                   zap.New(),
			GatewayCtlrName:          "test.example.com",
			GatewayClassName:         gcName,
			PodIP:                    "1.2.3.4",
			UpdateGatewayClassStatus: true,
		})
	})

	createGateway := func(gwNsName types.NamespacedName) *v1beta1.Gateway {
		return &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: gwNsName.Namespace,
				Name:      gwNsName.Name,
			},
			Spec: v1beta1.GatewaySpec{
				GatewayClassName: gcName,
			},
		}
	}

	itShouldUpsertGatewayClass := func() {
		// Add GatewayClass to the cluster

		gc := &v1beta1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name: gcName,
			},
		}

		err := k8sclient.Create(context.Background(), gc)
		Expect(err).ShouldNot(HaveOccurred())

		// UpsertGatewayClass

		batch := []interface{}{
			&events.UpsertEvent{
				Resource: gc,
			},
		}
		handler.HandleEventBatch(context.Background(), batch)

		// Ensure GatewayClass is accepted

		clusterGc := &v1beta1.GatewayClass{}
		err = k8sclient.Get(context.Background(), client.ObjectKeyFromObject(gc), clusterGc)

		Expect(err).ShouldNot(HaveOccurred())

		expectedConditions := []metav1.Condition{
			{
				Type:               string(v1beta1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				ObservedGeneration: 0,
				LastTransitionTime: fakeClockTime,
				Reason:             "Accepted",
				Message:            "GatewayClass is accepted",
			},
		}

		Expect(clusterGc.Status.Conditions).To(Equal(expectedConditions))
	}

	itShouldUpsertGateway := func(gwNsName types.NamespacedName, seqNumber int64) {
		batch := []interface{}{
			&events.UpsertEvent{
				Resource: createGateway(gwNsName),
			},
		}

		handler.HandleEventBatch(context.Background(), batch)

		depNsName := types.NamespacedName{
			Namespace: "nginx-gateway",
			Name:      fmt.Sprintf("nginx-gateway-%d", seqNumber),
		}

		dep := &v1.Deployment{}
		err := k8sclient.Get(context.Background(), depNsName, dep)

		Expect(err).ShouldNot(HaveOccurred())

		Expect(dep.ObjectMeta.Namespace).To(Equal("nginx-gateway"))
		Expect(dep.ObjectMeta.Name).To(Equal(depNsName.Name))
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement("static-mode"))
		expectedGwFlag := fmt.Sprintf("--gateway=%s", gwNsName.String())
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement(expectedGwFlag))
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement("--update-gatewayclass-status=false"))
		expectedLockFlag := fmt.Sprintf("--leader-election-lock-name=%s", gwNsName.Name)
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement(expectedLockFlag))
	}

	itShouldPanicWhenUpsertingGateway := func(gwNsName types.NamespacedName) {
		batch := []interface{}{
			&events.UpsertEvent{
				Resource: createGateway(gwNsName),
			},
		}

		handle := func() {
			handler.HandleEventBatch(context.Background(), batch)
		}

		Expect(handle).Should(Panic())
	}

	Describe("Core cases", Ordered, func() {
		var gwNsName1, gwNsName2 types.NamespacedName

		BeforeAll(func() {
			gwNsName1 = types.NamespacedName{
				Namespace: "test-ns-1",
				Name:      "test-gw-1",
			}
			gwNsName2 = types.NamespacedName{
				Namespace: "test-ns-2",
				Name:      "test-gw-2",
			}

			handler = newEventHandler(
				gcName,
				statusUpdater,
				k8sclient,
				zap.New(),
				embeddedfiles.StaticModeDeploymentYAML,
			)
		})

		When("upserting GatewayClass", func() {
			It("should make GatewayClass Accepted", func() {
				itShouldUpsertGatewayClass()
			})
		})

		When("upserting first Gateway", func() {
			It("should create first Deployment", func() {
				itShouldUpsertGateway(gwNsName1, 1)
			})
		})

		When("upserting first Gateway again", func() {
			It("must retain Deployment", func() {
				itShouldUpsertGateway(gwNsName1, 1)
			})
		})

		When("upserting second Gateway", func() {
			It("should create second Deployment", func() {
				itShouldUpsertGateway(gwNsName2, 2)
			})
		})

		When("deleting first Gateway", func() {
			It("should remove first Deployment", func() {
				batch := []interface{}{
					&events.DeleteEvent{
						Type:           &v1beta1.Gateway{},
						NamespacedName: gwNsName1,
					},
				}

				handler.HandleEventBatch(context.Background(), batch)
				deps := &v1.DeploymentList{}

				err := k8sclient.List(context.Background(), deps)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(deps.Items).To(HaveLen(1))
				Expect(deps.Items[0].ObjectMeta.Name).To(Equal("nginx-gateway-2"))
			})
		})

		When("deleting second Gateway", func() {
			It("should remove second Deployment", func() {
				batch := []interface{}{
					&events.DeleteEvent{
						Type:           &v1beta1.Gateway{},
						NamespacedName: gwNsName2,
					},
				}

				handler.HandleEventBatch(context.Background(), batch)

				deps := &v1.DeploymentList{}

				err := k8sclient.List(context.Background(), deps)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(deps.Items).To(HaveLen(0))
			})
		})

		When("upserting Gateway for a different GatewayClass", func() {
			It("should not create Deployment", func() {
				gw := &v1beta1.Gateway{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-gw-3",
						Namespace: "test-ns-3",
					},
					Spec: v1beta1.GatewaySpec{
						GatewayClassName: "some-class",
					},
				}

				batch := []interface{}{
					&events.UpsertEvent{
						Resource: gw,
					},
				}

				handler.HandleEventBatch(context.Background(), batch)

				deps := &v1.DeploymentList{}
				err := k8sclient.List(context.Background(), deps)

				Expect(err).ShouldNot(HaveOccurred())
				Expect(deps.Items).To(HaveLen(0))
			})
		})

		When("upserting GatewayClass that is not set in command-line argument", func() {
			It("should set the proper status if this controller is referenced", func() {
				gc := &v1beta1.GatewayClass{
					ObjectMeta: metav1.ObjectMeta{
						Name: "unknown-gc",
					},
					Spec: v1beta1.GatewayClassSpec{
						ControllerName: "test.example.com",
					},
				}
				err := k8sclient.Create(context.Background(), gc)
				Expect(err).ShouldNot(HaveOccurred())

				batch := []interface{}{
					&events.UpsertEvent{
						Resource: gc,
					},
				}

				handler.HandleEventBatch(context.Background(), batch)

				unknownGC := &v1beta1.GatewayClass{}
				err = k8sclient.Get(context.Background(), client.ObjectKeyFromObject(gc), unknownGC)
				Expect(err).ShouldNot(HaveOccurred())

				expectedConditions := []metav1.Condition{
					{
						Type:               string(v1beta1.GatewayClassConditionStatusAccepted),
						Status:             metav1.ConditionFalse,
						ObservedGeneration: 0,
						LastTransitionTime: fakeClockTime,
						Reason:             string(conditions.GatewayClassReasonGatewayClassConflict),
						Message:            string(conditions.GatewayClassMessageGatewayClassConflict),
					},
				}

				Expect(unknownGC.Status.Conditions).To(Equal(expectedConditions))
			})
		})
	})

	Describe("Edge cases", func() {
		var gwNsName types.NamespacedName

		BeforeEach(func() {
			gwNsName = types.NamespacedName{
				Namespace: "test-ns",
				Name:      "test-gw",
			}

			handler = newEventHandler(
				gcName,
				statusUpdater,
				k8sclient,
				zap.New(),
				embeddedfiles.StaticModeDeploymentYAML,
			)
		})

		DescribeTable("Edge cases for events",
			func(e interface{}) {
				batch := []interface{}{e}

				handle := func() {
					handler.HandleEventBatch(context.TODO(), batch)
				}

				Expect(handle).Should(Panic())
			},
			Entry("should panic for an unknown event type",
				&struct{}{}),
			Entry("should panic for an unknown type of resource in upsert event",
				&events.UpsertEvent{
					Resource: &v1beta1.HTTPRoute{},
				}),
			Entry("should panic for an unknown type of resource in delete event",
				&events.DeleteEvent{
					Type: &v1beta1.HTTPRoute{},
				}),
		)

		When("upserting Gateway when GatewayClass doesn't exist", func() {
			It("should panic", func() {
				itShouldPanicWhenUpsertingGateway(gwNsName)
			})
		})

		When("upserting Gateway when Deployment can't be created", func() {
			It("should panic", func() {
				itShouldUpsertGatewayClass()

				// Create a deployment so that the Handler will fail to create it because it already exists.

				dep := &v1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nginx-gateway",
						Name:      "nginx-gateway-1",
					},
				}

				err := k8sclient.Create(context.Background(), dep)
				Expect(err).ShouldNot(HaveOccurred())

				itShouldPanicWhenUpsertingGateway(gwNsName)
			})
		})

		When("deleting Gateway when Deployment can't be deleted", func() {
			It("should panic", func() {
				itShouldUpsertGatewayClass()
				itShouldUpsertGateway(gwNsName, 1)

				// Delete the deployment so that the Handler will fail to delete it because it doesn't exist.

				dep := &v1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "nginx-gateway",
						Name:      "nginx-gateway-1",
					},
				}

				err := k8sclient.Delete(context.Background(), dep)
				Expect(err).ShouldNot(HaveOccurred())

				batch := []interface{}{
					&events.DeleteEvent{
						Type:           &v1beta1.Gateway{},
						NamespacedName: gwNsName,
					},
				}

				handle := func() {
					handler.HandleEventBatch(context.Background(), batch)
				}

				Expect(handle).Should(Panic())
			})
		})

		When("deleting GatewayClass", func() {
			It("should panic", func() {
				itShouldUpsertGatewayClass()

				batch := []interface{}{
					&events.DeleteEvent{
						Type: &v1beta1.GatewayClass{},
						NamespacedName: types.NamespacedName{
							Name: gcName,
						},
					},
				}

				handle := func() {
					handler.HandleEventBatch(context.Background(), batch)
				}

				Expect(handle).Should(Panic())
			})
		})

		When("upserting Gateway with broken static Deployment YAML", func() {
			It("it should panic", func() {
				handler = newEventHandler(
					gcName,
					statusUpdater,
					k8sclient,
					zap.New(),
					[]byte("broken YAML"),
				)

				itShouldUpsertGatewayClass()
				itShouldPanicWhenUpsertingGateway(types.NamespacedName{Namespace: "test-ns", Name: "test-gw"})
			})
		})
	})
})
