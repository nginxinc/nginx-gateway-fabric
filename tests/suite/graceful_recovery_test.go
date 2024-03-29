package suite

import (
	"fmt"
	"net/http"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Graceful Recovery test", Label("nfr", "graceful-recovery"), func() {
	files := []string{
		"graceful-recovery/cafe.yaml",
		"graceful-recovery/cafe-secret.yaml",
		"graceful-recovery/gateway.yaml",
		"graceful-recovery/cafe-routes.yaml",
	}
	ns := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "graceful-recovery",
		},
	}

	BeforeEach(func() {
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
	})

	It("sends traffic", func() {
		teaURL := "https://cafe.example.com/tea"
		coffeeURL := "http://cafe.example.com/coffee"
		if portFwdPort != 0 {
			teaURL = fmt.Sprintf("https://cafe.example.com:%s/tea", strconv.Itoa(portFwdPort))
			coffeeURL = fmt.Sprintf("http://cafe.example.com:%s/coffee", strconv.Itoa(portFwdPort))
		}
		status, body, err := framework.Get(teaURL, address, timeoutConfig.RequestTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring("URI: /tea"))

		status, body, err = framework.Get(coffeeURL, address, timeoutConfig.RequestTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring("URI: /coffee"))
	})
})
