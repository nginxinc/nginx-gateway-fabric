package main

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("NginxGateway", Ordered, Label("functional", "nginxGateway"), func() {
	var (
		ngfPodName string

		namespace          = "nginx-gateway"
		nginxGatewayNsname = types.NamespacedName{Name: releaseName + "-config", Namespace: namespace}

		files = []string{
			"nginxgateway/nginx-gateway.yaml",
		}
	)

	AfterAll(func() {
		// cleanup and restore NGF instance
		teardown(releaseName)
		setup(getDefaultSetupCfg())
	})

	getNginxGateway := func(nsname types.NamespacedName) (ngfAPI.NginxGateway, error) {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
		defer cancel()

		var nginxGateway ngfAPI.NginxGateway

		if err := k8sClient.Get(ctx, nsname, &nginxGateway); err != nil {
			return nginxGateway, err
		}

		return nginxGateway, nil
	}

	verifyAndReturnNginxGateway := func(nsname types.NamespacedName) ngfAPI.NginxGateway {
		nginxGateway, err := getNginxGateway(nsname)
		Expect(err).ToNot(HaveOccurred())
		Expect(nginxGateway).ToNot(BeNil())

		Expect(nginxGateway.Status.Conditions).To(HaveLen(1))
		condition := nginxGateway.Status.Conditions[0]

		Expect(condition.Type).To(Equal("Valid"))
		Expect(condition.Reason).To(Equal("Valid"))

		return nginxGateway
	}

	getNGFPodName := func() (string, error) {
		podNames, err := framework.GetReadyNGFPodNames(
			k8sClient,
			ngfNamespace,
			releaseName,
			timeoutConfig.GetTimeout,
		)
		if err != nil {
			return "", err
		}

		if len(podNames) != 1 {
			return "", fmt.Errorf("expected 1 pod name, got %d", len(podNames))
		}

		return podNames[0], nil
	}

	When("testing NGF on startup", func() {
		When("log level is set to debug", func() {
			It("outputs debug logs and the status is accepted and true", func() {
				ngfPodName, err := getNGFPodName()
				Expect(err).ToNot(HaveOccurred())

				_ = verifyAndReturnNginxGateway(nginxGatewayNsname)

				logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
					Container: "nginx-gateway",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logs).To(ContainSubstring("\"level\":\"debug\""))
			})
		})

		When("default log level is used", func() {
			It("only outputs info logs and the status is accepted and true", func() {
				teardown(releaseName)

				cfg := getDefaultSetupCfg()
				cfg.infoLogLevel = true
				setup(cfg)

				ngfPodName, err := getNGFPodName()
				Expect(err).ToNot(HaveOccurred())

				nginxGateway := verifyAndReturnNginxGateway(nginxGatewayNsname)
				Expect(nginxGateway.Status.Conditions[0].ObservedGeneration).To(Equal(int64(1)))

				logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
					Container: "nginx-gateway",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logs).ToNot(ContainSubstring("\"level\":\"debug\""))
			})
		})
	})

	When("testing on an existing NGF instance", Ordered, func() {
		BeforeAll(func() {
			var err error
			ngfPodName, err = getNGFPodName()
			Expect(err).ToNot(HaveOccurred())
		})

		When("NginxGateway is updated", func() {
			It("captures the change, the status is accepted and true,"+
				" and the observed generation is incremented", func() {
				// previous test has left the log level at info, this test will change the log level to debug

				var observedGeneration int64

				nginxGateway := verifyAndReturnNginxGateway(nginxGatewayNsname)
				observedGeneration = nginxGateway.Status.Conditions[0].ObservedGeneration

				logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
					Container: "nginx-gateway",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logs).ToNot(ContainSubstring("\"level\":\"debug\""))

				Expect(resourceManager.ApplyFromFiles(files, namespace)).To(Succeed())
				// need to wait until files are applied, no current function because this is a nginx-gateway crd
				time.Sleep(2 * time.Second)

				nginxGateway = verifyAndReturnNginxGateway(nginxGatewayNsname)
				Expect(nginxGateway.Status.Conditions[0].ObservedGeneration).To(Equal(observedGeneration + 1))

				logs, err = resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
					Container: "nginx-gateway",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logs).To(ContainSubstring(
					"\"current\":\"debug\",\"msg\":\"Log level changed\",\"prev\":\"info\"",
				))
			})
		})

		When("NginxGateway is deleted", func() {
			It("captures the deletion and default values are used", func() {
				Expect(resourceManager.DeleteFromFiles(files, namespace)).To(Succeed())
				time.Sleep(2 * time.Second) // need to wait until deletion is fully processed

				_, err := getNginxGateway(nginxGatewayNsname)
				Expect(err).Should(HaveOccurred())

				logs, err = resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
					Container: "nginx-gateway",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logs).To(ContainSubstring("NginxGateway configuration was deleted; using defaults"))
			})
		})
	})
})