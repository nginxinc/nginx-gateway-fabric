package events_test

import (
	"context"
	"errors"
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
		preparer = events.NewFirstEventBatchPreparerImpl(
			fakeReader,
			[]client.Object{&v1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: gcName}}},
			[]client.ObjectList{
				&v1beta1.HTTPRouteList{},
			})
	})

	Describe("Normal cases", func() {
		AfterEach(func() {
			Expect(fakeReader.GetCallCount()).Should(Equal(1))
			Expect(fakeReader.ListCallCount()).Should(Equal(1))
		})

		It("should prepare zero events when resources don't exist", func() {
			fakeReader.GetCalls(
				func(ctx context.Context, name types.NamespacedName, object client.Object, opts ...client.GetOption) error {
					Expect(name).Should(Equal(types.NamespacedName{Name: gcName}))
					Expect(object).Should(BeAssignableToTypeOf(&v1beta1.GatewayClass{}))

					return apierrors.NewNotFound(schema.GroupResource{}, "test")
				},
			)
			fakeReader.ListReturns(nil)

			batch, err := preparer.Prepare(context.Background())

			Expect(batch).Should(BeEmpty())
			Expect(err).Should(BeNil())
		})

		It("should prepare one event for each resource type", func() {
			gatewayClass := v1beta1.GatewayClass{ObjectMeta: metav1.ObjectMeta{Name: gcName}}

			fakeReader.GetCalls(
				func(ctx context.Context, name types.NamespacedName, object client.Object, opts ...client.GetOption) error {
					Expect(name).Should(Equal(types.NamespacedName{Name: gcName}))
					Expect(object).Should(BeAssignableToTypeOf(&v1beta1.GatewayClass{}))

					reflect.Indirect(reflect.ValueOf(object)).Set(reflect.Indirect(reflect.ValueOf(&gatewayClass)))
					return nil
				},
			)

			httpRoute := v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "test"}}

			fakeReader.ListCalls(func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
				Expect(option).To(BeEmpty())

				switch typedList := list.(type) {
				case *v1beta1.HTTPRouteList:
					typedList.Items = append(typedList.Items, httpRoute)
				default:
					Fail(fmt.Sprintf("unknown type: %T", typedList))
				}

				return nil
			})

			expectedBatch := events.EventBatch{
				&events.UpsertEvent{Resource: &gatewayClass},
				&events.UpsertEvent{Resource: &httpRoute},
			}

			batch, err := preparer.Prepare(context.Background())

			Expect(batch).Should(Equal(expectedBatch))
			Expect(err).Should(BeNil())
		})
	})

	Describe("Edge cases", func() {
		Describe("EachListItem cases", func() {
			BeforeEach(func() {
				fakeReader.GetReturns(apierrors.NewNotFound(schema.GroupResource{}, "test"))
				fakeReader.ListCalls(
					func(ctx context.Context, list client.ObjectList, option ...client.ListOption) error {
						httpRoute := v1beta1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Name: "test"}}
						typedList := list.(*v1beta1.HTTPRouteList)
						typedList.Items = append(typedList.Items, httpRoute)

						return nil
					},
				)
			})

			It("should return error if EachListItem passes a wrong object type", func() {
				preparer.SetEachListItem(func(obj runtime.Object, fn func(runtime.Object) error) error {
					return fn(&fakeRuntimeObject{})
				})

				batch, err := preparer.Prepare(context.Background())
				Expect(batch).To(BeNil())
				Expect(err).To(MatchError("cannot cast *events_test.fakeRuntimeObject to client.Object"))
			})

			It("should return error if EachListItem returns an error", func() {
				testError := errors.New("test")

				preparer.SetEachListItem(func(obj runtime.Object, fn func(runtime.Object) error) error {
					return testError
				})

				batch, err := preparer.Prepare(context.Background())
				Expect(batch).To(BeNil())
				Expect(err).To(MatchError(testError))
			})
		})

		DescribeTable("Reader returns errors",
			func(obj client.Object) {
				readerError := errors.New("test")

				fakeReader.GetReturns(nil)
				fakeReader.ListReturns(nil)

				switch obj.(type) {
				case *v1beta1.GatewayClass:
					fakeReader.GetReturns(readerError)
				case *v1beta1.HTTPRoute:
					fakeReader.ListReturnsOnCall(0, readerError)
				default:
					Fail(fmt.Sprintf("Unknown type: %T", obj))
				}

				batch, err := preparer.Prepare(context.Background())
				Expect(batch).To(BeNil())
				Expect(err).To(MatchError(readerError))
			},
			Entry("GatewayClass", &v1beta1.GatewayClass{}),
			Entry("HTTPRoute", &v1beta1.HTTPRoute{}),
		)
	})
})

type fakeRuntimeObject struct{}

func (f *fakeRuntimeObject) GetObjectKind() schema.ObjectKind {
	return nil
}

func (f *fakeRuntimeObject) DeepCopyObject() runtime.Object {
	return nil
}
