package predicate

import (
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestGatewayClassPredicate(t *testing.T) {
	p := GatewayClassPredicate{ControllerName: "nginx-ctlr"}

	gc := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: "nginx-ctlr",
		},
	}

	if !p.Create(event.CreateEvent{Object: gc}) {
		t.Error("GatewayClassPredicate.Create() returned false; expected true")
	}
	if !p.Update(event.UpdateEvent{ObjectNew: gc}) {
		t.Error("GatewayClassPredicate.Update() returned false; expected true")
	}

	gc.Spec.ControllerName = "unknown"
	if p.Create(event.CreateEvent{Object: gc}) {
		t.Error("GatewayClassPredicate.Create() returned true; expected false")
	}
	if p.Update(event.UpdateEvent{ObjectNew: gc}) {
		t.Error("GatewayClassPredicate.Update() returned true; expected false")
	}
}
