package provisioner

import (
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
)

// store stores the cluster state needed by the provisioner and allows to update it from the events.
type store struct {
	gatewayClasses map[types.NamespacedName]*v1beta1.GatewayClass
	gateways       map[types.NamespacedName]*v1beta1.Gateway
}

func newStore() *store {
	return &store{
		gatewayClasses: make(map[types.NamespacedName]*v1beta1.GatewayClass),
		gateways:       make(map[types.NamespacedName]*v1beta1.Gateway),
	}
}

func (s *store) update(batch events.EventBatch) {
	for _, event := range batch {
		switch e := event.(type) {
		case *events.UpsertEvent:
			switch obj := e.Resource.(type) {
			case *v1beta1.GatewayClass:
				s.gatewayClasses[client.ObjectKeyFromObject(obj)] = obj
			case *v1beta1.Gateway:
				s.gateways[client.ObjectKeyFromObject(obj)] = obj
			default:
				panic(fmt.Errorf("unknown resource type %T", e.Resource))
			}
		case *events.DeleteEvent:
			switch e.Type.(type) {
			case *v1beta1.GatewayClass:
				delete(s.gatewayClasses, e.NamespacedName)
			case *v1beta1.Gateway:
				delete(s.gateways, e.NamespacedName)
			default:
				panic(fmt.Errorf("unknown resource type %T", e.Type))
			}
		default:
			panic(fmt.Errorf("unknown event type %T", e))
		}
	}
}
