package suite

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var (
	files = []string{
		"ip-family/gateway.yaml",
		"ip-family/cafe.yaml",
		"ip-family/cafe-routes.yaml",
	}

	nginxProxyFile = "ip-family/nginxproxy.yaml"

	namespace      = "ip-family"
	nginxProxyName = "nginx-proxy"
	ns             core.Namespace

	baseHTTPURL = "http://cafe.example.com"
	teaURL      = baseHTTPURL + "/tea"
	coffeeURL   = baseHTTPURL + "/coffee"
)

var _ = Describe("IPFamily", Ordered, Label("functional", "ip-family"), func() {
	BeforeEach(func() {
		ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: namespace,
			},
		}

		if portFwdPort != 0 {
			coffeeURL = fmt.Sprintf("%s:%d/coffee", baseHTTPURL, portFwdPort)
			teaURL = fmt.Sprintf("%s:%d/tea", baseHTTPURL, portFwdPort)
		}

		Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles([]string{nginxProxyFile}, ns.Name)).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.DeleteFromFiles([]string{nginxProxyFile}, ns.Name)).To(Succeed())
		Expect(resourceManager.DeleteNamespace(ns.Name)).To(Succeed())

		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.CreateTimeout)
		defer cancel()

		key := types.NamespacedName{Name: gatewayClassName}
		var gwClass v1.GatewayClass
		Expect(k8sClient.Get(ctx, key, &gwClass)).To(Succeed())

		gwClass.Spec.ParametersRef = nil
		Expect(k8sClient.Update(ctx, &gwClass)).To(Succeed())
	})

	updateGatewayClass := func() error {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.CreateTimeout)
		defer cancel()

		key := types.NamespacedName{Name: gatewayClassName}
		var gwClass v1.GatewayClass
		if err := k8sClient.Get(ctx, key, &gwClass); err != nil {
			return err
		}

		gwClass.Spec.ParametersRef = &v1.ParametersReference{
			Group: ngfAPI.GroupName,
			Kind:  v1.Kind("NginxProxy"),
			Name:  nginxProxyName,
		}

		return k8sClient.Update(ctx, &gwClass)
	}

	updateNginxProxy := func(ipFamily ngfAPI.IPFamilyType) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.CreateTimeout)
		defer cancel()

		key := types.NamespacedName{Name: nginxProxyName, Namespace: ns.Name}
		var nginxProxy ngfAPI.NginxProxy
		if err := k8sClient.Get(ctx, key, &nginxProxy); err != nil {
			return err
		}

		nginxProxy.Spec.IPFamily = ipFamilyTypePtr(ipFamily)
		return k8sClient.Update(ctx, &nginxProxy)
	}

	Describe("NGF is configured with Dual, IPv4 and IPv6 IPFamily", func() {
		It("Successfully configures coffee with IPv4 and tea with IPv6, NGF installed with Dual IPFamily", func() {
			Eventually(
				func() error {
					return checkWorkingTraffic(teaURL, address, true)
				}).
				WithTimeout(timeoutConfig.RequestTimeout).
				WithPolling(2 * time.Second).
				Should(Succeed())

			Eventually(
				func() error {
					return checkWorkingTraffic(coffeeURL, address, false)
				}).
				WithTimeout(timeoutConfig.RequestTimeout).
				WithPolling(2 * time.Second).
				Should(Succeed())
		})
		It("Successfully configures coffee route(IPv4), "+
			"tea route(IPv6) has accepted condition set to InvalidIPFamily", func() {
			Expect(updateGatewayClass()).Should(Succeed())
			Expect(updateNginxProxy(ngfAPI.IPv4)).Should(Succeed())

			Eventually(
				func() error {
					return checkWorkingTraffic(coffeeURL, address, false)
				}).
				WithTimeout(timeoutConfig.RequestTimeout).
				WithPolling(2 * time.Second).
				Should(Succeed())

			Eventually(
				func() error {
					return verifyRequestFailureAndRouteConditionToBeInvalidIPFamily(
						teaURL,
						address,
						"tea",
						ns.Name,
						"Service configured with IPv6 family but NginxProxy is configured with IPv4")
				}).
				WithTimeout(timeoutConfig.RequestTimeout).
				WithPolling(2 * time.Second).
				Should(Succeed())
		})
		It("Successfully configures tea route(IPv6), "+
			"coffee route(IPv4) has accepted condition set to InvalidIPFamily", func() {
			Expect(updateGatewayClass()).Should(Succeed())
			Expect(updateNginxProxy(ngfAPI.IPv6)).Should(Succeed())

			Eventually(
				func() error {
					return checkWorkingTraffic(teaURL, address, true)
				}).
				WithTimeout(timeoutConfig.RequestTimeout).
				WithPolling(2 * time.Second).
				Should(Succeed())

			Eventually(
				func() error {
					return verifyRequestFailureAndRouteConditionToBeInvalidIPFamily(
						coffeeURL,
						address,
						"coffee",
						ns.Name,
						"Service configured with IPv4 family but NginxProxy is configured with IPv6")
				}).
				WithTimeout(timeoutConfig.RequestTimeout * 2).
				WithPolling(2 * time.Second).
				Should(Succeed())
		})
	})
})

