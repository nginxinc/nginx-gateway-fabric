package manager

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	defer func() {
		newReconciler = reconciler.NewImplementation
	}()

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
		expectedMgrAddCallCount int
		msg                     string
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

	cfg := controllerConfig{
		objectType:           &v1beta1.HTTPRoute{},
		namespacedNameFilter: filter.CreateFilterForGatewayClass("test"),
		k8sPredicate:         predicate.ServicePortsChangedPredicate{},
		fieldIndexes: map[string]client.IndexerFunc{
			index.KubernetesServiceNameIndexField: index.ServiceNameIndexFunc,
		},
	}

	eventCh := make(chan interface{})

	for _, test := range tests {
		newReconciler = func(c reconciler.Config) *reconciler.Implementation {
			if c.Getter != test.fakes.mgr.GetClient() {
				t.Errorf("regiterController() created a reconciler config with Getter %p but expected %p for case of %q", c.Getter, test.fakes.mgr.GetClient(), test.msg)
			}
			if c.ObjectType != cfg.objectType {
				t.Errorf("registerController() created a reconciler config with ObjectType %T but expected %T for case of %q", c.ObjectType, cfg.objectType, test.msg)
			}
			if c.EventCh != eventCh {
				t.Errorf("registerController() created a reconciler config with EventCh %v but expected %v for case of %q", c.EventCh, eventCh, test.msg)
			}
			// comparing functions is not allowed in Go, so we're comparing the pointers
			if reflect.ValueOf(c.NamespacedNameFilter).Pointer() != reflect.ValueOf(cfg.namespacedNameFilter).Pointer() {
				t.Errorf("registerController() created a reconciler config with NamespacedNameFilter %p but expected %p for case of %q", c.NamespacedNameFilter, cfg.namespacedNameFilter, test.msg)
			}

			return reconciler.NewImplementation(c)
		}

		err := registerController(context.Background(), test.fakes.mgr, eventCh, cfg)

		if !errors.Is(err, test.expectedErr) {
			t.Errorf("registerController() returned %q but expected %q for case of %q", err, test.expectedErr, test.msg)
		}

		indexCallCount := test.fakes.indexer.IndexFieldCallCount()
		if indexCallCount != 1 {
			t.Errorf("registerController() called indexer.IndexField() %d times but expected 1 for case of %q", indexCallCount, test.msg)
		} else {
			_, objType, field, indexFunc := test.fakes.indexer.IndexFieldArgsForCall(0)

			if objType != cfg.objectType {
				t.Errorf("registerController() called indexer.IndexField() with object type %T but expected %T for case of %q", objType, cfg.objectType, test.msg)
			}
			if field != index.KubernetesServiceNameIndexField {
				t.Errorf("registerController() called indexer.IndexField() with field %q but expected %q for case of %q", field, index.KubernetesServiceNameIndexField, test.msg)
			}
			// comparing functions is not allowed in Go, so we're comparing the pointers
			if reflect.ValueOf(indexFunc).Pointer() != reflect.ValueOf(index.ServiceNameIndexFunc).Pointer() {
				t.Errorf("registerController() called indexer.IndexField() with indexFunc %p but expected %p for case of %q", indexFunc, index.ServiceNameIndexFunc, test.msg)
			}
		}

		addCallCount := test.fakes.mgr.AddCallCount()
		if addCallCount != test.expectedMgrAddCallCount {
			t.Errorf("registerController() called mgr.Add() %d times but expected %d times for case of %q", addCallCount, test.expectedMgrAddCallCount, test.msg)
		}
	}
}
