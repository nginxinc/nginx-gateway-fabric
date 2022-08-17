package events_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events/eventsfakes"
)

var _ = Describe("FirstEventBatchPreparer", func() {
	var (
		fakeReader *eventsfakes.FakeReader
		preparer   *events.FirstEventBatchPreparerImpl
	)

	const gcName = "my-class"

	BeforeEach(func() {
		fakeReader = &eventsfakes.FakeReader{}
		preparer = events.NewFirstEventBatchPreparerImpl(fakeReader, gcName)
	})

	Describe("Normal cases", func() {
		AfterEach(func() {
			Expect(fakeReader.GetCallCount()).Should(Equal(1))
			Expect(fakeReader.ListCallCount()).Should(Equal(4))
		})

		It("should prepare zero events when resources don't exist", func() {
			fakeReader.GetCalls(func(ctx context.Context, name types.NamespacedName, object client.Object) error {
				Expect(name).Should(Equal(types.NamespacedName{Name: gcName}))
				Expect(object).Should(BeAssignableToTypeOf(&v1beta1.GatewayClass{}))

				return apierrors.NewNotFound(schema.GroupResource{}, "test")
			})
			fakeReader.ListReturns(nil)

			batch, err := preparer.Prepare(context.Background())

			Expect(batch).Should(BeEmpty())
			Expect(err).Should(BeNil())
		})

		It("should prepare one event for each resource type", func() {
			const resourceName = "resource"

			gatewayClass := v1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: gcName}}

			fakeReader.GetCalls(func(ctx context.Context, name types.NamespacedName, object client.Object) error {
				Expect(name).Should(Equal(types.NamespacedName{Name: gcName}))
				Expect(object).Should(BeAssignableToTypeOf(&v1beta1.GatewayClass{}))

				reflect.Indirect(reflect.ValueOf(object)).Set(reflect.Indirect(reflect.ValueOf(&gatewayClass)))
				return nil
			})

			service := apiv1.Service{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}
			secret := apiv1.Secret{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}
			gateway := v1beta1.Gateway{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}
			httpRoute := v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: resourceName}}

			fakeReader.ListCalls(func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
				Expect(option).To(BeEmpty())

				switch typedList := list.(type) {
				case *apiv1.ServiceList:
					typedList.Items = append(typedList.Items, service)
				case *apiv1.SecretList:
					typedList.Items = append(typedList.Items, secret)
				case *v1beta1.GatewayList:
					typedList.Items = append(typedList.Items, gateway)
				case *v1beta1.HTTPRouteList:
					typedList.Items = append(typedList.Items, httpRoute)
				default:
					Fail(fmt.Sprintf("unknown type: %T", typedList))
				}

				return nil
			})

			expectedBatch := events.EventBatch{
				&events.UpsertEvent{Resource: &gatewayClass},
				&events.UpsertEvent{Resource: &service},
				&events.UpsertEvent{Resource: &secret},
				&events.UpsertEvent{Resource: &gateway},
				&events.UpsertEvent{Resource: &httpRoute},
			}

			batch, err := preparer.Prepare(context.Background())

			Expect(batch).Should(Equal(expectedBatch))
			Expect(err).Should(BeNil())
		})
	})

	Describe("Edge cases", func() {
		DescribeTable("CachedReader returns errors",
			func(obj client.Object) {
				readerError := errors.New("test")

				fakeReader.GetReturns(nil)
				fakeReader.ListReturns(nil)

				switch obj.(type) {
				case *v1beta1.GatewayClass:
					fakeReader.GetReturns(readerError)
				case *apiv1.Service:
					fakeReader.ListReturnsOnCall(0, readerError)
				case *apiv1.Secret:
					fakeReader.ListReturnsOnCall(1, readerError)
				case *v1beta1.Gateway:
					fakeReader.ListReturnsOnCall(2, readerError)
				case *v1beta1.HTTPRoute:
					fakeReader.ListReturnsOnCall(3, readerError)
				default:
					Fail(fmt.Sprintf("Unknown type: %T", obj))
				}

				batch, err := preparer.Prepare(context.Background())
				Expect(batch).To(BeNil())
				Expect(err).To(MatchError(readerError))
			},
			Entry("Service", &apiv1.Service{}),
			Entry("Secret", &apiv1.Secret{}),
			Entry("GatewayClass", &v1beta1.GatewayClass{}),
			Entry("Gateway", &v1beta1.Gateway{}),
			Entry("HTTPRoute", &v1beta1.HTTPRoute{}),
		)
	})
})
