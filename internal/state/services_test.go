package state

import (
	"context"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	apiv1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
)

func createEndpointSlice(name, serviceName string, addresses []string, ports []int32, addressType discoveryV1.AddressType) *discoveryV1.EndpointSlice {
	es := &discoveryV1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "test",
			Labels: map[string]string{
				"kubernetes.io/service-name": serviceName,
			},
		},
		AddressType: addressType,
		Endpoints: []discoveryV1.Endpoint{
			{
				Addresses: addresses,
				Conditions: discoveryV1.EndpointConditions{
					Ready: helpers.GetBoolPointer(true),
				},
			},
			{
				Addresses: []string{"1.0.0.1", "1.0.0.2", "1.0.0.3"},
				Conditions: discoveryV1.EndpointConditions{
					Serving:     helpers.GetBoolPointer(true),
					Terminating: helpers.GetBoolPointer(true),
				},
			},
			{
				Addresses:  []string{"2.0.0.1", "2.0.0.2", "2.0.0.3"},
				Conditions: discoveryV1.EndpointConditions{
					// nil conditions should be treated as not ready
				},
			},
		},
		Ports: make([]discoveryV1.EndpointPort, len(ports)),
	}

	for i, p := range ports {
		port := p
		es.Ports[i] = discoveryV1.EndpointPort{Port: &port}
	}

	return es
}

func createService(name string) *apiv1.Service {
	return &apiv1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      name,
		},
		Spec: apiv1.ServiceSpec{
			Ports: []apiv1.ServicePort{
				{
					Port: 80,
					TargetPort: intstr.IntOrString{
						Type:   intstr.Int,
						IntVal: 8080,
					},
				},
				{
					Port: 443,
					TargetPort: intstr.IntOrString{
						Type:   intstr.String,
						StrVal: "8443",
					},
				},
			},
		},
	}
}

func createExpectedEndpoints(addresses []string, port int32) []Endpoint {
	expEndpoints := make([]Endpoint, len(addresses))
	for idx, addr := range addresses {
		expEndpoints[idx] = Endpoint{Address: addr, Port: port}
	}

	return expEndpoints
}

