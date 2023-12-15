package suite

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Dataplane performance", Ordered, Label("performance"), func() {
	files := []string{
		"dp-perf/coffee.yaml",
		"dp-perf/gateway.yaml",
		"dp-perf/cafe-routes.yaml",
	}
	ns := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "dp-perf",
		},
	}

	var port, parentDir, resultsFilePath, addr string
	targetURL := "http://cafe.example.com"
	var outFile *os.File

	t1 := framework.Target{
		Method: "GET",
		URL:    fmt.Sprintf("%s%s", targetURL, "/latte"),
	}
	t2 := framework.Target{
		Method: "GET",
		URL:    fmt.Sprintf("%s%s", targetURL, "/coffee"),
		Header: http.Header{"version": []string{"v2"}},
	}
	t3 := framework.Target{
		Method: "GET",
		URL:    fmt.Sprintf("%s%s", targetURL, "/coffee?TEST=v2"),
	}
	t4 := framework.Target{
		Method: "GET",
		URL:    fmt.Sprintf("%s%s", targetURL, "/tea"),
	}
	t5 := framework.Target{
		Method: "POST",
		URL:    fmt.Sprintf("%s%s", targetURL, "/tea"),
	}

	BeforeAll(func() {
		pwd, err := os.Getwd()
		Expect(err).ToNot(HaveOccurred())
		parentDir = filepath.Dir(pwd)
		resultsFilePath = fmt.Sprintf("%v/results/dp-perf/%v/", parentDir, version)
		Expect(os.MkdirAll(resultsFilePath, 0o777)).To(Succeed())
		resultsFilePath = fmt.Sprintf("%v/%v.md", resultsFilePath, version)
		if portFwdPort != 0 {
			port = fmt.Sprintf(":%s", strconv.Itoa(portFwdPort))
		}
		addr = fmt.Sprintf("http://%s%s", address, port)
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
		outFile, err = os.OpenFile(resultsFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o777)
		Expect(err).To(BeNil())
	})

	It("writes the system info to a results file", func() {
		Expect(framework.WriteSystemInfoToFile(outFile, clusterInfo)).To(Succeed())
	})

	DescribeTable("Run each load test",
		func(target framework.Target, description string, counter int) {
			text := fmt.Sprintf("\n## Test%d: %s\n\n```text\n", counter, description)
			_, err := fmt.Fprint(outFile, text)
			Expect(err).To(BeNil())
			Expect(framework.RunLoadTest(
				[]framework.Target{target},
				1000,
				30*time.Second,
				description,
				outFile,
				addr,
			)).To(Succeed())
			_, err = fmt.Fprint(outFile, "```\n")
			Expect(err).To(BeNil())
		},
		Entry("Running latte path based routing", t1, "Running latte path based routing", 1),
		Entry("Running coffee header based routing", t2, "Running coffee header based routing", 2),
		Entry("Running coffee query based routing", t3, "Running coffee query based routing", 3),
		Entry("Running tea GET method based routing", t4, "Running tea GET method based routing", 4),
		Entry("Running tea POST method based routing", t5, "Running tea POST method based routing", 5),
	)

	AfterAll(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
		outFile.Close()
	})
})
