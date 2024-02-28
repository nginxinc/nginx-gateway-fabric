package suite

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Basic test example", Ordered, Label("functional"), func() {
	var outFile *os.File
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

	BeforeAll(func() {
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

		resultsDir, err := framework.CreateResultsDir("functional", version)
		Expect(err).ToNot(HaveOccurred())

		filename := filepath.Join(resultsDir, fmt.Sprintf("%s.md", version))
		outFile, err = framework.CreateResultsFile(filename)
		Expect(err).ToNot(HaveOccurred())
		Expect(framework.WriteSystemInfoToFile(outFile, clusterInfo, *plusEnabled)).To(Succeed())
	})

	AfterAll(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
		outFile.Close()
	})

	It("sends traffic", func() {
		url := "http://hello.example.com/hello"
		if portFwdPort != 0 {
			url = fmt.Sprintf("http://hello.example.com:%s/hello", strconv.Itoa(portFwdPort))
		}
		status, body, err := framework.Get(url, address, timeoutConfig.RequestTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(status).To(Equal(http.StatusOK))
		Expect(body).To(ContainSubstring("URI: /hello"))
	})
})
