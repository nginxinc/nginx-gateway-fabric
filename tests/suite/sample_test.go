package suite

import (
	"fmt"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Basic test example", func() {
	files := []string{
		"hello/hello.yaml",
		"hello/gateway.yaml",
		"hello/route.yaml",
	}
	ns := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "hello",
		},
	}

	BeforeEach(func() {
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsReady(k8sClient, ns.Name)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
	})

	It("sends traffic", func() {
		url := fmt.Sprintf("http://hello.example.com:%s/hello", strconv.Itoa(portFwdPort))
		body, err := framework.GET(url, timeoutConfig.RequestTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(body).To(ContainSubstring("URI: /hello"))
	})
})
