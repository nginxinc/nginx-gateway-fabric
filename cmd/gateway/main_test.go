package main

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	type testCase struct {
		expSubMsg string
		podIP     string
		expErr    bool
	}
	DescribeTable("should validate an IP address",
		func(tc testCase) {
			err := validateIP(tc.podIP)
			if !tc.expErr {
				Expect(err).ToNot(HaveOccurred())
			} else {
				Expect(err.Error()).To(ContainSubstring(tc.expSubMsg))
			}
		},
		Entry("var not set", testCase{podIP: "", expErr: true, expSubMsg: "must be set"}),
		Entry("var set to invalid value", testCase{podIP: "invalid", expErr: true, expSubMsg: "must be a valid"}),
		Entry("var set to valid value", testCase{podIP: "1.2.3.4", expErr: false}),
	) // should validate an IP address
})
