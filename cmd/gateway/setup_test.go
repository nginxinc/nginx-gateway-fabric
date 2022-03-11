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
				return errors.New("mock error")
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
		type testCase struct {
			Param    string
			Domain   string
			ExpError bool
		}

		var mockFlags *flag.FlagSet
		var gatewayCtlrName string

		tester := func(t testCase) {
			err := mockFlags.Set(gatewayCtlrName, t.Param)
			Expect(err).ToNot(HaveOccurred())

			v := GatewayControllerParam(domain, t.Domain)
			Expect(v.V).ToNot(BeNil())

			err = v.V(mockFlags)
			if t.ExpError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).ToNot(HaveOccurred())
			}
		}
		runner := func(table []testCase) {
			for i := range table {
				tester(table[i])
			}
		}

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
			t := testCase{
				"k8s-gateway.nginx.org/nginx-gateway/my-gateway",
				"nginx-gateway",
				false,
			}
			tester(t)
		}) // should parse full gateway-ctlr-name

		It("should fail with too many path elements", func() {
			t := testCase{
				"k8s-gateway.nginx.org/nginx-gateway/my-gateway/broken",
				"nginx-gateway",
				true,
			}
			tester(t)
		}) // should fail with too many path elements

		It("should fail with too few path elements", func() {
			table := []testCase{
				{
					Param:    "nginx-gateway/my-gateway",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
				{
					Param:    "my-gateway",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
			}

			runner(table)
		}) // should fail with too few path elements

		It("should verify constraints", func() {
			table := []testCase{
				{
					// bad domain
					Param:    "invalid-domain/nginx-gateway/my-gateway",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
				{
					// bad domain
					Param:    "/default/my-gateway",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
				{
					// bad namespace
					Param:    "k8s-gateway.nginx.org/default/my-gateway",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
				{
					// bad namespace
					Param:    "k8s-gateway.nginx.org//my-gateway",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
				{
					// bad name
					Param:    "k8s-gateway.nginx.org/default/",
					Domain:   "nginx-gateway",
					ExpError: true,
				},
			}

			runner(table)
		}) // should verify constraints
	}) // CLI argument validation
}) // end Main
