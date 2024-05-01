package suite

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	gatewayv1alpha2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/conditions"
	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Tracing", Label("functional"), func() {
	files := []string{
		"hello-world/apps.yaml",
		"hello-world/gateway.yaml",
		"hello-world/routes.yaml",
	}
	var ns core.Namespace

	var collectorPodName, helloURL, worldURL, helloworldURL string

	BeforeEach(func() {
		ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "helloworld",
			},
		}

		output, err := installCollector()
		Expect(err).ToNot(HaveOccurred(), string(output))

		collectorPodNames, err := resourceManager.GetPodNames(
			collectorNamespace,
			client.MatchingLabels{
				"app.kubernetes.io/name": "opentelemetry-collector",
			},
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(collectorPodNames).To(HaveLen(1))

		collectorPodName = collectorPodNames[0]

		Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		url := "http://foo.example.com"
		helloURL = url + "/hello"
		worldURL = url + "/world"
		helloworldURL = url + "/helloworld"
		if portFwdPort != 0 {
			helloURL = fmt.Sprintf("%s:%s/hello", url, strconv.Itoa(portFwdPort))
			worldURL = fmt.Sprintf("%s:%s/world", url, strconv.Itoa(portFwdPort))
			helloworldURL = fmt.Sprintf("%s:%s/helloworld", url, strconv.Itoa(portFwdPort))
		}
	})

	AfterEach(func() {
		output, err := uninstallCollector()
		Expect(err).ToNot(HaveOccurred(), string(output))

		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.DeleteNamespace(ns.Name)).To(Succeed())
	})

	updateGatewayClass := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.CreateTimeout)
		defer cancel()

		key := types.NamespacedName{Name: "nginx"}
		var gwClass gatewayv1.GatewayClass
		if err := k8sClient.Get(ctx, key, &gwClass); err != nil {
			return err
		}

		gwClass.Spec.ParametersRef = &gatewayv1.ParametersReference{
			Group: ngfAPI.GroupName,
			Kind:  gatewayv1.Kind("NginxProxy"),
			Name:  "nginx-proxy",
		}

		return k8sClient.Update(ctx, &gwClass)
	}

	sendTraceRequests := func(ctx context.Context, url string, count int) {
		for range count {
			status, _, err := framework.GetWithRetry(ctx, url, address, timeoutConfig.RequestTimeout)
			Expect(err).ToNot(HaveOccurred())
			Expect(status).To(Equal(http.StatusOK))
		}
	}

	It("sends tracing spans for one policy attached to one route", func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		sendTraceRequests(ctx, helloURL, 5)

		// verify that no traces exist yet
		logs, err := resourceManager.GetPodLogs(collectorNamespace, collectorPodName, &core.PodLogOptions{})
		Expect(err).ToNot(HaveOccurred())
		Expect(logs).ToNot(ContainSubstring("service.name: Str(ngf:helloworld:gateway:my-test-svc)"))

		// install tracing configuration
		traceFiles := []string{
			"tracing/nginxproxy.yaml",
			"tracing/policy-single.yaml",
		}
		Expect(resourceManager.ApplyFromFiles(traceFiles, ns.Name)).To(Succeed())
		Expect(updateGatewayClass()).To(Succeed())

		Eventually(
			func() error {
				return verifyGatewayClassResolvedRefs()
			}).
			WithTimeout(timeoutConfig.GetTimeout).
			WithPolling(500 * time.Millisecond).
			Should(Succeed())

		Eventually(
			func() error {
				return verifyPolicyStatus()
			}).
			WithTimeout(timeoutConfig.GetTimeout).
			WithPolling(500 * time.Millisecond).
			Should(Succeed())

		// send traffic and verify that traces exist for hello app
		findTraces := func() bool {
			sendTraceRequests(ctx, helloURL, 25)
			sendTraceRequests(ctx, worldURL, 25)
			sendTraceRequests(ctx, helloworldURL, 25)

			logs, err := resourceManager.GetPodLogs(collectorNamespace, collectorPodName, &core.PodLogOptions{})
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(logs, "service.name: Str(ngf:helloworld:gateway:my-test-svc)")
		}

		// wait for expected first line to show up
		Eventually(findTraces, "1m", "5s").Should(BeTrue())

		logs, err = resourceManager.GetPodLogs(collectorNamespace, collectorPodName, &core.PodLogOptions{})
		Expect(err).ToNot(HaveOccurred())

		Expect(logs).To(ContainSubstring("http.method: Str(GET)"))
		Expect(logs).To(ContainSubstring("http.target: Str(/hello)"))
		Expect(logs).To(ContainSubstring("testkey1: Str(testval1)"))
		Expect(logs).To(ContainSubstring("testkey2: Str(testval2)"))

		// verify traces don't exist for other apps
		Expect(logs).ToNot(ContainSubstring("http.target: Str(/world)"))
		Expect(logs).ToNot(ContainSubstring("http.target: Str(/helloworld)"))
	})

	It("sends tracing spans for one policy attached to multiple routes", func() {
		// install tracing configuration
		traceFiles := []string{
			"tracing/nginxproxy.yaml",
			"tracing/policy-multiple.yaml",
		}
		Expect(resourceManager.ApplyFromFiles(traceFiles, ns.Name)).To(Succeed())
		Expect(updateGatewayClass()).To(Succeed())

		Eventually(
			func() error {
				return verifyGatewayClassResolvedRefs()
			}).
			WithTimeout(timeoutConfig.GetTimeout).
			WithPolling(500 * time.Millisecond).
			Should(Succeed())

		Eventually(
			func() error {
				return verifyPolicyStatus()
			}).
			WithTimeout(timeoutConfig.GetTimeout).
			WithPolling(500 * time.Millisecond).
			Should(Succeed())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// send traffic and verify that traces exist for hello app
		findTraces := func() bool {
			sendTraceRequests(ctx, helloURL, 25)
			sendTraceRequests(ctx, worldURL, 25)
			sendTraceRequests(ctx, helloworldURL, 25)

			logs, err := resourceManager.GetPodLogs(collectorNamespace, collectorPodName, &core.PodLogOptions{})
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(logs, "service.name: Str(ngf:helloworld:gateway:my-test-svc)")
		}

		// wait for expected first line to show up
		Eventually(findTraces, "1m", "5s").Should(BeTrue())

		logs, err := resourceManager.GetPodLogs(collectorNamespace, collectorPodName, &core.PodLogOptions{})
		Expect(err).ToNot(HaveOccurred())

		Expect(logs).To(ContainSubstring("http.method: Str(GET)"))
		Expect(logs).To(ContainSubstring("http.target: Str(/hello)"))
		Expect(logs).To(ContainSubstring("http.target: Str(/world)"))
		Expect(logs).To(ContainSubstring("testkey1: Str(testval1)"))
		Expect(logs).To(ContainSubstring("testkey2: Str(testval2)"))

		// verify traces don't exist for helloworld apps
		Expect(logs).ToNot(ContainSubstring("http.target: Str(/helloworld)"))
	})
})

func verifyGatewayClassResolvedRefs() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var gc v1.GatewayClass
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: gatewayClassName}, &gc); err != nil {
		return err
	}

	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(conditions.GatewayClassResolvedRefs) && cond.Status == metav1.ConditionTrue {
			return nil
		}
	}

	return errors.New("ResolvedRefs status not set to true on GatewayClass")
}

func verifyPolicyStatus() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	var pol ngfAPI.ObservabilityPolicy
	key := types.NamespacedName{Name: "test-observability-policy", Namespace: "helloworld"}
	if err := k8sClient.Get(ctx, key, &pol); err != nil {
		return err
	}

	var count int
	for _, ancestor := range pol.Status.Ancestors {
		for _, cond := range ancestor.Conditions {
			if cond.Type == string(gatewayv1alpha2.PolicyConditionAccepted) && cond.Status == metav1.ConditionTrue {
				count++
			}
		}
	}

	if count != len(pol.Status.Ancestors) {
		return errors.New("Policy not accepted")
	}

	return nil
}
