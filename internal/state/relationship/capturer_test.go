package relationship_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	v1 "k8s.io/api/core/v1"
	discoveryV1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/nginxinc/nginx-kubernetes-gateway/internal/helpers"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/manager/index"
	"github.com/nginxinc/nginx-kubernetes-gateway/internal/state/relationship"
)

func createBackendRefs(backendNames ...v1beta1.ObjectName) []v1beta1.HTTPBackendRef {
	refs := make([]v1beta1.HTTPBackendRef, 0, len(backendNames))
	for _, name := range backendNames {
		refs = append(refs, v1beta1.HTTPBackendRef{
			BackendRef: v1beta1.BackendRef{
				BackendObjectReference: v1beta1.BackendObjectReference{
					Kind:      (*v1beta1.Kind)(helpers.GetStringPointer("Service")),
					Name:      name,
					Namespace: (*v1beta1.Namespace)(helpers.GetStringPointer("test")),
				},
			},
		})
	}

	return refs
}

func createRules(backendRefs ...[]v1beta1.HTTPBackendRef) []v1beta1.HTTPRouteRule {
	rules := make([]v1beta1.HTTPRouteRule, 0, len(backendRefs))
	for _, refs := range backendRefs {
		rules = append(rules, v1beta1.HTTPRouteRule{BackendRefs: refs})
	}

	return rules
}

func createRoute(name string, rules []v1beta1.HTTPRouteRule) *v1beta1.HTTPRoute {
	return &v1beta1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{Namespace: "test", Name: name},
		Spec:       v1beta1.HTTPRouteSpec{Rules: rules},
	}
}

