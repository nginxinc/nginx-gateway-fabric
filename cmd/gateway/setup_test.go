package main_test

import (
	"errors"

	"github.com/go-logr/logr"
	. "github.com/nginxinc/nginx-gateway-kubernetes/cmd/gateway"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var domain string

func MockValidator(called *int, succeed bool) Validator {
	return func() (bool, error) {
		*called++

		if !succeed {
			return succeed, errors.New("Mock error")
		}
		return succeed, nil
	}
}

var _ = Describe("Main", func() {
	Describe("Generic Validator", func() {
		It("should call all validators", func() {
			var called int
			table := []struct {
				ExpectedCalls int
				Success       bool
				Validators    []Validator
			}{
				{
					0,
					true,
					[]Validator{},
				},
				{
					1,
					true,
					[]Validator{
						MockValidator(&called, true),
					},
				},
				{
					2,
					true,
					[]Validator{
						MockValidator(&called, true),
						MockValidator(&called, true),
					},
				},
				{
					3,
					true,
					[]Validator{
						MockValidator(&called, true),
						MockValidator(&called, true),
						MockValidator(&called, true),
					},
				},
				{
					3,
					false,
					[]Validator{
						MockValidator(&called, false),
						MockValidator(&called, true),
						MockValidator(&called, true),
					},
				},
				{
					3,
					false,
					[]Validator{
						MockValidator(&called, true),
						MockValidator(&called, true),
						MockValidator(&called, false),
					},
				},
			}

			for i := range table {
				called = 0
				ret := ValidateArguments(logr.Discard(), table[i].Validators...)
				Expect(ret).To(Equal(table[i].Success))
				Expect(called).To(Equal(table[i].ExpectedCalls))
			}
		}) // should call all validators
	}) // Generic Validator

	Describe("CLI argument validation", func() {
		BeforeEach(func() {
			domain = "k8s-gateway.nginx.org"
		})
		It("should parse full gateway-ctlr-name", func() {
			gatewayCtlrName := "k8s-gateway.nginx.org/nginx-gateway/my-gateway"

			v := GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		}) // should parse full gateway-ctlr-name

		It("should fail with too many path elements", func() {
			gatewayCtlrName := "k8s-gateway.nginx.org/nginx-gateway/my-gateway/broken"

			v := GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())
		}) // should fail with too many path elements

		It("should fail with too few path elements", func() {
			gatewayCtlrName := "nginx-gateway/my-gateway"

			v := GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			gatewayCtlrName = "my-gateway"

			v = GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())
		}) // should fail with too few path elements

		It("should verify constraints", func() {
			// bad domain
			gatewayCtlrName := "invalid-domain/nginx-gateway/my-gateway"

			v := GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad domain
			gatewayCtlrName = "/default/my-gateway"

			v = GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad namespace
			gatewayCtlrName = "k8s-gateway.nginx.org/default/my-gateway"

			v = GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad namespace
			gatewayCtlrName = "k8s-gateway.nginx.org//my-gateway"

			v = GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad name
			gatewayCtlrName = "k8s-gateway.nginx.org/default/"

			v = GatewayControllerParam(true, domain, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())
		}) // should verify constraints
	}) // CLI argument validation
}) // end Main
