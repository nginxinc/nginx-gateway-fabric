package events

import (
	"context"
	"fmt"

	apiv1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . FirstEventBatchPreparer

// FirstEventBatchPreparer prepares the first batch of events to be processed by the EventHandler.
// The first batch includes the UpsertEvents for all relevant resources in the cluster.
type FirstEventBatchPreparer interface {
	// Prepare prepares the first event batch.
	Prepare(ctx context.Context) (EventBatch, error)
}

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . CachedReader

// CachedReader allows getting and listing resources from a cache.
// This interface is introduced for testing to mock a subset of methods from
// sigs.k8s.io/controller-runtime/pkg/cache.Cache.
type CachedReader interface {
	WaitForCacheSync(ctx context.Context) bool
	Get(ctx context.Context, key client.ObjectKey, obj client.Object) error
	List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error
}

// FirstEventBatchPreparerImpl is an implementation of FirstEventBatchPreparer.
type FirstEventBatchPreparerImpl struct {
	reader CachedReader
	gcName string
}

// NewFirstEventBatchPreparerImpl creates a new FirstEventBatchPreparerImpl.
func NewFirstEventBatchPreparerImpl(reader CachedReader, gcName string) *FirstEventBatchPreparerImpl {
	return &FirstEventBatchPreparerImpl{
		reader: reader,
		gcName: gcName,
	}
}

func (p *FirstEventBatchPreparerImpl) Prepare(ctx context.Context) (EventBatch, error) {
	synced := p.reader.WaitForCacheSync(ctx)
	if !synced {
		return nil, fmt.Errorf("cache is not synced")
	}

	var gc v1beta1.GatewayClass
	gcExist := true
	err := p.reader.Get(ctx, types.NamespacedName{Name: p.gcName}, &gc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			gcExist = false
		} else {
			return nil, err
		}
	}

	var svcList apiv1.ServiceList
	var secretList apiv1.SecretList
	var gwList v1beta1.GatewayList
	var hrList v1beta1.HTTPRouteList

	objLists := []client.ObjectList{&svcList, &secretList, &gwList, &hrList}
	for _, list := range objLists {
		err := p.reader.List(ctx, list)
		if err != nil {
			return nil, err
		}
	}

	gcCount := 0
	if gcExist {
		gcCount = 1
	}

	batch := make([]interface{}, 0, gcCount+len(svcList.Items)+len(secretList.Items)+len(gwList.Items)+len(hrList.Items))

	// Note: the order of the events doesn't matter.

	if gcExist {
		batch = append(batch, &UpsertEvent{Resource: &gc})
	}

	for i := range svcList.Items {
		batch = append(batch, &UpsertEvent{Resource: &svcList.Items[i]})
	}
	for i := range secretList.Items {
		batch = append(batch, &UpsertEvent{Resource: &secretList.Items[i]})
	}
	for i := range gwList.Items {
		batch = append(batch, &UpsertEvent{Resource: &gwList.Items[i]})
	}
	for i := range hrList.Items {
		batch = append(batch, &UpsertEvent{Resource: &hrList.Items[i]})
	}

	return batch, nil
}
