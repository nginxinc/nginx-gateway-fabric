package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
			return nginxGateway, errors.New("failed to get nginxGateway")
		}

		return nginxGateway, nil
	}

	verifyAndReturnNginxGateway := func(nsname types.NamespacedName) (ngfAPI.NginxGateway, error) {
		nginxGateway, err := getNginxGateway(nsname)
		if err != nil {
			return nginxGateway, err
		}

		if nginxGateway.Status.Conditions == nil {
			return nginxGateway, errors.New("nginxGateway is has no conditions")
		}

		if len(nginxGateway.Status.Conditions) != 1 {
			return nginxGateway,
				fmt.Errorf("expected nginxGateway to have only one condition, instead has %d conditions",
					len(nginxGateway.Status.Conditions))
		}

		condition := nginxGateway.Status.Conditions[0]

		if condition.Type != "Valid" {
			return nginxGateway, fmt.Errorf("expected nginxGateway condition type to be Valid,"+
				" instead has type %s", condition.Type)
		}

		if condition.Reason != "Valid" {
			return nginxGateway, fmt.Errorf("expected nginxGateway reason to be Valid,"+
				" instead is %s", condition.Reason)
		}

		return nginxGateway, nil
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

				_, err = verifyAndReturnNginxGateway(nginxGatewayNsname)
				Expect(err).ToNot(HaveOccurred())

				Eventually(
					func() bool {
						logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
							Container: "nginx-gateway",
						})
						if err != nil {
							return false
						}

						return strings.Contains(logs, "\"level\":\"debug\"")
					}).WithTimeout(timeoutConfig.GetTimeout).
					WithPolling(500 * time.Millisecond).
					Should(BeTrue())
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

				nginxGateway, err := verifyAndReturnNginxGateway(nginxGatewayNsname)
				Expect(err).ToNot(HaveOccurred())

				Expect(nginxGateway.Status.Conditions[0].ObservedGeneration).To(Equal(int64(1)))

				Eventually(
					func() bool {
						logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
							Container: "nginx-gateway",
						})
						if err != nil {
							return false
						}

						return !strings.Contains(logs, "\"level\":\"debug\"")
					}).WithTimeout(timeoutConfig.GetTimeout).
					WithPolling(500 * time.Millisecond).
					Should(BeTrue())
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

				nginxGateway, err := verifyAndReturnNginxGateway(nginxGatewayNsname)
				Expect(err).ToNot(HaveOccurred())

				observedGeneration = nginxGateway.Status.Conditions[0].ObservedGeneration

				logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
					Container: "nginx-gateway",
				})
				Expect(err).ToNot(HaveOccurred())

				Expect(logs).ToNot(ContainSubstring("\"level\":\"debug\""))

				Expect(resourceManager.ApplyFromFiles(files, namespace)).To(Succeed())

				Eventually(
					func() bool {
						nginxGateway, err := verifyAndReturnNginxGateway(nginxGatewayNsname)
						if err != nil {
							return false
						}

						return nginxGateway.Status.Conditions[0].ObservedGeneration == observedGeneration+1
					}).WithTimeout(timeoutConfig.UpdateTimeout).
					WithPolling(500 * time.Millisecond).
					Should(BeTrue())

				Eventually(
					func() bool {
						logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
							Container: "nginx-gateway",
						})
						if err != nil {
							return false
						}

						return strings.Contains(logs,
							"\"current\":\"debug\",\"msg\":\"Log level changed\",\"prev\":\"info\"")
					}).WithTimeout(timeoutConfig.GetTimeout).
					WithPolling(500 * time.Millisecond).
					Should(BeTrue())
			})
		})

		When("NginxGateway is deleted", func() {
			It("captures the deletion and default values are used", func() {
				Expect(resourceManager.DeleteFromFiles(files, namespace)).To(Succeed())

				Eventually(
					func() error {
						_, err := verifyAndReturnNginxGateway(nginxGatewayNsname)
						return err
					}).WithTimeout(timeoutConfig.DeleteTimeout).
					WithPolling(500 * time.Millisecond).
					Should(MatchError("failed to get nginxGateway"))

				Eventually(
					func() bool {
						logs, err := resourceManager.GetPodLogs(ngfNamespace, ngfPodName, &core.PodLogOptions{
							Container: "nginx-gateway",
						})
						if err != nil {
							return false
						}

						return strings.Contains(logs, "NginxGateway configuration was deleted; using defaults")
					}).WithTimeout(timeoutConfig.GetTimeout).
					WithPolling(500 * time.Millisecond).
					Should(BeTrue())
			})
		})
	})
})
