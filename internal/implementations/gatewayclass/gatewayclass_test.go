package implementation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	implementation "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/gatewayclass"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

var _ = Describe("GatewayClassImplementation", func() {
	var (
		eventCh chan interface{}
		impl    sdk.GatewayClassImpl
	)

	const (
		className          = "my-class"
		unrelatedClassName = "not-my-class"
	)

	BeforeEach(func() {
		eventCh = make(chan interface{})

		impl = implementation.NewGatewayClassImplementation(config.Config{
			Logger:           zap.New(),
			GatewayClassName: className,
		}, eventCh)
	})

	Describe("Implementation processes GatewayClass", func() {
		It("should process upsert", func() {
			gc := &v1alpha2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: className,
				},
			}

			go func() {
				impl.Upsert(gc)
			}()

			Eventually(eventCh).Should(Receive(Equal(&events.UpsertEvent{Resource: gc})))
		})

		It("should process remove", func() {
			nsname := types.NamespacedName{Name: className}

			go func() {
				impl.Remove(nsname)
			}()

			Eventually(eventCh).Should(Receive(Equal(
				&events.DeleteEvent{
					NamespacedName: nsname,
					Type:           &v1alpha2.GatewayClass{},
				})))
		})
	})

	Describe("Implementation ignores unrelated GatewayClass", func() {
		It("should ignore upsert", func() {
			gc := &v1alpha2.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: unrelatedClassName,
				},
			}

			impl.Upsert(gc)

			Expect(eventCh).ShouldNot(Receive())
		})

		It("should ignore remove", func() {
			nsname := types.NamespacedName{Name: unrelatedClassName}

			impl.Remove(nsname)

			Expect(eventCh).ShouldNot(Receive())
		})
	})
})
