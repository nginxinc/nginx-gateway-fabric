package implementations

import (
	"fmt"
	"reflect"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

type ImplementationImpl[T sdk.ObjectConstraint] struct {
	logger       logr.Logger
	eventCh      chan<- interface{}
	filter       NamespacedNameFilter
	resourceKind string
}

var _ sdk.Implementation[*v1beta1.Gateway] = &ImplementationImpl[*v1beta1.Gateway]{}

type NamespacedNameFilter func(nsname types.NamespacedName) (bool, string)

func NewImplementation[T sdk.ObjectConstraint](logger logr.Logger, eventCh chan<- interface{}) *ImplementationImpl[T] {
	return NewImplementationWithFilter[T](logger, eventCh, nil)
}

func NewImplementationWithFilter[T sdk.ObjectConstraint](
	logger logr.Logger,
	eventCh chan<- interface{},
	filter NamespacedNameFilter,
) *ImplementationImpl[T] {
	var obj T
	return &ImplementationImpl[T]{
		logger:       logger,
		eventCh:      eventCh,
		filter:       filter,
		resourceKind: reflect.TypeOf(obj).Elem().Name(),
	}
}

func (impl *ImplementationImpl[T]) Upsert(obj T) {
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

func (impl *ImplementationImpl[T]) Remove(nsname types.NamespacedName) {
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

	var t T

	impl.eventCh <- &events.DeleteEvent{
		NamespacedName: nsname,
		Type:           t,
	}
}
