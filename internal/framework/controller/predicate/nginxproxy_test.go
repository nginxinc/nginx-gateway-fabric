package predicate

import (
	"testing"

	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

func TestNginxProxyPredicate(t *testing.T) {
	g := NewWithT(t)
	scheme := runtime.NewScheme()
	utilruntime.Must(v1beta1.AddToScheme(scheme))
	utilruntime.Must(ngfAPI.AddToScheme(scheme))

	gc := v1beta1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "gc",
		},
		Spec: v1beta1.GatewayClassSpec{
			ControllerName: "nginx-ctlr",
			ParametersRef: &v1beta1.ParametersReference{
				Group:     ngfAPI.GroupName,
				Kind:      v1beta1.Kind("NginxProxy"),
				Name:      "np",
				Namespace: helpers.GetPointer(v1beta1.Namespace("nginx-gateway")),
			},
		},
	}

	p := NginxProxyPredicate{
		Client:         fake.NewClientBuilder().WithScheme(scheme).WithObjects(&gc).Build(),
		ControllerName: "nginx-ctlr",
	}

	np := &ngfAPI.NginxProxy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "np",
			Namespace: "nginx-gateway",
		},
	}

	g.Expect(p.Create(event.CreateEvent{Object: np})).To(BeTrue())
	g.Expect(p.Update(event.UpdateEvent{ObjectNew: np})).To(BeTrue())
	g.Expect(p.Delete(event.DeleteEvent{Object: np})).To(BeTrue())

	// Unlink the two objects by name
	np.Name = "new-np"
	g.Expect(p.Create(event.CreateEvent{Object: np})).To(BeFalse())
	g.Expect(p.Update(event.UpdateEvent{ObjectNew: np})).To(BeFalse())
	g.Expect(p.Delete(event.DeleteEvent{Object: np})).To(BeFalse())

	// Unlink the two objects by controller name
	np.Name = "np"
	p.ControllerName = "some-other-ctlr"
	g.Expect(p.Create(event.CreateEvent{Object: np})).To(BeFalse())
	g.Expect(p.Update(event.UpdateEvent{ObjectNew: np})).To(BeFalse())
	g.Expect(p.Delete(event.DeleteEvent{Object: np})).To(BeFalse())
}
