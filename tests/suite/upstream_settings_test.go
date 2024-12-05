package main

import (
	"context"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("UpstreamSettingsPolicy", Ordered, Label("uspolicy"), func() {
	var (
		files = []string{
			"upstream-settings-policy/cafe.yaml",
			"upstream-settings-policy/gateway.yaml",
			"upstream-settings-policy/grpc-backend.yaml",
			"upstream-settings-policy/routes.yaml",
		}

		namespace = "uspolicy"
	)

	BeforeAll(func() {
		ns := &core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, namespace)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(namespace)).To(Succeed())
	})

	AfterAll(func() {
		Expect(resourceManager.DeleteNamespace(namespace)).To(Succeed())
	})

	When("UpstreamSettingsPolicy are applied to the resources", func() {
		snippetsFilter := []string{
			"upstream-settings-policy/valid-usps.yaml",
		}

		BeforeAll(func() {
			Expect(resourceManager.ApplyFromFiles(snippetsFilter, namespace)).To(Succeed())
		})

		AfterAll(func() {
			Expect(resourceManager.DeleteFromFiles(snippetsFilter, namespace)).To(Succeed())
		})

		Specify("usPolicies are accepted", func() {
			usPolicies := []string{
				"coffee-svc-usp",
				"grpc-svc-usp",
				"tea-multiple-svc-usp",
			}

			for _, name := range usPolicies {
				nsname := types.NamespacedName{Name: name, Namespace: namespace}

				err := waitForUSPolicyStatus(nsname, metav1.ConditionTrue, v1alpha2.PolicyReasonAccepted)
				Expect(err).ToNot(HaveOccurred(), fmt.Sprintf("%s was not accepted", name))
			}
		})

		Context("verify working traffic", func() {
			It("should return a 200 response for HTTPRoute", func() {
				port := 80
				if portFwdPort != 0 {
					port = portFwdPort
				}
				baseCoffeeURL := fmt.Sprintf("http://cafe.example.com:%d%s", port, "/coffee")
				baseTeaURL := fmt.Sprintf("http://cafe.example.com:%d%s", port, "/tea")

				Eventually(
					func() error {
						return expectRequestToSucceed(baseCoffeeURL, address, "URI: /coffee")
					}).
					WithTimeout(timeoutConfig.RequestTimeout).
					WithPolling(500 * time.Millisecond).
					Should(Succeed())

				Eventually(
					func() error {
						return expectRequestToSucceed(baseTeaURL, address, "URI: /tea")
					}).
					WithTimeout(timeoutConfig.RequestTimeout).
					WithPolling(500 * time.Millisecond).
					Should(Succeed())
			})
		})

		Context("nginx directives", func() {
			var conf *framework.Payload

			BeforeAll(func() {
				podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
				Expect(err).ToNot(HaveOccurred())
				Expect(podNames).To(HaveLen(1))

				ngfPodName := podNames[0]

				conf, err = resourceManager.GetNginxConfig(ngfPodName, ngfNamespace)
				Expect(err).ToNot(HaveOccurred())
			})

			// TODO: important
			// The directive file and field value need to be updated based on the
			// implementation of the UpstreamSettingsPolicy and how they are specified in the config files.
			DescribeTable("are set properly for",
				func(expCfgs []framework.ExpectedNginxField) {
					for _, expCfg := range expCfgs {
						Expect(framework.ValidateNginxFieldExists(conf, expCfg)).To(Succeed())
					}
				},
				Entry("HTTPRoute", []framework.ExpectedNginxField{
					{
						Directive:             "zone default_coffee_80",
						Value:                 "128k",
						Upstreams:             []string{"default_coffee_80"},
						File:                  "nginx.conf",
						ValueSubstringAllowed: true,
					},
					{
						Directive: "keepalive",
						Value:     "12",
						Upstreams: []string{"default_coffee_80"},
						File:      "nginx.conf",
					},
					{
						Directive: "keepalive_requests",
						Value:     "31",
						Upstreams: []string{"default_coffee_80"},
						File:      "nginx.conf",
					},
					{
						Directive: "keepalive_time",
						Value:     "20s",
						Upstreams: []string{"default_coffee_80"},
						File:      "nginx.conf",
					},
					{
						Directive: "keepalive_timeout",
						Value:     "40s",
						Upstreams: []string{"default_coffee_80"},
						File:      "nginx.conf",
					},
				}),
				Entry("GRPCRoute", []framework.ExpectedNginxField{
					{
						Directive:             "zone default_grpc-backend_8080",
						Value:                 "512k",
						Upstreams:             []string{"default_grpc-backend_8080"},
						File:                  "nginx.conf",
						ValueSubstringAllowed: true,
					},
					{
						Directive: "keepalive",
						Value:     "10",
						Upstreams: []string{"default_grpc-backend_8080"},
						File:      "nginx.conf",
					},
					{
						Directive: "keepalive_requests",
						Value:     "3",
						Upstreams: []string{"default_grpc-backend_8080"},
						File:      "nginx.conf",
					},
					{
						Directive: "keepalive_time",
						Value:     "10s",
						Upstreams: []string{"default_grpc-backend_8080"},
						File:      "nginx.conf",
					},
					{
						Directive: "keepalive_timeout",
						Value:     "50s",
						Upstreams: []string{"default_grpc-backend_8080"},
						File:      "nginx.conf",
					},
				}),
			)
		})
	})

	When("When multiple UpstreamSettings Policy target the same service", func() {
		Specify("configuring distinct settings then directives are merged", func() {
			files := []string{"upstream-settings-policy/valid-merge-usps.yaml"}
			Expect(resourceManager.ApplyFromFiles(files, namespace)).To(Succeed())

			nsname := types.NamespacedName{Name: "coffee-svc-usp-1", Namespace: namespace}
			Expect(waitForUSPolicyStatus(nsname, metav1.ConditionTrue, v1alpha2.PolicyReasonTargetNotFound)).To(Succeed())

			nsname = types.NamespacedName{Name: "coffee-svc-usp-2", Namespace: namespace}
			Expect(waitForUSPolicyStatus(nsname, metav1.ConditionTrue, v1alpha2.PolicyReasonTargetNotFound)).To(Succeed())

			Context("verify working traffic", func() {
				It("should return a 200 response for HTTPRoute", func() {
					port := 80
					if portFwdPort != 0 {
						port = portFwdPort
					}
					baseURL := fmt.Sprintf("http://cafe.example.com:%d%s", port, "/coffee")

					Eventually(
						func() error {
							return expectRequestToSucceed(baseURL, address, "URI: /coffee")
						}).
						WithTimeout(timeoutConfig.RequestTimeout).
						WithPolling(500 * time.Millisecond).
						Should(Succeed())
				})
			})

			Context("nginx directives", func() {
				var conf *framework.Payload

				BeforeAll(func() {
					podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
					Expect(err).ToNot(HaveOccurred())
					Expect(podNames).To(HaveLen(1))

					ngfPodName := podNames[0]

					conf, err = resourceManager.GetNginxConfig(ngfPodName, ngfNamespace)
					Expect(err).ToNot(HaveOccurred())
				})

				// TODO: important
				// The directive file and field value need to be updated based on the
				// implementation of the UpstreamSettingsPolicy and how they are specified in the config files.
				DescribeTable("are set properly for",
					func(expCfgs []framework.ExpectedNginxField) {
						for _, expCfg := range expCfgs {
							Expect(framework.ValidateNginxFieldExists(conf, expCfg)).To(Succeed())
						}
					},
					Entry("HTTPRoute", []framework.ExpectedNginxField{
						{
							Directive:             "zone default_coffee_80",
							Value:                 "64k",
							Upstreams:             []string{"default_coffee_80"},
							File:                  "nginx.conf",
							ValueSubstringAllowed: true,
						},
						{
							Directive: "keepalive",
							Value:     "100",
							Upstreams: []string{"default_coffee_80"},
							File:      "nginx.conf",
						},
						{
							Directive: "keepalive_requests",
							Value:     "55",
							Upstreams: []string{"default_coffee_80"},
							File:      "nginx.conf",
						},
						{
							Directive: "keepalive_time",
							Value:     "1m",
							Upstreams: []string{"default_coffee_80"},
							File:      "nginx.conf",
						},
						{
							Directive: "keepalive_timeout",
							Value:     "5h",
							Upstreams: []string{"default_coffee_80"},
							File:      "nginx.conf",
						},
					}),
				)
			})

			Expect(resourceManager.DeleteFromFiles(files, namespace)).To(Succeed())
		})
		Specify("configuring overlapping settings, the policy created first wins", func() {
			files := []string{"upstream-settings-policy/valid-usps-first-wins.yaml"}
			Expect(resourceManager.ApplyFromFiles(files, namespace)).To(Succeed())

			nsname := types.NamespacedName{Name: "coffee-svc-usp-1", Namespace: namespace}
			Expect(waitForUSPolicyStatus(nsname, metav1.ConditionTrue, v1alpha2.PolicyReasonTargetNotFound)).To(Succeed())

			nsname = types.NamespacedName{Name: "coffee-svc-usp-2", Namespace: namespace}
			Expect(waitForUSPolicyStatus(nsname, metav1.ConditionTrue, v1alpha2.PolicyReasonTargetNotFound)).To(Succeed())

			Context("verify working traffic", func() {
				It("should return a 200 response for HTTPRoute", func() {
					port := 80
					if portFwdPort != 0 {
						port = portFwdPort
					}
					baseURL := fmt.Sprintf("http://cafe.example.com:%d%s", port, "/coffee")

					Eventually(
						func() error {
							return expectRequestToSucceed(baseURL, address, "URI: /coffee")
						}).
						WithTimeout(timeoutConfig.RequestTimeout).
						WithPolling(500 * time.Millisecond).
						Should(Succeed())
				})
			})

			Context("nginx directives", func() {
				var conf *framework.Payload

				BeforeAll(func() {
					podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
					Expect(err).ToNot(HaveOccurred())
					Expect(podNames).To(HaveLen(1))

					ngfPodName := podNames[0]

					conf, err = resourceManager.GetNginxConfig(ngfPodName, ngfNamespace)
					Expect(err).ToNot(HaveOccurred())
				})

				// TODO: important
				// The directive file and field value need to be updated based on the
				// implementation of the UpstreamSettingsPolicy and how they are specified in the config files.
				DescribeTable("are set properly for",
					func(expCfgs []framework.ExpectedNginxField) {
						for _, expCfg := range expCfgs {
							Expect(framework.ValidateNginxFieldExists(conf, expCfg)).To(Succeed())
						}
					},
					Entry("HTTPRoute", []framework.ExpectedNginxField{
						{
							Directive:             "zone default_coffee_80",
							Value:                 "128k",
							Upstreams:             []string{"default_coffee_80"},
							File:                  "nginx.conf",
							ValueSubstringAllowed: true,
						},
					}),
				)
			})
			Expect(resourceManager.DeleteFromFiles(files, namespace)).To(Succeed())
		})
	})

	When("UpstreamSettings Policy is invalid", func() {
		Specify("if service mentioned in the target ref is not present", func() {
			files := []string{"upstream-settings-policy/invalid-usps.yaml"}

			Expect(resourceManager.ApplyFromFiles(files, namespace)).To(Succeed())

			nsname := types.NamespacedName{Name: "soda-svc-usp", Namespace: namespace}
			Expect(waitForUSPolicyStatus(nsname, metav1.ConditionFalse, v1alpha2.PolicyReasonTargetNotFound)).To(Succeed())

			Expect(resourceManager.DeleteFromFiles(files, namespace)).To(Succeed())
		})
	})
})

