package provisioner

import (
	"context"

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
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/status/statusfakes"
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

		gwNsName, depNsName types.NamespacedName
		gw                  *v1beta1.Gateway
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

		gwNsName = types.NamespacedName{
			Namespace: "test-ns",
			Name:      "test-gw",
		}
		gw = &v1beta1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: gwNsName.Namespace,
				Name:      gwNsName.Name,
			},
			Spec: v1beta1.GatewaySpec{
				GatewayClassName: gcName,
			},
		}

		depNsName = types.NamespacedName{
			Namespace: "nginx-gateway",
			Name:      "nginx-gateway-test-ns-test-gw",
		}
	})

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

	itShouldUpsertGateway := func() {
		batch := []interface{}{
			&events.UpsertEvent{
				Resource: gw,
			},
		}

		handler.HandleEventBatch(context.Background(), batch)

		dep := &v1.Deployment{}
		err := k8sclient.Get(context.Background(), depNsName, dep)

		Expect(err).ShouldNot(HaveOccurred())

		Expect(dep.ObjectMeta.Namespace).To(Equal("nginx-gateway"))
		Expect(dep.ObjectMeta.Name).To(Equal("nginx-gateway-test-ns-test-gw"))
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement("static-mode"))
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement("--gateway=test-ns/test-gw"))
		Expect(dep.Spec.Template.Spec.Containers[0].Args).To(ContainElement("--update-gatewayclass-status=false"))
	}

	itShouldPanicWhenUpsertingGateway := func() {
		batch := []interface{}{
			&events.UpsertEvent{
				Resource: gw,
			},
		}

		handle := func() {
			handler.HandleEventBatch(context.Background(), batch)
		}

		Expect(handle).Should(Panic())
	}

	Describe("Core cases", Ordered, func() {
		BeforeAll(func() {
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

		When("upserting Gateway", func() {
			It("should create Deployment", func() {
				itShouldUpsertGateway()
			})
		})

		When("upserting Gateway again", func() {
			It("must retain Deployment", func() {
				itShouldUpsertGateway()
			})
		})

		When("deleting Gateway", func() {
			It("should remove Deployment", func() {
				batch := []interface{}{
					&events.DeleteEvent{
						Type:           &v1beta1.Gateway{},
						NamespacedName: gwNsName,
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
						Name:      "test-gw-2",
						Namespace: "test-ns-2",
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
	})

	Describe("Edge cases", func() {
		BeforeEach(func() {
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
				itShouldPanicWhenUpsertingGateway()
			})
		})

		When("upserting Gateway when Deployment can't be created", func() {
			It("should panic", func() {
				itShouldUpsertGatewayClass()

				// Create a deployment so that the Handler will fail to create it because it already exists.

				dep := &v1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: depNsName.Namespace,
						Name:      depNsName.Name,
					},
				}

				err := k8sclient.Create(context.Background(), dep)
				Expect(err).ShouldNot(HaveOccurred())

				itShouldPanicWhenUpsertingGateway()
			})
		})

		When("deleting Gateway when Deployment can't be deleted", func() {
			It("should panic", func() {
				itShouldUpsertGatewayClass()
				itShouldUpsertGateway()

				// Delete the deployment so that the Handler will fail to delete it because it doesn't exist.

				dep := &v1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: depNsName.Namespace,
						Name:      depNsName.Name,
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
			itShouldPanicWhenUpsertingGateway()
		})
	})
})
