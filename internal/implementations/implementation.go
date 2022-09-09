package implementations

import (
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

type ImplementationImpl struct {
	logger       logr.Logger
	objectType   client.Object
	eventCh      chan<- interface{}
	filter       NamespacedNameFilter
	resourceKind string
}

var _ sdk.Implementation = &ImplementationImpl{}

type NamespacedNameFilter func(nsname types.NamespacedName) (bool, string)

func NewImplementation(objectType client.Object, logger logr.Logger, eventCh chan<- interface{}) *ImplementationImpl {
	return NewImplementationWithFilter(objectType, logger, eventCh, nil)
}

func NewImplementationWithFilter(
	objectType client.Object,
	logger logr.Logger,
	eventCh chan<- interface{},
	filter NamespacedNameFilter,
) *ImplementationImpl {
	return &ImplementationImpl{
		objectType:   objectType,
		logger:       logger,
		eventCh:      eventCh,
		filter:       filter,
		resourceKind: reflect.TypeOf(objectType).Elem().Name(),
	}
}

func (impl *ImplementationImpl) Upsert(obj client.Object) {
	t := reflect.TypeOf(impl.objectType)
	if t != reflect.TypeOf(obj) {
		panic(fmt.Sprintf("Upsert called with wrong type. Expected %s, got %s", t, reflect.TypeOf(obj)))
	}

	nsname := types.NamespacedName{Namespace: obj.GetNamespace(), Name: obj.GetName()}

	if impl.filter != nil {
		if ignore, msg := impl.filter(nsname); ignore {
			impl.logger.Info(msg,
				"namespace", nsname.Namespace,
				"name", nsname.Name,
			)
			return
		}
	}

	impl.logger.Info(fmt.Sprintf("%s was upserted", impl.resourceKind),
		"namespace", obj.GetNamespace(),
		"name", obj.GetName(),
	)

	impl.eventCh <- &events.UpsertEvent{
		Resource: obj,
	}
}

func (impl *ImplementationImpl) Remove(nsname types.NamespacedName) {
	if impl.filter != nil {
		if ignore, msg := impl.filter(nsname); ignore {
			impl.logger.Info(msg,
				"namespace", nsname.Namespace,
				"name", nsname.Name,
			)
			return
		}
	}

	impl.logger.Info(fmt.Sprintf("%s was removed", impl.resourceKind),
		"namespace", nsname.Namespace,
		"name", nsname.Name,
	)

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           impl.objectType,
	}
}
