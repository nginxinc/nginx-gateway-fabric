package suite

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"text/template"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	core "k8s.io/api/core/v1"
	ctlr "sigs.k8s.io/controller-runtime"
	v1 "sigs.k8s.io/gateway-api/apis/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nginxinc/nginx-gateway-fabric/tests/framework"
)

var _ = Describe("Reconfiguration Performance Testing", Ordered, Label("reconfiguration"), func() {
	var (
		scrapeInterval        = 15 * time.Second
		queryRangeStep        = 5 * time.Second
		promInstance          framework.PrometheusInstance
		promPortForwardStopCh = make(chan struct{})

		outFile *os.File
	)

	BeforeAll(func() {
		resultsDir, err := framework.CreateResultsDir("reconfig", version)
		Expect(err).ToNot(HaveOccurred())

		filename := filepath.Join(resultsDir, framework.CreateResultsFilename("md", version, *plusEnabled))
		outFile, err = framework.CreateResultsFile(filename)
		Expect(err).ToNot(HaveOccurred())
		Expect(framework.WriteSystemInfoToFile(outFile, clusterInfo, *plusEnabled)).To(Succeed())

		promCfg := framework.PrometheusConfig{
			ScrapeInterval: scrapeInterval,
		}

		promInstance, err = framework.InstallPrometheus(resourceManager, promCfg)
		Expect(err).ToNot(HaveOccurred())

		k8sConfig := ctlr.GetConfigOrDie()

		if !clusterInfo.IsGKE {
			Expect(promInstance.PortForward(k8sConfig, promPortForwardStopCh)).To(Succeed())
		}
	})

	BeforeEach(func() {
		teardown(releaseName)
	})

	AfterAll(func() {
		teardown(releaseName)
		close(promPortForwardStopCh)
		Expect(framework.UninstallPrometheus(resourceManager)).To(Succeed())

		// might want to call cleanupResources here with 150 or the max resources.
	})

	createUniqueResources := func(resourceCount int, fileName string) error {
		for i := 1; i <= resourceCount; i++ {
			nsName := "namespace" + strconv.Itoa(i)
			// Command to run sed and capture its output
			//nolint:gosec
			sedCmd := exec.Command("sed",
				"-e",
				"s/coffee/coffee"+nsName+"/g",
				"-e",
				"s/tea/tea"+nsName+"/g",
				fileName,
			)
			// Command to apply using kubectl
			kubectlCmd := exec.Command("kubectl", "apply", "-n", nsName, "-f", "-")

			sedOutput, err := sedCmd.Output()
			if err != nil {
				fmt.Println(err.Error() + ": " + string(sedOutput))
				return err
			}
			kubectlCmd.Stdin = bytes.NewReader(sedOutput)

			output, err := kubectlCmd.CombinedOutput()
			if err != nil {
				fmt.Println(err.Error() + ": " + string(output))
				return err
			}
		}
		return nil
	}

	createResourcesGWLast := func(resourceCount int) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
		defer cancel()

		for i := 1; i <= resourceCount; i++ {
			ns := core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "namespace" + strconv.Itoa(i),
				},
			}
			Expect(k8sClient.Create(ctx, &ns)).To(Succeed())
		}

		ns := core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "reconfig",
			},
		}
		Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(
			[]string{
				"reconfig/certificate-ns-and-cafe-secret.yaml",
				"reconfig/reference-grant.yaml",
			},
			ns.Name)).To(Succeed())

		Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe.yaml")).To(Succeed())

		Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe-routes.yaml")).To(Succeed())

		time.Sleep(60 * time.Second)

		Expect(resourceManager.ApplyFromFiles([]string{"reconfig/gateway.yaml"}, ns.Name)).To(Succeed())

		return nil
	}

	createResourcesRoutesLast := func(resourceCount int) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
		defer cancel()

		for i := 1; i <= resourceCount; i++ {
			ns := core.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "namespace" + strconv.Itoa(i),
				},
			}
			Expect(k8sClient.Create(ctx, &ns)).To(Succeed())
		}

		Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe.yaml")).To(Succeed())

		time.Sleep(60 * time.Second)

		ns := core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "reconfig",
			},
		}
		Expect(resourceManager.Apply([]client.Object{&ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(
			[]string{
				"reconfig/certificate-ns-and-cafe-secret.yaml",
				"reconfig/reference-grant.yaml",
				"reconfig/gateway.yaml",
			},
			ns.Name)).To(Succeed())

		Expect(createUniqueResources(resourceCount, "manifests/reconfig/cafe-routes.yaml")).To(Succeed())

		return nil
	}

	checkResourceCreation := func(resourceCount int) error {
		ctx, cancel := context.WithTimeout(context.Background(), timeoutConfig.GetTimeout)
		defer cancel()

		var namespaces core.NamespaceList
		if err := k8sClient.List(ctx, &namespaces); err != nil {
			return fmt.Errorf("error getting namespaces: %w", err)
		}
		Expect(len(namespaces.Items)).To(BeNumerically(">=", resourceCount))

		var routes v1.HTTPRouteList
		if err := k8sClient.List(ctx, &routes); err != nil {
			return fmt.Errorf("error getting HTTPRoutes: %w", err)
		}
		Expect(len(routes.Items)).To(BeNumerically(">=", resourceCount*3))

		var pods core.PodList
		if err := k8sClient.List(ctx, &pods); err != nil {
			return fmt.Errorf("error getting Pods: %w", err)
		}
		Expect(len(pods.Items)).To(BeNumerically(">=", resourceCount*2))

		return nil
	}

	cleanupResources := func(resourceCount int) {
		for i := 1; i <= resourceCount; i++ {
			nsName := "namespace" + strconv.Itoa(i)
			Expect(resourceManager.DeleteNamespace(nsName)).To(Succeed())
		}

		ns := core.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "reconfig",
			},
		}

		Expect(resourceManager.DeleteFromFiles([]string{
			"reconfig/certificate-ns-and-cafe-secret.yaml",
			"reconfig/reference-grant.yaml",
			"reconfig/gateway.yaml",
		}, ns.Name)).To(Succeed())
	}

	runTestWithMetrics := func(
		testName string,
		resourceCount int,
		test func(resourceCount int) error,
		startWithNGFSetup bool,
	) {
		var (
			metricExistTimeout = 2 * time.Minute
			metricExistPolling = 1 * time.Second
			ngfPodName         string
			startTime          time.Time
		)

		getStartTime := func() time.Time { return startTime }
		modifyStartTime := func() { startTime = startTime.Add(500 * time.Millisecond) }

		if startWithNGFSetup {
			setup(getDefaultSetupCfg())

			podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
			Expect(err).ToNot(HaveOccurred())
			Expect(podNames).To(HaveLen(1))
			ngfPodName = podNames[0]
			startTime = time.Now()

			queries := []string{
				fmt.Sprintf(`container_memory_usage_bytes{pod="%s",container="nginx-gateway"}`, ngfPodName),
				fmt.Sprintf(`container_cpu_usage_seconds_total{pod="%s",container="nginx-gateway"}`, ngfPodName),
				// We don't need to check all nginx_gateway_fabric_* metrics, as they are collected at the same time
				fmt.Sprintf(`nginx_gateway_fabric_nginx_reloads_total{pod="%s"}`, ngfPodName),
			}

			for _, q := range queries {
				Eventually(
					framework.CreateMetricExistChecker(
						promInstance,
						q,
						getStartTime,
						modifyStartTime,
					),
				).WithTimeout(metricExistTimeout).WithPolling(metricExistPolling).Should(Succeed())
			}
		} else {
			output, err := framework.InstallGatewayAPI(getDefaultSetupCfg().gwAPIVersion)
			Expect(err).ToNot(HaveOccurred(), string(output))
		}

		Expect(test(resourceCount)).To(Succeed())
		Expect(checkResourceCreation(resourceCount)).To(Succeed())

		if !startWithNGFSetup {
			setup(getDefaultSetupCfg())

			podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
			Expect(err).ToNot(HaveOccurred())
			Expect(podNames).To(HaveLen(1))
			ngfPodName = podNames[0]
			startTime = time.Now()

			// if i do a new instance of NGF each time, I might not need start time and can just do the endtime.
			queries := []string{
				fmt.Sprintf(`container_memory_usage_bytes{pod="%s",container="nginx-gateway"}`, ngfPodName),
				fmt.Sprintf(`container_cpu_usage_seconds_total{pod="%s",container="nginx-gateway"}`, ngfPodName),
				// We don't need to check all nginx_gateway_fabric_* metrics, as they are collected at the same time
				fmt.Sprintf(`nginx_gateway_fabric_nginx_reloads_total{pod="%s"}`, ngfPodName),
			}

			for _, q := range queries {
				Eventually(
					framework.CreateMetricExistChecker(
						promInstance,
						q,
						getStartTime,
						modifyStartTime,
					),
				).WithTimeout(metricExistTimeout).WithPolling(metricExistPolling).Should(Succeed())
			}
		}

		time.Sleep(2 * scrapeInterval)

		endTime := time.Now()

		Eventually(
			framework.CreateEndTimeFinder(
				promInstance,
				fmt.Sprintf(`rate(container_cpu_usage_seconds_total{pod="%s",container="nginx-gateway"}[2m])`, ngfPodName),
				startTime,
				&endTime,
				queryRangeStep,
			),
		).WithTimeout(metricExistTimeout).WithPolling(metricExistPolling).Should(Succeed())

		getEndTime := func() time.Time { return endTime }
		noOpModifier := func() {}

		queries := []string{
			fmt.Sprintf(`container_memory_usage_bytes{pod="%s",container="nginx-gateway"}`, ngfPodName),
			// We don't need to check all nginx_gateway_fabric_* metrics, as they are collected at the same time
			fmt.Sprintf(`nginx_gateway_fabric_nginx_reloads_total{pod="%s"}`, ngfPodName),
		}

		for _, q := range queries {
			Eventually(
				framework.CreateMetricExistChecker(
					promInstance,
					q,
					getEndTime,
					noOpModifier,
				),
			).WithTimeout(metricExistTimeout).WithPolling(metricExistPolling).Should(Succeed())
		}

		reloadCount, err := framework.GetReloadCount(promInstance, ngfPodName)
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(reloadCount)

		reloadAvgTime, err := framework.GetReloadAvgTime(promInstance, ngfPodName)
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(reloadAvgTime)

		reloadBuckets, err := framework.GetReloadBuckets(promInstance, ngfPodName)
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(reloadBuckets)

		eventsCount, err := framework.GetEventsCount(promInstance, ngfPodName)
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(eventsCount)

		eventsAvgTime, err := framework.GetEventsAvgTime(promInstance, ngfPodName)
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(eventsAvgTime)

		eventsBuckets, err := framework.GetEventsBuckets(promInstance, ngfPodName)
		Expect(err).ToNot(HaveOccurred())
		fmt.Println(eventsBuckets)

		results := reconfigTestResults{
			Name:               testName,
			EventsBuckets:      eventsBuckets,
			ReloadBuckets:      reloadBuckets,
			NumResources:       resourceCount,
			NGINXReloads:       int(reloadCount),
			NGINXReloadAvgTime: int(reloadAvgTime),
			EventsCount:        int(eventsCount),
			EventsAvgTime:      int(eventsAvgTime),
		}

		err = writeReconfigResults(outFile, results)
		Expect(err).ToNot(HaveOccurred())

		cleanupResources(30)
	}

	It("test 1", func() {
		runTestWithMetrics("1", 30, createResourcesGWLast, false)
	})

	It("test 2", func() {
		runTestWithMetrics("2", 30, createResourcesRoutesLast, true)
	})

	It("test 3", func() {
		runTestWithMetrics("3", 30, createResourcesGWLast, true)
	})

	//It("test 2", func() {
	//	Expect(createResourcesRoutesLast(30)).To(Succeed())
	//	Expect(checkResourceCreation(30)).To(Succeed())
	//	cleanupResources(30)
	//})
})

