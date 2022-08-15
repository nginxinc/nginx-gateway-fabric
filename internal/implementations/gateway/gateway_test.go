package implementation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	implementation "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gateway"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

var _ = Describe("GatewayImplementation", func() {
	var (
		eventCh chan interface{}
		impl    sdk.GatewayImpl
	)

	BeforeEach(func() {
		eventCh = make(chan interface{})

		impl = implementation.NewGatewayImplementation(config.Config{
			Logger: zap.New(),
		}, eventCh)
	})

	Describe("Implementation processes Gateways", func() {
		It("should process upsert", func() {
			gc := &v1beta1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test-add",
					Name:      "gateway",
				},
			}

			go func() {
				impl.Upsert(gc)
			}()

			Eventually(eventCh).Should(Receive(Equal(&events.UpsertEvent{Resource: gc})))
		})

		It("should process remove", func() {
			nsname := types.NamespacedName{Namespace: "test-remove", Name: "gateway"}

			go func() {
				impl.Remove(nsname)
			}()

			Eventually(eventCh).Should(Receive(Equal(
				&events.DeleteEvent{
					NamespacedName: nsname,
					Type:           &v1beta1.Gateway{},
				})))
		})
	})
})
