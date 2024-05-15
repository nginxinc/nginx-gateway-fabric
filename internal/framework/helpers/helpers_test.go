package helpers

import (
	"testing"

	. "github.com/onsi/gomega"

	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
)

func TestMustCastObject(t *testing.T) {
	g := NewWithT(t)

	var obj client.Object = &gatewayv1.Gateway{}

	g.Expect(func() {
		_ = MustCastObject[*gatewayv1.Gateway](obj)
	}).ToNot(Panic())

	g.Expect(func() {
		_ = MustCastObject[*gatewayv1alpha3.BackendTLSPolicy](obj)
	}).To(Panic())
}
