package predicate

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestGatewayClassPredicate(t *testing.T) {
	g := NewGomegaWithT(t)

	p := GatewayClassPredicate{ControllerName: "nginx-ctlr"}

	gc := &v1beta1.GatewayClass{
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: "nginx-ctlr",
		},
	}

	g.Expect(p.Create(event.CreateEvent{Object: gc})).To(BeTrue())
	g.Expect(p.Update(event.UpdateEvent{ObjectNew: gc})).To(BeTrue())

	gc.Spec.ControllerName = "unknown"
	g.Expect(p.Create(event.CreateEvent{Object: gc})).To(BeFalse())
	g.Expect(p.Update(event.UpdateEvent{ObjectNew: gc})).To(BeFalse())
}
