package relationship_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestRelationships(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Relationships Suite")
}