type reconfigTestResults struct {
	Name                 string
	EventsBuckets        []framework.Bucket
	ReloadBuckets        []framework.Bucket
	NumResources         int
	TimeToReadyTotal     int
	TimeToReadyAvgSingle int
	NGINXReloads         int
	NGINXReloadAvgTime   int
	EventsCount          int
	EventsAvgTime        int
}

const reconfigResultTemplate = `
## Test {{ .Name }} NumResources {{ .NumResources }}

### Reloads and Time to Ready

- TimeToReadyTotal: {{ .TimeToReadyTotal }}
- TimeToReadyAvgSingle: {{ .TimeToReadyAvgSingle }}
- NGINX Reloads: {{ .NGINXReloads }}
- NGINX Reload Average Time: {{ .NGINXReloadAvgTime }}
- Reload distribution:
{{- range .ReloadBuckets }}
	- {{ .Le }}ms: {{ .Val }}
{{- end }}

### Event Batch Processing

- Event Batch Total: {{ .EventsCount }}
- Event Batch Processing Average Time: {{ .EventsAvgTime }}ms
- Event Batch Processing distribution:
{{- range .EventsBuckets }}
	- {{ .Le }}ms: {{ .Val }}
{{- end }}

`

func writeReconfigResults(dest io.Writer, results reconfigTestResults) error {
	tmpl, err := template.New("results").Parse(reconfigResultTemplate)
	if err != nil {
		return err
	}

	return tmpl.Execute(dest, results)
}
