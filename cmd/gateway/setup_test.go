package main_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	flag "github.com/spf13/pflag"

	. "github.com/nginxinc/nginx-kubernetes-gateway/cmd/gateway"
)

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
			Flag             string
			Value            string
			ValidatorContext ValidatorContext
			ExpError         bool
		}

		const (
			expectError   = true
			expectSuccess = false
		)

		var mockFlags *flag.FlagSet

		tester := func(t testCase) {
			err := mockFlags.Set(t.Flag, t.Value)
			Expect(err).ToNot(HaveOccurred())

			err = t.ValidatorContext.V(mockFlags)

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

		Describe("gateway-ctlr-name validation", func() {
			prepareTestCase := func(value string, expError bool) testCase {
				return testCase{
					Flag:             "gateway-ctlr-name",
					Value:            value,
					ValidatorContext: GatewayControllerParam("k8s-gateway.nginx.org", "nginx-gateway"),
					ExpError:         expError,
				}
			}

			BeforeEach(func() {
				mockFlags = flag.NewFlagSet("mock", flag.PanicOnError)
				_ = mockFlags.String("gateway-ctlr-name", "", "mock gateway-ctlr-name")
				err := mockFlags.Parse([]string{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				mockFlags = nil
			})

			It("should parse full gateway-ctlr-name", func() {
				t := prepareTestCase(
					"k8s-gateway.nginx.org/nginx-gateway/my-gateway",
					expectSuccess,
				)
				tester(t)
			}) // should parse full gateway-ctlr-name

			It("should fail with too many path elements", func() {
				t := prepareTestCase(
					"k8s-gateway.nginx.org/nginx-gateway/my-gateway/broken",
					expectError)
				tester(t)
			}) // should fail with too many path elements

			It("should fail with too few path elements", func() {
				table := []testCase{
					prepareTestCase(
						"nginx-gateway/my-gateway",
						expectError,
					),
					prepareTestCase(
						"my-gateway",
						expectError,
					),
				}

				runner(table)
			}) // should fail with too few path elements

			It("should verify constraints", func() {
				table := []testCase{
					prepareTestCase(
						// bad domain
						"invalid-domain/nginx-gateway/my-gateway",
						expectError,
					),
					prepareTestCase(
						// bad domain
						"/default/my-gateway",
						expectError,
					),
					prepareTestCase(
						// bad namespace
						"k8s-gateway.nginx.org/default/my-gateway",
						expectError),
					prepareTestCase(
						// bad namespace
						"k8s-gateway.nginx.org//my-gateway",
						expectError,
					),
					prepareTestCase(
						// bad name
						"k8s-gateway.nginx.org/default/",
						expectError,
					),
				}

				runner(table)
			}) // should verify constraints
		}) // gateway-ctlr-name validation

		Describe("gatewayclass validation", func() {
			prepareTestCase := func(value string, expError bool) testCase {
				return testCase{
					Flag:             "gatewayclass",
					Value:            value,
					ValidatorContext: GatewayClassParam(),
					ExpError:         expError,
				}
			}

			BeforeEach(func() {
				mockFlags = flag.NewFlagSet("mock", flag.PanicOnError)
				_ = mockFlags.String("gatewayclass", "", "mock gatewayclass")
				err := mockFlags.Parse([]string{})
				Expect(err).ToNot(HaveOccurred())
			})
			AfterEach(func() {
				mockFlags = nil
			})

			It("should succeed on valid name", func() {
				t := prepareTestCase(
					"nginx",
					expectSuccess,
				)
				tester(t)
			}) // should succeed on valid name

			It("should fail with invalid name", func() {
				t := prepareTestCase(
					"$nginx",
					expectError)
				tester(t)
			}) // should fail with invalid name"
		}) // gatewayclass validation
	}) // CLI argument validation
}) // end Main
