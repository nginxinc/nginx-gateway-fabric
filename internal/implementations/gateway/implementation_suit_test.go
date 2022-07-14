package implementation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestImplementation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Implementation Suite")
}
