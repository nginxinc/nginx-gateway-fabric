package main_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGateway(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gateway Suite")
}