var _ = Describe("ServiceStore", func() {
	var (
		store         ServiceStore
		fakeK8sClient client.Client
	)

	var (
		fooAddresses1    = []string{"9.0.0.1", "9.0.0.2", "9.0.0.3"}
		fooAddresses2    = []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}
		fooIPV6Addresses = []string{"FE80:CD00:0:CDE:1257:0:211E:729C"}

		barAddresses     = []string{"13.0.0.1", "13.0.0.2"}
		bar8081Addresses = []string{"12.0.0.1", "12.0.0.2"}

		ports    = []int32{8080, 8443}
		port8081 = []int32{8081}

		fooSvc        = createService("foo")
		barSvc        = createService("bar")
		noEndpointSvc = createService("no-endpoints")

		// foo endpoints
		fooEndpointSlice1    = createEndpointSlice("foo-1", "foo", fooAddresses1, ports, discoveryV1.AddressTypeIPv4)
		fooEndpointSlice2    = createEndpointSlice("foo-2", "foo", fooAddresses2, ports, discoveryV1.AddressTypeIPv4)
		fooEndpointSliceIPV6 = createEndpointSlice("foo-ipv6", "foo", fooIPV6Addresses, port8081, discoveryV1.AddressTypeIPv6)

		// bar endpoints
		barEndpointSlice1        = createEndpointSlice("bar", "bar", barAddresses, ports, discoveryV1.AddressTypeIPv4)
		fooEndpointSlicePort8081 = createEndpointSlice("bar-diff-ports", "bar", bar8081Addresses, port8081, discoveryV1.AddressTypeIPv4)

		fooExpectedAddresses = append(fooAddresses1, fooAddresses2...)
	)

	Describe("Resolve", Ordered, func() {
		BeforeAll(func() {
			scheme := runtime.NewScheme()
			err := discoveryV1.AddToScheme(scheme)
			Expect(err).ToNot(HaveOccurred())

			fakeK8sClient = fake.NewClientBuilder().
				WithScheme(scheme).
				WithRuntimeObjects(fooEndpointSlice1, fooEndpointSlice2, fooEndpointSlicePort8081, fooEndpointSliceIPV6, barEndpointSlice1).
				Build()

			store = NewServiceStore(fakeK8sClient)
		})

		testResolve := func(svcName string, svcPort int32, expEndpoints []Endpoint) {
			endpoints, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: svcName}, svcPort)
			Expect(err).To(BeNil())

			Expect(endpoints).To(ConsistOf(expEndpoints))
		}

		It("should add a service", func() {
			store.Upsert(fooSvc)
		})

		It("should resolve the service", func() {
			testResolve("foo", 80, createExpectedEndpoints(fooExpectedAddresses, 8080))
			testResolve("foo", 443, createExpectedEndpoints(fooExpectedAddresses, 8443))
		})

		It("should add a new service", func() {
			store.Upsert(barSvc)
		})

		It("should resolve the service", func() {
			testResolve("bar", 80, createExpectedEndpoints(barAddresses, 8080))
			testResolve("bar", 443, createExpectedEndpoints(barAddresses, 8443))
		})
		When("port does not exist in service spec", func() {
			It("should return an error", func() {
				endpoints, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "bar"}, 8080)
				Expect(err).ToNot(BeNil())
				Expect(endpoints).To(BeNil())
			})
		})

		It("should update the service", func() {
			barSvcUpdated := barSvc.DeepCopy()
			barSvcUpdated.Spec.Ports[0].TargetPort.IntVal = 8081

			store.Upsert(barSvcUpdated)
		})
		It("should resolve the updated service", func() {
			testResolve("bar", 80, createExpectedEndpoints(bar8081Addresses, 8081))
		})

		It("should delete the service", func() {
			store.Delete(types.NamespacedName{Namespace: "test", Name: "bar"})
		})

		It("should fail to resolve the service", func() {
			endpoints, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "bar"}, 80)

			Expect(err).To(HaveOccurred())
			Expect(endpoints).To(BeNil())
		})
		It("should still resolve remaining service", func() {
			testResolve("foo", 80, createExpectedEndpoints(fooExpectedAddresses, 8080))
			testResolve("foo", 443, createExpectedEndpoints(fooExpectedAddresses, 8443))
		})
		When("resolving a service has no valid endpoints", func() {
			It("should return an error", func() {
				// Delete all valid foo endpoints
				Expect(fakeK8sClient.Delete(context.TODO(), fooEndpointSlice1)).To(Succeed())
				Expect(fakeK8sClient.Delete(context.TODO(), fooEndpointSlice2)).To(Succeed())

				endpoints, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "foo"}, 80)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no valid endpoints found"))
				Expect(endpoints).To(BeNil())
			})
		})
		It("should add a service with no endpoints", func() {
			store.Upsert(noEndpointSvc)
		})
		When("resolving a service with no endpoints", func() {
			It("should return an error", func() {
				endpoints, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "no-endpoints"}, 80)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no endpoints found"))
				Expect(endpoints).To(BeNil())
			})
		})
		When("resolving a service with no ready endpoints", func() {
			It("should return an error", func() {
				endpoints, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "no-endpoints"}, 80)

				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("no endpoints found"))
				Expect(endpoints).To(BeNil())
			})
		})
	})
})

