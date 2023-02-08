package commander_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommander(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commander Suite")
}
