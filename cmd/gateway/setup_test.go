package main_test

import (
	"errors"

	. "github.com/nginxinc/nginx-gateway-kubernetes/cmd/gateway"
	flag "github.com/spf13/pflag"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var domain string

func MockValidator(name string, called *int, succeed bool) ValidatorContext {
	return ValidatorContext{
		name,
		func(_ *flag.FlagSet) error {
			*called++

			if !succeed {
				return errors.New("Mock error")
			}
			return nil
		},
	}
}

var _ = Describe("Main", func() {
	Describe("Generic Validator", func() {
		var mockFlags *flag.FlagSet
		BeforeEach(func() {
			mockFlags = flag.NewFlagSet("mock", flag.PanicOnError)
			_ = mockFlags.String("validator-1", "", "validator-1")
			_ = mockFlags.String("validator-2", "", "validator-2")
			_ = mockFlags.String("validator-3", "", "validator-3")
			err := mockFlags.Parse([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			mockFlags = nil
		})
		It("should call all validators", func() {
			var called int
			table := []struct {
				ExpectedCalls int
				Success       bool
				Contexts      []ValidatorContext
			}{
				{
					0,
					true,
					[]ValidatorContext{},
				},
				{
					0,
					true,
					[]ValidatorContext{
						MockValidator("no-flag-set", &called, true),
					},
				},
				{
					1,
					true,
					[]ValidatorContext{
						MockValidator("validator-1", &called, true),
					},
				},
				{
					1,
					true,
					[]ValidatorContext{
						MockValidator("no-flag-set", &called, true),
						MockValidator("validator-1", &called, true),
					},
				},
				{
					2,
					true,
					[]ValidatorContext{
						MockValidator("validator-1", &called, true),
						MockValidator("validator-2", &called, true),
					},
				},
				{
					3,
					true,
					[]ValidatorContext{
						MockValidator("validator-1", &called, true),
						MockValidator("validator-2", &called, true),
						MockValidator("validator-3", &called, true),
					},
				},
				{
					3,
					false,
					[]ValidatorContext{
						MockValidator("validator-1", &called, false),
						MockValidator("validator-2", &called, true),
						MockValidator("validator-3", &called, true),
					},
				},
				{
					3,
					false,
					[]ValidatorContext{
						MockValidator("validator-1", &called, true),
						MockValidator("validator-2", &called, true),
						MockValidator("validator-3", &called, false),
					},
				},
			}

			for i := range table {
				called = 0
				msgs := ValidateArguments(mockFlags, table[i].Contexts...)
				Expect(msgs == nil).To(Equal(table[i].Success))
				Expect(called).To(Equal(table[i].ExpectedCalls))
			}
		}) // should call all validators
	}) // Generic Validator

	Describe("CLI argument validation", func() {
		var mockFlags *flag.FlagSet
		var gatewayCtlrName string
		BeforeEach(func() {
			domain = "k8s-gateway.nginx.org"
			gatewayCtlrName = "gateway-ctlr-name"

			mockFlags = flag.NewFlagSet("mock", flag.PanicOnError)
			_ = mockFlags.String("gateway-ctlr-name", "", "mock gateway-ctlr-name")
			err := mockFlags.Parse([]string{})
			Expect(err).ToNot(HaveOccurred())
		})
		AfterEach(func() {
			mockFlags = nil
		})
		It("should parse full gateway-ctlr-name", func() {
			err := mockFlags.Set(gatewayCtlrName, "k8s-gateway.nginx.org/nginx-gateway/my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v := GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).ToNot(HaveOccurred())
		}) // should parse full gateway-ctlr-name

		It("should fail with too many path elements", func() {
			err := mockFlags.Set(gatewayCtlrName, "k8s-gateway.nginx.org/nginx-gateway/my-gateway/broken")
			Expect(err).ToNot(HaveOccurred())

			v := GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())
		}) // should fail with too many path elements

		It("should fail with too few path elements", func() {
			err := mockFlags.Set(gatewayCtlrName, "nginx-gateway/my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v := GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())

			err = mockFlags.Set(gatewayCtlrName, "my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v = GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())
		}) // should fail with too few path elements

		It("should verify constraints", func() {
			// bad domain
			err := mockFlags.Set(gatewayCtlrName, "invalid-domain/nginx-gateway/my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v := GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())

			// bad domain
			err = mockFlags.Set(gatewayCtlrName, "/default/my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v = GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())

			// bad namespace
			err = mockFlags.Set(gatewayCtlrName, "k8s-gateway.nginx.org/default/my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v = GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())

			// bad namespace
			err = mockFlags.Set(gatewayCtlrName, "k8s-gateway.nginx.org//my-gateway")
			Expect(err).ToNot(HaveOccurred())

			v = GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())

			// bad name
			err = mockFlags.Set(gatewayCtlrName, "k8s-gateway.nginx.org/default/")
			Expect(err).ToNot(HaveOccurred())

			v = GatewayControllerParam(domain, "nginx-gateway")
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			Expect(err).To(HaveOccurred())
		}) // should verify constraints
	}) // CLI argument validation
}) // end Main
