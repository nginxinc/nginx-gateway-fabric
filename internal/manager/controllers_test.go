package manager

import (
	"context"
	"errors"
	"reflect"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gcustom"
	"github.com/onsi/gomega/types"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/filter"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/managerfakes"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/predicate"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/reconciler"
)

func TestRegisterController(t *testing.T) {
	type fakes struct {
		mgr     *managerfakes.FakeManager
		indexer *managerfakes.FakeFieldIndexer
	}

	getDefaultFakes := func() fakes {
		scheme = runtime.NewScheme()
		utilruntime.Must(v1beta1.AddToScheme(scheme))

		indexer := &managerfakes.FakeFieldIndexer{}

		mgr := &managerfakes.FakeManager{}
		mgr.GetClientReturns(fake.NewClientBuilder().Build())
		mgr.GetSchemeReturns(scheme)
		mgr.GetLoggerReturns(zap.New())
		mgr.GetFieldIndexerReturns(indexer)

		return fakes{
			mgr:     mgr,
			indexer: indexer,
		}
	}

	testError := errors.New("test error")

	tests := []struct {
		fakes                   fakes
		expectedErr             error
		msg                     string
		expectedMgrAddCallCount int
	}{
		{
			fakes:                   getDefaultFakes(),
			expectedErr:             nil,
			expectedMgrAddCallCount: 1,
			msg:                     "normal case",
		},
		{
			fakes: func(f fakes) fakes {
				f.indexer.IndexFieldReturns(testError)
				return f
			}(getDefaultFakes()),
			expectedErr:             testError,
			expectedMgrAddCallCount: 0,
			msg:                     "preparing index fails",
		},
		{
			fakes: func(f fakes) fakes {
				f.mgr.AddReturns(testError)
				return f
			}(getDefaultFakes()),
			expectedErr:             testError,
			expectedMgrAddCallCount: 1,
			msg:                     "building controller fails",
		},
	}

	objectType := &v1beta1.HTTPRoute{}
	namespacedNameFilter := filter.CreateFilterForGatewayClass("test")
	fieldIndexes := index.CreateEndpointSliceFieldIndices()

	eventCh := make(chan<- interface{})

	beSameFunctionPointer := func(expected interface{}) types.GomegaMatcher {
		return gcustom.MakeMatcher(func(f interface{}) (bool, error) {
			// comparing functions is not allowed in Go, so we're comparing the pointers
			return reflect.ValueOf(expected).Pointer() == reflect.ValueOf(f).Pointer(), nil
		})
	}

	for _, test := range tests {
		t.Run(test.msg, func(t *testing.T) {
			g := NewGomegaWithT(t)

			newReconciler := func(c reconciler.Config) *reconciler.Implementation {
				g.Expect(c.Getter).To(BeIdenticalTo(test.fakes.mgr.GetClient()))
				g.Expect(c.ObjectType).To(BeIdenticalTo(objectType))
				g.Expect(c.EventCh).To(BeIdenticalTo(eventCh))
				g.Expect(c.NamespacedNameFilter).Should(beSameFunctionPointer(namespacedNameFilter))

				return reconciler.NewImplementation(c)
			}

			err := registerController(
				context.Background(),
				objectType,
				test.fakes.mgr,
				eventCh,
				withNamespacedNameFilter(namespacedNameFilter),
				withK8sPredicate(predicate.ServicePortsChangedPredicate{}),
				withFieldIndices(fieldIndexes),
				withNewReconciler(newReconciler),
			)

			if test.expectedErr == nil {
				g.Expect(err).To(BeNil())
			} else {
				g.Expect(err).To(MatchError(test.expectedErr))
			}

			indexCallCount := test.fakes.indexer.IndexFieldCallCount()

			g.Expect(indexCallCount).To(Equal(1))

			_, objType, field, indexFunc := test.fakes.indexer.IndexFieldArgsForCall(0)

			g.Expect(objType).To(BeIdenticalTo(objectType))
			g.Expect(field).To(BeIdenticalTo(index.KubernetesServiceNameIndexField))

			expectedIndexFunc := fieldIndexes[index.KubernetesServiceNameIndexField]
			g.Expect(indexFunc).To(beSameFunctionPointer(expectedIndexFunc))

			addCallCount := test.fakes.mgr.AddCallCount()
			g.Expect(addCallCount).To(Equal(test.expectedMgrAddCallCount))
		})
	}
}
