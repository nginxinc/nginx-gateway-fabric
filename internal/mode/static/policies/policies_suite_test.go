package policies_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPolicies(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Policies Suite")
}
