package implementation_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestEndpointSliceImplementation(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Endpoint Slice Implementation Suite")
}