func waitForUSPolicyStatus(
	usPolicyNsNames types.NamespacedName,
	condStatus metav1.ConditionStatus,
	condReason v1alpha2.PolicyConditionReason,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetStatusTimeout)
	defer cancel()

	GinkgoWriter.Printf(
		"Waiting for UpstreamSettings Policy %q to have the condition Accepted/True/Accepted\n",
		usPolicyNsNames,
	)

	return wait.PollUntilContextCancel(
		ctx,
		500*time.Millisecond,
		true, /* poll immediately */
		func(ctx context.Context) (bool, error) {
			var usPolicy ngfAPI.UpstreamSettingsPolicy
			var err error

			if err := k8sClient.Get(ctx, usPolicyNsNames, &usPolicy); err != nil {
				return false, err
			}

			if len(usPolicy.Status.Ancestors) == 0 {
				GinkgoWriter.Printf("UpstreamSettingsPolicy %q does not have an ancestor status yet\n", usPolicy)

				return false, nil
			}

			if len(usPolicy.Status.Ancestors) != 1 {
				return false, fmt.Errorf("policy has %d ancestors, expected 1", len(usPolicy.Status.Ancestors))
			}

			ancestors := usPolicy.Status.Ancestors

			for i, ancestor := range ancestors {
				if err := ancestorMustEqualTargetRef(ancestor, usPolicy.GetTargetRefs()[i], usPolicy.Namespace); err != nil {
					return false, err
				}

				err = ancestorStatusMustHaveAcceptedCondition(ancestor, condStatus, condReason)
			}
			return err == nil, err
		},
	)
}
