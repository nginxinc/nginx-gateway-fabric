package predicate

import (
	"testing"

	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/event"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGatewayClassPredicate(t *testing.T) {
	t.Parallel()
	g := NewWithT(t)

	p := GatewayClassPredicate{ControllerName: "nginx-ctlr"}

	gc := &v1.GatewayClass{
		Spec: v1.GatewayClassSpec{
			ControllerName: "nginx-ctlr",
		},
	}

	g.Expect(p.Create(event.CreateEvent{Object: gc})).To(BeTrue())
	g.Expect(p.Update(event.UpdateEvent{ObjectNew: gc})).To(BeTrue())
	g.Expect(p.Delete(event.DeleteEvent{Object: gc})).To(BeTrue())

	gc2 := &v1.GatewayClass{
		Spec: v1.GatewayClassSpec{
			ControllerName: "unknown",
		},
	}
	g.Expect(p.Create(event.CreateEvent{Object: gc2})).To(BeFalse())
	g.Expect(p.Update(event.UpdateEvent{ObjectOld: gc, ObjectNew: gc2})).To(BeTrue())
	g.Expect(p.Update(event.UpdateEvent{ObjectOld: gc2, ObjectNew: gc})).To(BeTrue())
	g.Expect(p.Update(event.UpdateEvent{ObjectOld: gc2, ObjectNew: gc2})).To(BeFalse())
	g.Expect(p.Delete(event.DeleteEvent{Object: nil})).To(BeFalse())
	g.Expect(p.Delete(event.DeleteEvent{Object: gc2})).To(BeFalse())
	g.Expect(p.Delete(event.DeleteEvent{Object: &v1.HTTPRoute{}})).To(BeFalse())
}
