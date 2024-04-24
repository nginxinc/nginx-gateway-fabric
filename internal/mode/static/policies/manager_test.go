package policies_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies/policiesfakes"
)

var _ = Describe("Policy Manager", func() {
	orangeGVK := schema.GroupVersionKind{Group: "fruit", Version: "1", Kind: "orange"}
	orangePolicy := &policiesfakes.FakePolicy{
		GetNameStub: func() string {
			return "orange"
		},
	}

	appleGVK := schema.GroupVersionKind{Group: "fruit", Version: "1", Kind: "apple"}
	applePolicy := &policiesfakes.FakePolicy{
		GetNameStub: func() string {
			return "apple"
		},
	}

	mustExtractGVK := func(object client.Object) schema.GroupVersionKind {
		switch object.GetName() {
		case "apple":
			return appleGVK
		case "orange":
			return orangeGVK
		default:
			return schema.GroupVersionKind{}
		}
	}

	mgr := policies.NewManager(
		mustExtractGVK,
		policies.ManagerConfig{
			Validator: &policiesfakes.FakeValidator{
				ValidateStub:  func(_ policies.Policy) error { return errors.New("apple error") },
				ConflictsStub: func(_ policies.Policy, _ policies.Policy) bool { return true },
			},
			Generator: func(_ policies.Policy) []byte {
				return []byte("apple")
			},
			GVK: appleGVK,
		},
		policies.ManagerConfig{
			Validator: &policiesfakes.FakeValidator{
				ValidateStub:  func(_ policies.Policy) error { return errors.New("orange error") },
				ConflictsStub: func(_ policies.Policy, _ policies.Policy) bool { return false },
			},
			Generator: func(_ policies.Policy) []byte {
				return []byte("orange")
			},
			GVK: orangeGVK,
		},
	)

	Context("Validation", func() {
		When("Policy is registered with manager", func() {
			It("Validates the policy", func() {
				err := mgr.Validate(applePolicy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("apple error"))

				err = mgr.Validate(orangePolicy)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("orange error"))
			})
			It("Returns whether the policies conflict", func() {
				Expect(mgr.Conflicts(applePolicy, applePolicy)).To(BeTrue())
				Expect(mgr.Conflicts(orangePolicy, orangePolicy)).To(BeFalse())
			})
		})
		When("Policy is not registered with manager", func() {
			It("Panics on call to validate", func() {
				validate := func() {
					_ = mgr.Validate(&policiesfakes.FakePolicy{})
				}

				Expect(validate).To(Panic())
			})
			It("panics on call to conflicts", func() {
				conflict := func() {
					_ = mgr.Conflicts(&policiesfakes.FakePolicy{}, &policiesfakes.FakePolicy{})
				}

				Expect(conflict).To(Panic())
			})
		})
	})
	Context("Generation", func() {
		When("Policy is registered with manager", func() {
			It("Generates the configuration for the policy", func() {
				Expect(mgr.Generate(applePolicy)).To(Equal([]byte("apple")))
				Expect(mgr.Generate(orangePolicy)).To(Equal([]byte("orange")))
			})
		})
		When("Policy is not registered with manager", func() {
			It("Panics on generate", func() {
				generate := func() {
					_ = mgr.Generate(&policiesfakes.FakePolicy{})
				}

				Expect(generate).To(Panic())
			})
		})
	})
})
