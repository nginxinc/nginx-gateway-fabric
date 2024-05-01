package suite

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	crClient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	collectorNamespace        = "collector"
	collectorChartReleaseName = "otel-collector"
	// FIXME(pleshakov): Find a automated way to keep the version updated here similar to dependabot.
	// https://github.com/nginxinc/nginx-gateway-fabric/issues/1665
	collectorChartVersion = "0.73.1"
)

var _ = Describe("Telemetry test with OTel collector", Label("telemetry"), func() {
	BeforeEach(func() {
		// Because NGF reports telemetry on start, we need to install the collector first.

		// Install collector
		output, err := installCollector()
		Expect(err).ToNot(HaveOccurred(), string(output))

		// Install NGF
		// Note: the BeforeSuite call doesn't install NGF for 'telemetry' label

		setup(
			getDefaultSetupCfg(),
			"--set", "nginxGateway.productTelemetry.enable=true",
		)
	})

	AfterEach(func() {
		output, err := uninstallCollector()
		Expect(err).ToNot(HaveOccurred(), string(output))
	})

	It("sends telemetry", func() {
		names, err := resourceManager.GetPodNames(
			collectorNamespace,
			crClient.MatchingLabels{
				"app.kubernetes.io/name": "opentelemetry-collector",
			},
		)

		Expect(err).ToNot(HaveOccurred())
		Expect(names).To(HaveLen(1))

		name := names[0]

		// We assert that all data points were sent
		// For some data points, as a sanity check, we assert on sent values.

		info, err := resourceManager.GetClusterInfo()
		Expect(err).ToNot(HaveOccurred())

		ngfDeployment, err := resourceManager.GetNGFDeployment(ngfNamespace, releaseName)
		Expect(err).ToNot(HaveOccurred())

		matchFirstExpectedLine := func() bool {
			logs, err := resourceManager.GetPodLogs(collectorNamespace, name, &core.PodLogOptions{})
			Expect(err).ToNot(HaveOccurred())
			return strings.Contains(logs, "dataType: Str(ngf-product-telemetry)")
		}

		// Wait until the collector has received the telemetry data
		Eventually(matchFirstExpectedLine, "30s", "5s").Should(BeTrue())

		logs, err := resourceManager.GetPodLogs(collectorNamespace, name, &core.PodLogOptions{})
		Expect(err).ToNot(HaveOccurred())

		assertConsecutiveLinesInLogs(
			logs,
			[]string{
				"ImageSource:",
				"ProjectName: Str(NGF)",
				"ProjectVersion:",
				"ProjectArchitecture:",
				fmt.Sprintf("ClusterID: Str(%s)", info.ID),
				"ClusterVersion:",
				"ClusterPlatform:",
				fmt.Sprintf("InstallationID: Str(%s)", ngfDeployment.UID),
				fmt.Sprintf("ClusterNodeCount: Int(%d)", info.NodeCount),
				"FlagNames: Slice",
				"FlagValues: Slice",
				"GatewayCount: Int(0)",
				"GatewayClassCount: Int(1)",
				"HTTPRouteCount: Int(0)",
				"SecretCount: Int(0)",
				"ServiceCount: Int(0)",
				"EndpointCount: Int(0)",
				"GRPCRouteCount: Int(0)",
				"BackendTLSPolicyCount: Int(0)",
				"NGFReplicaCount: Int(1)",
			},
		)
	})
})

func installCollector() ([]byte, error) {
	repoAddArgs := []string{
		"repo",
		"add",
		"open-telemetry",
		"https://open-telemetry.github.io/opentelemetry-helm-charts",
	}

	if output, err := exec.Command("helm", repoAddArgs...).CombinedOutput(); err != nil {
		return output, err
	}

	args := []string{
		"install",
		collectorChartReleaseName,
		"open-telemetry/opentelemetry-collector",
		"--create-namespace",
		"--namespace", collectorNamespace,
		"--version", collectorChartVersion,
		"-f", "manifests/telemetry/collector-values.yaml",
		"--wait",
	}

	return exec.Command("helm", args...).CombinedOutput()
}

func uninstallCollector() ([]byte, error) {
	args := []string{
		"uninstall", collectorChartReleaseName,
		"--namespace", collectorNamespace,
	}

	output, err := exec.Command("helm", args...).CombinedOutput()
	if err != nil {
		return output, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	err = k8sClient.Delete(ctx, &core.Namespace{ObjectMeta: metav1.ObjectMeta{Name: collectorNamespace}})
	if err != nil && !apierrors.IsNotFound(err) {
		return nil, err
	}

	return nil, resourceManager.DeleteNamespace(collectorNamespace)
}

func assertConsecutiveLinesInLogs(logs string, expectedLines []string) {
	lines := strings.Split(logs, "\n")

	// find first expected line in lines

	i := 0

	for ; i < len(lines); i++ {
		if strings.Contains(lines[i], expectedLines[0]) {
			i++
			break
		}
	}

	if i == len(lines) {
		Fail(fmt.Sprintf("Expected first line not found: %s, \n%s", expectedLines[0], logs))
	}

	linesLeft := len(lines) - i
	expectedLinesLeft := len(expectedLines) - 1

	if linesLeft < expectedLinesLeft {
		format := "Not enough lines remains in the logs, expected %d, got %d\n%s"
		Fail(fmt.Sprintf(format, linesLeft, expectedLinesLeft, logs))
	}

	for j := 1; j < len(expectedLines); j++ {
		Expect(lines[i]).To(ContainSubstring(expectedLines[j]))
		i++
	}
}
