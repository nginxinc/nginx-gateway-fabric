package reconciler_test

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler/reconcilerfakes"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

type getFunc func(ctx context.Context, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error

type result struct {
	err             error
	reconcileResult reconcile.Result
}

var _ = Describe("Reconciler", func() {
	var (
		rec        *reconciler.Implementation
		fakeGetter *reconcilerfakes.FakeGetter
		eventCh    chan interface{}

		hr1NsName = types.NamespacedName{
			Namespace: "test",
			Name:      "hr-1",
		}

		hr1 = &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: hr1NsName.Namespace,
				Name:      hr1NsName.Name,
			},
		}

		hr2NsName = types.NamespacedName{
			Namespace: "test",
			Name:      "hr-2",
		}

		hr2 = &v1beta1.HTTPRoute{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: hr2NsName.Namespace,
				Name:      hr2NsName.Name,
			},
		}
	)

	getReturnsHRForHR := func(hr *v1beta1.HTTPRoute) getFunc {
		return func(
			ctx context.Context,
			nsname types.NamespacedName,
			object client.Object,
			option ...client.GetOption,
		) error {
			Expect(object).To(BeAssignableToTypeOf(&v1beta1.HTTPRoute{}))
			Expect(nsname).To(Equal(client.ObjectKeyFromObject(hr)))

			hr.DeepCopyInto(object.(*v1beta1.HTTPRoute))

			return nil
		}
	}

	getReturnsNotFoundErrorForHR := func(hr *v1beta1.HTTPRoute) getFunc {
		return func(
			ctx context.Context,
			nsname types.NamespacedName,
			object client.Object,
			option ...client.GetOption,
		) error {
			Expect(object).To(BeAssignableToTypeOf(&v1beta1.HTTPRoute{}))
			Expect(nsname).To(Equal(client.ObjectKeyFromObject(hr)))

			return apierrors.NewNotFound(schema.GroupResource{}, "not found")
		}
	}

	startReconcilingWithContext := func(ctx context.Context, nsname types.NamespacedName) <-chan result {
		resultCh := make(chan result)

		go func() {
			defer GinkgoRecover()

			res, err := rec.Reconcile(ctx, reconcile.Request{NamespacedName: nsname})
			resultCh <- result{err: err, reconcileResult: res}

			close(resultCh)
		}()

		return resultCh
	}

	startReconciling := func(nsname types.NamespacedName) <-chan result {
		return startReconcilingWithContext(context.Background(), nsname)
	}

	BeforeEach(func() {
		fakeGetter = &reconcilerfakes.FakeGetter{}
		eventCh = make(chan interface{})
	})

	Describe("Normal cases", func() {
		When("Reconciler doesn't have a filter", func() {
			BeforeEach(func() {
				rec = reconciler.NewImplementation(reconciler.Config{
					Getter:     fakeGetter,
					ObjectType: &v1beta1.HTTPRoute{},
					EventCh:    eventCh,
				})
			})

			It("should upsert HTTPRoute", func() {
				fakeGetter.GetCalls(getReturnsHRForHR(hr1))

				resultCh := startReconciling(hr1NsName)

				Eventually(eventCh).Should(Receive(Equal(&events.UpsertEvent{Resource: hr1})))
				Eventually(resultCh).Should(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
			})

			It("should delete HTTPRoute", func() {
				fakeGetter.GetCalls(getReturnsNotFoundErrorForHR(hr1))

				resultCh := startReconciling(hr1NsName)

				Eventually(eventCh).Should(Receive(Equal(&events.DeleteEvent{
					NamespacedName: hr1NsName,
					Type:           &v1beta1.HTTPRoute{},
				})))
				Eventually(resultCh).Should(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
			})
		})

		When("Reconciler has a NamespacedNameFilter", func() {
			BeforeEach(func() {
				filter := func(nsname types.NamespacedName) (bool, string) {
					if nsname != hr1NsName {
						return false, "ignore"
					}
					return true, ""
				}

				rec = reconciler.NewImplementation(reconciler.Config{
					Getter:               fakeGetter,
					ObjectType:           &v1beta1.HTTPRoute{},
					EventCh:              eventCh,
					NamespacedNameFilter: filter,
				})
			})

			When("HTTPRoute is not ignored", func() {
				It("should upsert HTTPRoute", func() {
					fakeGetter.GetCalls(getReturnsHRForHR(hr1))

					resultCh := startReconciling(hr1NsName)

					Eventually(eventCh).Should(Receive(Equal(&events.UpsertEvent{Resource: hr1})))
					Eventually(resultCh).Should(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
				})

				It("should delete HTTPRoute", func() {
					fakeGetter.GetCalls(getReturnsNotFoundErrorForHR(hr1))

					resultCh := startReconciling(hr1NsName)

					Eventually(eventCh).Should(Receive(Equal(&events.DeleteEvent{
						NamespacedName: hr1NsName,
						Type:           &v1beta1.HTTPRoute{},
					})))
					Eventually(resultCh).Should(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
				})
			})

			When("HTTPRoute is ignored", func() {
				It("should not upsert HTTPRoute", func() {
					fakeGetter.GetCalls(getReturnsHRForHR(hr2))

					resultCh := startReconciling(hr2NsName)

					Consistently(eventCh).ShouldNot(Receive())
					Eventually(resultCh).Should(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
				})

				It("should not delete HTTPRoute", func() {
					fakeGetter.GetCalls(getReturnsNotFoundErrorForHR(hr2))

					resultCh := startReconciling(hr2NsName)

					Consistently(eventCh).ShouldNot(Receive())
					Eventually(resultCh).Should(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
				})
			})
		})
	})

	Describe("Edge cases", func() {
		BeforeEach(func() {
			rec = reconciler.NewImplementation(reconciler.Config{
				Getter:     fakeGetter,
				ObjectType: &v1beta1.HTTPRoute{},
				EventCh:    eventCh,
			})
		})

		It("should not reconcile when Getter returns error", func() {
			getError := errors.New("get error")
			fakeGetter.GetReturns(getError)

			resultCh := startReconciling(hr1NsName)

			Consistently(eventCh).ShouldNot(Receive())
			Eventually(resultCh).Should(Receive(Equal(result{err: getError, reconcileResult: reconcile.Result{}})))
		})

		DescribeTable("Reconciler should not block when ctx is done",
			func(get getFunc, nsname types.NamespacedName) {
				fakeGetter.GetCalls(get)

				ctx, cancel := context.WithCancel(context.Background())
				cancel()

				resultCh := startReconcilingWithContext(ctx, nsname)

				Consistently(eventCh).ShouldNot(Receive())
				Expect(resultCh).To(Receive(Equal(result{err: nil, reconcileResult: reconcile.Result{}})))
			},
			Entry("Upserting HTTPRoute", getReturnsHRForHR(hr1), hr1NsName),
			Entry("Deleting HTTPRoute", getReturnsNotFoundErrorForHR(hr1), hr1NsName),
		)
	})
})