var _ = Describe("Capturer", func() {
	var (
		capturer *relationship.CapturerImpl

		backendRef1 = createBackendRefs("svc1")
		backendRef2 = createBackendRefs("svc2")
		backendRef3 = createBackendRefs("svc3")
		backendRef4 = createBackendRefs("svc4")

		hr1 = createRoute("hr1", createRules(backendRef1))
		hr2 = createRoute("hr2", createRules(backendRef2, backendRef3, backendRef4))

		hrSvc1AndSvc2 = createRoute("hr-svc1-svc2", createRules(backendRef1, backendRef2))
		hrSvc1AndSvc3 = createRoute("hr-svc1-svc3", createRules(backendRef3, backendRef1))
		hrSvc1AndSvc4 = createRoute("hr-svc1-svc4", createRules(backendRef1, backendRef4))

		hr1Name           = types.NamespacedName{Namespace: hr1.Namespace, Name: hr1.Name}
		hr2Name           = types.NamespacedName{Namespace: hr2.Namespace, Name: hr2.Name}
		hrSvc1AndSvc2Name = types.NamespacedName{Namespace: hrSvc1AndSvc2.Namespace, Name: hrSvc1AndSvc2.Name}
		hrSvc1AndSvc3Name = types.NamespacedName{Namespace: hrSvc1AndSvc3.Namespace, Name: hrSvc1AndSvc3.Name}
		hrSvc1AndSvc4Name = types.NamespacedName{Namespace: hrSvc1AndSvc4.Namespace, Name: hrSvc1AndSvc4.Name}

		svc1 = types.NamespacedName{Namespace: "test", Name: "svc1"}
		svc2 = types.NamespacedName{Namespace: "test", Name: "svc2"}
		svc3 = types.NamespacedName{Namespace: "test", Name: "svc3"}
		svc4 = types.NamespacedName{Namespace: "test", Name: "svc4"}
	)

	Describe("Capture service relationships for routes", func() {
		BeforeEach(OncePerOrdered, func() {
			capturer = relationship.NewCapturerImpl()
		})

		assertServiceExists := func(svcName types.NamespacedName, exists bool, refCount int) {
			ExpectWithOffset(1, capturer.Exists(&v1.Service{}, svcName)).To(Equal(exists))
			ExpectWithOffset(1, capturer.GetRefCountForService(svcName)).To(Equal(refCount))
		}

		Describe("Normal cases", Ordered, func() {
			When("a route with a backend service is captured", func() {
				It("reports a service relationship", func() {
					capturer.Capture(hr1)

					assertServiceExists(svc1, true, 1)
				})
			})
			When("a route with multiple backend services is captured", func() {
				It("reports all service relationships for all captured routes", func() {
					capturer.Capture(hr2)

					assertServiceExists(svc1, true, 1)
					assertServiceExists(svc2, true, 1)
					assertServiceExists(svc3, true, 1)
					assertServiceExists(svc4, true, 1)
				})
			})
			When("one backend service is removed from a captured route", func() {
				It("removes the correct service relationship", func() {
					hr2Updated := hr2.DeepCopy()
					hr2Updated.Spec.Rules = hr2Updated.Spec.Rules[0:2] // remove the last rule

					capturer.Capture(hr2Updated)

					assertServiceExists(svc1, true, 1)
					assertServiceExists(svc2, true, 1)
					assertServiceExists(svc3, true, 1)
					assertServiceExists(svc4, false, 0)
				})
			})
			When("one backend service is added to a captured route", func() {
				It("adds the correct service relationship", func() {
					capturer.Capture(hr2)

					assertServiceExists(svc1, true, 1)
					assertServiceExists(svc2, true, 1)
					assertServiceExists(svc3, true, 1)
					assertServiceExists(svc4, true, 1)
				})
			})
			When("a route with multiple backend services is removed", func() {
				It("removes all service relationships", func() {
					capturer.Remove(&v1beta1.HTTPRoute{}, hr2Name)

					assertServiceExists(svc2, false, 0)
					assertServiceExists(svc3, false, 0)
					assertServiceExists(svc4, false, 0)

					// Service referenced by hr1 still exists
					assertServiceExists(svc1, true, 1)
				})
			})
			When("a route is removed", func() {
				It("removes service relationships", func() {
					capturer.Remove(&v1beta1.HTTPRoute{}, hr1Name)

					assertServiceExists(svc1, false, 0)
				})
			})
		})
		Describe("Multiple routes that reference the same service", Ordered, func() {
			When("multiple routes are captured that all reference the same service", func() {
				It("reports all service relationships", func() {
					capturer.Capture(hr1)
					capturer.Capture(hrSvc1AndSvc2)
					capturer.Capture(hrSvc1AndSvc3)
					capturer.Capture(hrSvc1AndSvc4)

					assertServiceExists(svc1, true, 4)
					assertServiceExists(svc2, true, 1)
					assertServiceExists(svc3, true, 1)
					assertServiceExists(svc4, true, 1)
				})
			})
			When("one route is removed", func() {
				It("reports remaining service relationships", func() {
					capturer.Remove(&v1beta1.HTTPRoute{}, hr1Name)

					// ref count for svc1 should decrease by one
					assertServiceExists(svc1, true, 3)

					// all other ref counts stay the same
					assertServiceExists(svc2, true, 1)
					assertServiceExists(svc3, true, 1)
					assertServiceExists(svc4, true, 1)
				})
			})
			When("another route is removed", func() {
				It("reports remaining service relationships", func() {
					capturer.Remove(&v1beta1.HTTPRoute{}, hrSvc1AndSvc2Name)

					// svc2 should no longer exist
					assertServiceExists(svc2, false, 0)

					// ref count for svc1 should decrease by one
					assertServiceExists(svc1, true, 2)

					// all other ref counts stay the same
					assertServiceExists(svc3, true, 1)
					assertServiceExists(svc4, true, 1)
				})
			})
			When("another route is removed", func() {
				It("reports remaining service relationships", func() {
					capturer.Remove(&v1beta1.HTTPRoute{}, hrSvc1AndSvc3Name)

					// svc3 should no longer exist
					assertServiceExists(svc3, false, 0)

					// svc2 should still not exist
					assertServiceExists(svc2, false, 0)

					// ref count for svc1 should decrease by one
					assertServiceExists(svc1, true, 1)

					// svc4 ref count should stay the same
					assertServiceExists(svc4, true, 1)
				})
				When("final route is removed", func() {
					It("removes all service relationships", func() {
						capturer.Remove(&v1beta1.HTTPRoute{}, hrSvc1AndSvc4Name)

						// no services should exist and all ref counts should be 0
						assertServiceExists(svc1, false, 0)
						assertServiceExists(svc2, false, 0)
						assertServiceExists(svc3, false, 0)
						assertServiceExists(svc4, false, 0)
					})
				})
				When("route is removed again", func() {
					It("service ref counts remain at 0", func() {
						capturer.Remove(&v1beta1.HTTPRoute{}, hrSvc1AndSvc4Name)

						// no services should exist and all ref counts should still be 0
						assertServiceExists(svc1, false, 0)
						assertServiceExists(svc2, false, 0)
						assertServiceExists(svc3, false, 0)
						assertServiceExists(svc4, false, 0)
					})
				})
			})
		})
		Describe("Capture endpoint slice relationships", func() {
			var (
				slice1 = &discoveryV1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "es1",
						Labels:    map[string]string{index.KubernetesServiceNameLabel: "svc1"},
					},
				}

				slice2 = &discoveryV1.EndpointSlice{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "es2",
						Labels:    map[string]string{index.KubernetesServiceNameLabel: "svc1"},
					},
				}

				slice1Name = types.NamespacedName{Namespace: slice1.Namespace, Name: slice1.Name}
				slice2Name = types.NamespacedName{Namespace: slice2.Namespace, Name: slice2.Name}
			)

			BeforeEach(OncePerOrdered, func() {
				capturer = relationship.NewCapturerImpl()
			})

			Describe("Normal cases", Ordered, func() {
				When("an endpoint slice is captured that has an unrelated service owner", func() {
					It("does not report an endpoint slice relationship", func() {
						capturer.Capture(slice1)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeFalse())
					})
				})
				When("a relationship is captured for the service owner", func() {
					It("adds an endpoint slice relationship", func() {
						capturer.Capture(hr1)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeTrue())
					})
				})
				When("another endpoint slice is captured with the same service owner", func() {
					It("adds another endpoint slice relationship", func() {
						capturer.Capture(slice2)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeTrue())
						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice2Name)).To(BeTrue())
					})
				})
				When("an endpoint slice is removed", func() {
					It("removes the endpoint slice relationship", func() {
						capturer.Remove(&discoveryV1.EndpointSlice{}, slice2Name)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice2Name)).To(BeFalse())

						// slice 1 relationship should still exist
						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeTrue())
					})
				})
				When("endpoint slice service owner changes to an unrelated service owner", func() {
					It("removes the endpoint slice relationship", func() {
						updatedSlice1 := slice1.DeepCopy()
						updatedSlice1.Labels[index.KubernetesServiceNameLabel] = "unrelated-svc"

						capturer.Capture(updatedSlice1)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeFalse())
					})
				})
				When("endpoint slice service owner changes to a related service owner", func() {
					It("adds an endpoint slice relationship", func() {
						capturer.Capture(slice1)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeTrue())
					})
				})
				When("service relationship is removed", func() {
					It("removes the endpoint slice relationship", func() {
						capturer.Remove(&v1beta1.HTTPRoute{}, hr1Name)

						Expect(capturer.Exists(&discoveryV1.EndpointSlice{}, slice1Name)).To(BeFalse())
					})
				})
			})
		})
		Describe("Edge cases", func() {
			BeforeEach(func() {
				capturer = relationship.NewCapturerImpl()
			})
			It("Capture does not panic when passed an unsupported resource type", func() {
				Expect(func() {
					capturer.Capture(&v1beta1.Gateway{})
				}).ToNot(Panic())
			})
			It("Remove does not panic when passed an unsupported resource type", func() {
				Expect(func() {
					capturer.Remove(&v1beta1.Gateway{}, types.NamespacedName{})
				}).ToNot(Panic())
			})
			It("Exist returns false if passed an unsupported resource type", func() {
				Expect(capturer.Exists(&v1beta1.Gateway{}, types.NamespacedName{})).To(BeFalse())
			})
		})
	})
})