func ipFamilyTypePtr(t ngfAPI.IPFamilyType) *ngfAPI.IPFamilyType {
	return &t
}

func verifyRequestFailureAndRouteConditionToBeInvalidIPFamily(
	appURL,
	address,
	routeName,
	namespace,
	expectedErrMessage string,
) error {
	err := verifyHTTPRouteConditionToBeInvalidIPFamily(routeName, namespace, expectedErrMessage)
	if err != nil {
		return err
	}

	err = expectRequestFailureInternalError(appURL, address)
	if err != nil {
		return err
	}

	return nil
}

func expectRequestFailureInternalError(appURL, address string) error {
	status, body, err := framework.Get(appURL, address, timeoutConfig.RequestTimeout)
	if err != nil {
		return fmt.Errorf("error while sending request: %s", err.Error())
	}

	if status != http.StatusInternalServerError {
		return errors.New("expected http status to be 500")
	}

	if body != "" && !strings.Contains(body, "500 Internal Server Error") {
		return fmt.Errorf("expected response body to have Internal Server Error, instead received: %s", body)
	}

	return nil
}

func checkWorkingTraffic(url, address string, ipv6Expected bool) error {
	status, body, err := framework.Get(url, address, timeoutConfig.RequestTimeout)
	if err != nil {
		return err
	}

	if status != http.StatusOK {
		return errors.New("http response status is not 200")
	}

	err = verifyIPType(body, ipv6Expected)
	return err
}

func verifyIPType(rspBody string, ipv6Expected bool) error {
	lines := strings.Split(rspBody, "\n")
	dataMap := make(map[string]string)

	for _, line := range lines {
		parts := strings.Split(line, ": ")
		if len(parts) == 2 {
			dataMap[parts[0]] = parts[1]
		}
	}

	ip := dataMap["Server address"]
	if ip == "" {
		return errors.New("server address not found in response body")
	}

	if ipv6Expected && isIPv4(ip) {
		return errors.New("expected IPv6 address, got IPv4")
	}
	if !ipv6Expected && !isIPv4(ip) {
		return errors.New("expected IPv4 address, got IPv6")
	}
	return nil
}

func isIPv4(str string) bool {
	ip := net.ParseIP(str)
	if ip != nil && ip.To4() == nil {
		return false
	}

	return true
}

func verifyHTTPRouteConditionToBeInvalidIPFamily(httpRouteName, namespace, expectedErrMessage string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetStatusTimeout)
	defer cancel()
	route := &v1.HTTPRoute{}

	return wait.PollUntilContextCancel(
		ctx,
		200*time.Millisecond,
		true,
		func(ctx context.Context) (bool, error) {
			httpRouteNsName := types.NamespacedName{Name: httpRouteName, Namespace: namespace}
			if err := k8sClient.Get(ctx, httpRouteNsName, route); err != nil {
				return false, fmt.Errorf("error getting HTTPRoute %s: %w", httpRouteName, err)
			}

			if len(route.Status.Parents) == 0 {
				GinkgoWriter.Printf("HTTPRoute %q does not status yet \n", httpRouteName)
				return false, nil
			}

			if len(route.Status.Parents) >= 1 && len(route.Status.Parents[0].Conditions) == 0 {
				GinkgoWriter.Printf("HTTPRoute %q does not have conditions yet \n", httpRouteName)
				return false, nil
			}

			for _, parent := range route.Status.Parents {
				for _, condition := range parent.Conditions {
					fmt.Println("condition.Type: ", condition.Type, condition.Message, condition.Status)
					if condition.Type == string(v1.RouteConditionResolvedRefs) &&
						condition.Status == metav1.ConditionFalse &&
						condition.Message == expectedErrMessage {
						return true, nil
					}
				}
			}

			return false, nil
		},
	)
}
