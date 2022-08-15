package implementation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/config"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/events"
	implementation "github.com/nginxinc/nginx-kubernetes-gateway/internal/implementations/endpointslice"
	"github.com/nginxinc/nginx-kubernetes-gateway/pkg/sdk"
)

var _ = Describe("EndpointSliceImplementation", func() {
	var (
		eventCh chan interface{}
		impl    sdk.EndpointSliceImpl
	)

	BeforeEach(func() {
		eventCh = make(chan interface{})

		impl = implementation.NewEndpointSliceImplementation(config.Config{
			Logger: zap.New(),
		}, eventCh)
	})

	const endpointSliceName = "my-endpoint-slice"
	const endpointSliceNamespace = "test"

	Describe("Implementation processes Endpoint Slices", func() {
		It("should process upsert", func() {
			endpointSlice := &discoveryV1.EndpointSlice{
				ObjectMeta: metav1.ObjectMeta{
					Name:      endpointSliceName,
					Namespace: endpointSliceNamespace,
				},
			}

			go func() {
				impl.Upsert(endpointSlice)
			}()

			Eventually(eventCh).Should(Receive(Equal(&events.UpsertEvent{Resource: endpointSlice})))
		})

		It("should process remove", func() {
			nsname := types.NamespacedName{Name: endpointSliceName, Namespace: endpointSliceNamespace}

			go func() {
				impl.Remove(nsname)
			}()

			Eventually(eventCh).Should(Receive(Equal(
				&events.DeleteEvent{
					NamespacedName: nsname,
					Type:           &discoveryV1.EndpointSlice{},
				})))
		})
	})
})
