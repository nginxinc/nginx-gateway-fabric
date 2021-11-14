package main_test

import (
	. "github.com/nginxinc/nginx-gateway-kubernetes/cmd/gateway"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	Describe("CLI argument validation", func() {
		It("should parse full gateway-ctlr-name", func() {
			gatewayCtlrName := "k8s-gateway.nginx.org/nginx-gateway/my-gateway"

			v := GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		}) // should parse full gateway-ctlr-name

		It("should parse ns/name gateway-ctlr-name", func() {
			gatewayCtlrName := "nginx-gateway/my-gateway"

			v := GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		}) // should parse ns/name gateway-ctlr-name

		It("should parse name gateway-ctlr-name", func() {
			gatewayCtlrName := "my-gateway"

			v := GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeTrue())
			Expect(err).ToNot(HaveOccurred())
		}) // should parse name gateway-ctlr-name

		It("should fail with too many path elements", func() {
			gatewayCtlrName := "nginx.org/nginx/my-gateway/broken"

			v := GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())
		}) // should fail with too many path elements

		It("should verify constraints", func() {
			// bad domain
			gatewayCtlrName := "invalid-domain/nginx-gateway/my-gateway"

			v := GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err := v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad domain
			gatewayCtlrName = "/default/my-gateway"

			v = GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad namespace
			gatewayCtlrName = "k8s-gateway.nginx.org/default/my-gateway"

			v = GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad namespace
			gatewayCtlrName = "k8s-gateway.nginx.org//my-gateway"

			v = GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())

			// bad name
			gatewayCtlrName = "k8s-gateway.nginx.org/default/"

			v = GatewayControllerParam(true, "nginx-gateway", gatewayCtlrName)
			Expect(v).ToNot(BeNil())

			r, err = v()
			Expect(r).To(BeFalse())
			Expect(err).To(HaveOccurred())
		}) // should verify constraints
	}) // CLI argument validation
}) // end Main