func TestCalculateEndpointSliceCapacity(t *testing.T) {

	addresses := []string{"10.0.0.1", "10.0.0.2", "10.0.0.3"}

	readyEndpoint1 := discoveryV1.Endpoint{
		Addresses:  addresses,
		Conditions: discoveryV1.EndpointConditions{Ready: helpers.GetBoolPointer(true)},
	}

	notReadyEndpoint := discoveryV1.Endpoint{
		Addresses:  addresses,
		Conditions: discoveryV1.EndpointConditions{Ready: helpers.GetBoolPointer(false)},
	}

	validEndpointSlice := discoveryV1.EndpointSlice{
		AddressType: discoveryV1.AddressTypeIPv4,
		Endpoints:   []discoveryV1.Endpoint{readyEndpoint1, readyEndpoint1, readyEndpoint1}, // in reality these endpoints would be different but for this test it doesn't matter
		Ports: []discoveryV1.EndpointPort{
			{
				Port: helpers.GetInt32Pointer(80),
			},
			{
				Port: helpers.GetInt32Pointer(443),
			},
		},
	}

	invalidAddressTypeEndpointSlice := discoveryV1.EndpointSlice{
		AddressType: discoveryV1.AddressTypeIPv6,
		Endpoints:   []discoveryV1.Endpoint{readyEndpoint1},
		Ports: []discoveryV1.EndpointPort{
			{
				Port: helpers.GetInt32Pointer(80),
			},
		},
	}

	invalidPortEndpointSlice := discoveryV1.EndpointSlice{
		AddressType: discoveryV1.AddressTypeIPv4,
		Endpoints:   []discoveryV1.Endpoint{readyEndpoint1},
		Ports: []discoveryV1.EndpointPort{
			{
				Port: helpers.GetInt32Pointer(8080),
			},
		},
	}

	notReadyEndpointSlice := discoveryV1.EndpointSlice{
		AddressType: discoveryV1.AddressTypeIPv4,
		Endpoints:   []discoveryV1.Endpoint{notReadyEndpoint, notReadyEndpoint}, // in reality these endpoints would be different but for this test it doesn't matter
		Ports: []discoveryV1.EndpointPort{
			{
				Port: helpers.GetInt32Pointer(80),
			},
			{
				Port: helpers.GetInt32Pointer(443),
			},
		},
	}

	mixedValidityEndpointSlice := discoveryV1.EndpointSlice{
		AddressType: discoveryV1.AddressTypeIPv4,
		Endpoints:   []discoveryV1.Endpoint{readyEndpoint1, notReadyEndpoint, readyEndpoint1}, // 6 valid endpoints
		Ports: []discoveryV1.EndpointPort{
			{
				Port: helpers.GetInt32Pointer(80),
			},
		},
	}

	testcases := []struct {
		msg            string
		endpointSlices []discoveryV1.EndpointSlice
		targetPort     int32
		expCapacity    int
	}{
		{
			msg: "multiple endpoint slices - multiple valid endpoints",
			endpointSlices: []discoveryV1.EndpointSlice{
				validEndpointSlice,
				validEndpointSlice}, // in reality these endpoints would be different but for this test it doesn't matter
			targetPort:  80,
			expCapacity: 18,
		},
		{
			msg: "multiple endpoint slices - some valid ",
			endpointSlices: []discoveryV1.EndpointSlice{
				validEndpointSlice,
				invalidAddressTypeEndpointSlice,
				validEndpointSlice,
				invalidPortEndpointSlice,
			},
			targetPort:  80,
			expCapacity: 18,
		},
		{
			msg:            "multiple endpoints - some valid ",
			endpointSlices: []discoveryV1.EndpointSlice{mixedValidityEndpointSlice},
			targetPort:     80,
			expCapacity:    6,
		},
		{
			msg:            "multiple endpoint slices - all invalid ",
			endpointSlices: []discoveryV1.EndpointSlice{invalidAddressTypeEndpointSlice, invalidPortEndpointSlice},
			targetPort:     80,
			expCapacity:    0,
		},
		{
			msg:            "multiple endpoints - all invalid ",
			endpointSlices: []discoveryV1.EndpointSlice{notReadyEndpointSlice},
			targetPort:     80,
			expCapacity:    0,
		},
	}

	for _, tc := range testcases {
		capacity := calculateEndpointSliceCapacity(tc.endpointSlices, tc.targetPort)
		if capacity != tc.expCapacity {
			t.Errorf("calculateEndpointSliceCapacity() returned %d but expected %d for test: %q", capacity, tc.expCapacity, tc.msg)
		}
	}
}
