package state

import (
	. "github.com/onsi/ginkgo/v2"
	apiv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	. "github.com/onsi/gomega"
)

var _ = Describe("ServiceStore", func() {
	var store ServiceStore

	BeforeEach(OncePerOrdered, func() {
		store = NewServiceStore()
	})

	Describe("Resolve Service", Ordered, func() {
		var svc *apiv1.Service
		var svcUpdated *apiv1.Service

		BeforeAll(func() {
			svc = &apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "service1",
				},
				Spec: apiv1.ServiceSpec{
					ClusterIP: "10.0.0.1",
				},
			}

			svcUpdated = svc.DeepCopy()
			// In reality, ClusterIP cannot change for a regular service.
			// However, it is ok to use this change to test the Upsert function.
			svcUpdated.Spec.ClusterIP = "10.0.0.2"
		})

		It("should add a service", func() {
			store.Upsert(svc)
		})

		It("should resolve the service", func() {
			address, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "service1"})

			Expect(address).To(Equal("10.0.0.1"))
			Expect(err).To(BeNil())
		})

		It("should update the service", func() {
			store.Upsert(svcUpdated)
		})

		It("should resolve the updated service", func() {
			address, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "service1"})

			Expect(address).To(Equal("10.0.0.2"))
			Expect(err).To(BeNil())

		})

		It("should delete the service", func() {
			store.Delete(types.NamespacedName{Namespace: "test", Name: "service1"})
		})

		It("should fail to resolve the service", func() {
			_, err := store.Resolve(types.NamespacedName{Namespace: "test", Name: "service1"})

			Expect(err).To(HaveOccurred())
		})
	})

	Describe("Edge cases", func() {
		BeforeEach(func() {
			store.Upsert(&apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "empty-ip",
				},
			})

			store.Upsert(&apiv1.Service{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "none-ip",
				},
				Spec: apiv1.ServiceSpec{
					ClusterIP: "None",
				},
			})
		})
		DescribeTable("Resolve returns error",
			func(nsname types.NamespacedName) {
				_, err := store.Resolve(nsname)

				Expect(err).To(HaveOccurred())
			},
			Entry("cluster ip is empty", types.NamespacedName{Namespace: "test", Name: "empty-ip"}),
			Entry("cluster ip is none", types.NamespacedName{Namespace: "test", Name: "none-ip"}),
			Entry("service doesn't exist", types.NamespacedName{Namespace: "test", Name: "service"}),
		)
	})
})
