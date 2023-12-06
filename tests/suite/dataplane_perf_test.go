package suite

import (
	"fmt"
	"os"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Dataplane performance", func() {
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

	pwd, err := os.Getwd()
	Expect(err).ToNot(HaveOccurred())
	script := fmt.Sprintf("%v/scripts/dp-perf/wrk-latency.sh", pwd)
	var port string
	var resultsFile string

	BeforeEach(func() {
		resultPath := fmt.Sprintf("%v/results/dp-perf/%v/", pwd, version)
		Expect(os.MkdirAll(resultPath, 0o777)).To(Succeed())
		resultsFile = fmt.Sprintf("%v/%v.md", resultPath, version)
		if portFwdPort != 0 {
			port = fmt.Sprintf(":%s", strconv.Itoa(portFwdPort))
		}
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
		Expect(framework.AddEntryToHostsFile("cafe.example.com", address)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
		err := framework.RemoveEntryFromHostsFile("cafe.example.com", address)
		if err != nil {
			fmt.Println("Error removing hosts entry - run `make reset-etc-hosts` instead")
		}
	})

	It("runs the script and writes a results file", func() {
		Expect(framework.RunScriptOutputToFile(script, []string{port}, resultsFile, clusterInfo)).To(Succeed())
		Expect(resultsFile).To(BeAnExistingFile())
	})
})
