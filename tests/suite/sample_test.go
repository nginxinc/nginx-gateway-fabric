package suite

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Basic test example", Label("functional"), func() {
	files := []string{
		"hello-world/apps.yaml",
		"hello-world/gateway.yaml",
		"hello-world/routes.yaml",
	}

	var ns core.Namespace

	BeforeEach(func() {
		ns = core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "helloworld",
			},
		}

		Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.DeleteNamespace(ns.Name)).To(Succeed())
	})

	It("sends traffic", func() {
		url := "http://foo.example.com/hello"
		if portFwdPort != 0 {
			url = fmt.Sprintf("http://foo.example.com:%s/hello", strconv.Itoa(portFwdPort))
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		status, body, err := framework.GetWithRetry(ctx, url, address, timeoutConfig.RequestTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring("URI: /hello"))
	})
})
