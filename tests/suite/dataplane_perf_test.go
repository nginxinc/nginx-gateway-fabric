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
	vegeta "github.com/tsenart/vegeta/v12/lib"
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
	parentDir := filepath.Dir(pwd)
	var port string
	var resultsFile string
	var addr string

	BeforeEach(func() {
		resultPath := fmt.Sprintf("%v/results/dp-perf/%v/", parentDir, version)
		Expect(os.MkdirAll(resultPath, 0o777)).To(Succeed())
		resultsFile = fmt.Sprintf("%v/%v.md", resultPath, version)
		if portFwdPort != 0 {
			port = fmt.Sprintf(":%s", strconv.Itoa(portFwdPort))
		}
		addr = fmt.Sprintf("http://%s%s", address, port)
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
	})

	AfterEach(func() {
		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
	})

	It("runs the test and writes a results file", func() {
		Expect(runDPPTestOutputToFile(resultsFile, clusterInfo, addr)).To(Succeed())
		Expect(resultsFile).To(BeAnExistingFile())
	})
})

func runDPPTestOutputToFile(outputPath string, ci framework.ClusterInfo, addr string) error {
	//nolint:gosec
	resFile, err := os.OpenFile(outputPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o777)
	if err != nil {
		return err
	}
	defer resFile.Close()

	if err = framework.WriteSystemInfoToFile(resFile, ci); err != nil {
		return fmt.Errorf("Could not write system info to file %v: %w", outputPath, err)
	}

	if err = runDPTest(resFile, addr); err != nil {
		return fmt.Errorf("Error running test: %w", err)
	}

	return nil
}

func runDPTest(outFile *os.File, addr string) error {
	targetURL := "http://cafe.example.com"

	tests := []struct {
		target      vegeta.Target
		description string
	}{
		{
			description: "Running latte path based routing",
			target: vegeta.Target{
				Method: "GET",
				URL:    fmt.Sprintf("%s%s", targetURL, "/latte"),
			},
		},
		{
			description: "Running coffee header based routing",
			target: vegeta.Target{
				Method: "GET",
				URL:    fmt.Sprintf("%s%s", targetURL, "/coffee"),
				Header: http.Header{"version": []string{"v2"}},
			},
		},
		{
			description: "Running coffee query based routing",
			target: vegeta.Target{
				Method: "GET",
				URL:    fmt.Sprintf("%s%s", targetURL, "/coffee?TEST=v2"),
			},
		},
		{
			description: "Running tea GET method based routing",
			target: vegeta.Target{
				Method: "GET",
				URL:    fmt.Sprintf("%s%s", targetURL, "/tea"),
			},
		},
		{
			description: "Running tea POST method based routing",
			target: vegeta.Target{
				Method: "POST",
				URL:    fmt.Sprintf("%s%s", targetURL, "/tea"),
			},
		},
	}

	for i, test := range tests {
		text := fmt.Sprintf("\n## Test%d: %s\n\n```text\n", i+1, test.description)
		if _, err := fmt.Fprint(outFile, text); err != nil {
			return err
		}
		if err := framework.RunLoadTest(
			[]vegeta.Target{test.target},
			1000,
			30*time.Second,
			test.description,
			outFile,
			addr,
		); err != nil {
			return err
		}
		if _, err := fmt.Fprint(outFile, "```\n"); err != nil {
			return err
		}
	}
	return nil
}
