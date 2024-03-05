package suite

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

// Longevity test is an NFR test, but does not include the "nfr" label. It needs to run on its own,
// outside of the scope of the other NFR tests. This is because it's a long-term test whose environment
// shouldn't be torn down.
var _ = Describe("Longevity", Label("longevity-setup", "longevity-teardown"), func() {
	var (
		files = []string{
			"longevity/cafe.yaml",
			"longevity/cafe-secret.yaml",
			"longevity/gateway.yaml",
			"longevity/cafe-routes.yaml",
			"longevity/cronjob.yaml",
		}
		promFile = []string{
			"longevity/prom.yaml",
		}

		ns = &core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "longevity",
			},
		}

		labelFilter = GinkgoLabelFilter()
	)

	BeforeEach(func() {
		if !strings.Contains(labelFilter, "longevity") {
			Skip("skipping longevity test unless 'longevity' label is explicitly defined when running")
		}
	})

	It("sets up the longevity test", Label("longevity-setup"), func() {
		if !strings.Contains(labelFilter, "longevity-setup") {
			Skip("'longevity-setup' label not specified; skipping...")
		}

		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(promFile, ngfNamespace)).To(Succeed())
		Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())
	})

	It("collects results", Label("longevity-teardown"), func() {
		if !strings.Contains(labelFilter, "longevity-teardown") {
			Skip("'longevity-teardown' label not specified; skipping...")
		}

		resultsDir, err := framework.CreateResultsDir("longevity", version)
		Expect(err).ToNot(HaveOccurred())

		filename := filepath.Join(resultsDir, fmt.Sprintf("%s.md", version))
		resultsFile, err := framework.CreateResultsFile(filename)
		Expect(err).ToNot(HaveOccurred())
		defer resultsFile.Close()

		Expect(framework.WriteSystemInfoToFile(resultsFile, clusterInfo, *plusEnabled)).To(Succeed())

		// gather wrk output
		homeDir, err := os.UserHomeDir()
		Expect(err).ToNot(HaveOccurred())

		Expect(framework.WriteContent(resultsFile, "\n## Traffic\n"))
		writeTrafficResults(resultsFile, homeDir, "coffee.txt", "HTTP")
		writeTrafficResults(resultsFile, homeDir, "tea.txt", "HTTPS")

		// gather any error logs
		names, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())

		Expect(framework.WriteContent(resultsFile, "\n## Error Logs\n"))
		writeErrorLogs(resultsFile, names[0], "nginx-gateway")
		writeErrorLogs(resultsFile, names[0], "nginx")

		Expect(resourceManager.DeleteFromFiles(files, ns.Name)).To(Succeed())
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
	})
})

func writeTrafficResults(resultsFile *os.File, homeDir, filename, testname string) {
	file := fmt.Sprintf("%s/%s", homeDir, filename)
	content, err := os.ReadFile(file)
	Expect(err).ToNot(HaveOccurred())

	formattedContent := fmt.Sprintf("%s:\n\n```text\n%s```\n", testname, string(content))
	Expect(framework.WriteContent(resultsFile, formattedContent)).To(Succeed())
}

func writeErrorLogs(resultsFile *os.File, pod, container string) {
	logReq := clientGoClient.CoreV1().Pods(ngfNamespace).GetLogs(pod, &core.PodLogOptions{Container: container})

	ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
	defer cancel()

	logs, err := logReq.Stream(ctx)
	Expect(err).ToNot(HaveOccurred())
	defer logs.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	Expect(err).ToNot(HaveOccurred())

	Expect(framework.WriteContent(resultsFile, fmt.Sprintf("\n### %s\n", container)))

	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		if isError(line) {
			Expect(framework.WriteContent(resultsFile, line)).To(Succeed())
		}
	}
	Expect(scanner.Err()).ToNot(HaveOccurred())
}

func isError(line string) bool {
	return strings.Contains(line, "error") || strings.Contains(line, "warn") || strings.Contains(line, "emerg")
}
