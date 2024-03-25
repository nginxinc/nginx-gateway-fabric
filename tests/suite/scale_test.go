package suite

import (
	"fmt"
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

var _ = Describe("Scale test", Ordered, Label("nfr", "scale"), func() {
	promManifest := []string{
		"scale/prom.yaml",
	}
	//matchesManifest := []string{
	//	"scale/matches.yaml",
	//}
	//upstreamsManifest := []string{
	//	"scale/upstreams.yaml",
	//}
	ns := &core.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "scale",
		},
	}

	var resultsDir string
	var outFile *os.File
	var ngfPodName string
	var err error

	BeforeAll(func() {
		// Need longer create timeout to allow for all the apps
		resourceManager.TimeoutConfig.CreateTimeout = 180 * time.Second
		Expect(resourceManager.Apply([]client.Object{ns})).To(Succeed())
		Expect(resourceManager.ApplyFromFiles(promManifest, ngfNamespace)).To(Succeed())
		resultsDir, err = framework.CreateResultsDir("scale", version)
		Expect(err).ToNot(HaveOccurred())

		filename := filepath.Join(resultsDir, fmt.Sprintf("%s.md", version))
		outFile, err = framework.CreateResultsFile(filename)
		Expect(err).ToNot(HaveOccurred())
		Expect(framework.WriteSystemInfoToFile(outFile, clusterInfo, *plusEnabled)).To(Succeed())

		podNames, err := framework.GetReadyNGFPodNames(k8sClient, ngfNamespace, releaseName, timeoutConfig.GetTimeout)
		Expect(err).ToNot(HaveOccurred())
		Expect(podNames).To(HaveLen(1))
		ngfPodName = podNames[0]
	})

	// TODO: Change values back to real values
	It("scales HTTP listeners to 64", func() {
		testName := "TestScale_Listeners"
		testResultsDir := filepath.Join(resultsDir, testName)
		Expect(framework.GenerateScaleListenerManifests(64, testResultsDir, false /*non-tls*/)).To(Succeed())
		runScaleTest(
			testResultsDir,
			testName,
			ngfPodName,
			ns,
			[]string{"# Listeners", "Time to Ready (s)", "Error"},
			"http",
			2,
			outFile,
		)
	})

	It("scales HTTPS listeners to 64", func() {
		testName := "TestScale_HTTPSListeners"
		testResultsDir := filepath.Join(resultsDir, testName)
		Expect(framework.GenerateScaleListenerManifests(64, testResultsDir, true)).To(Succeed())
		runScaleTest(
			testResultsDir,
			testName,
			ngfPodName,
			ns,
			[]string{"# Listeners", "Time to Ready (s)", "Error"},
			"https",
			64,
			outFile,
		)
	})

	It("scales HTTPRoutes to 1000", func() {
		testName := "TestScale_HTTPRoutes"
		testResultsDir := filepath.Join(resultsDir, testName)
		Expect(framework.GenerateScaleHTTPRouteManifests(1000, testResultsDir)).To(Succeed())
		runScaleTest(
			testResultsDir,
			testName,
			ngfPodName,
			ns,
			[]string{"# HTTPRoutes", "Time to Ready (s)", "Error"},
			"http",
			2,
			outFile,
		)
	})

	//It("scales upstream servers to 648", func() {
	//	testName := "TestScale_UpstreamServers"
	//	manifestDir := filepath.Join(resultsDir, testName)
	//	Expect(resourceManager.ApplyFromFiles(upstreamsManifest, ns.Name)).To(Succeed())
	//
	//	// clean-up files and resources
	//})

	AfterAll(func() {
		Expect(resourceManager.Delete([]client.Object{ns})).To(Succeed())
		outFile.Close()
	})
})

func runScaleTest(
	resultsDir, testName, ngfPodName string,
	ns *core.Namespace,
	resultHeaders []string,
	protocol string,
	numIterations int,
	resultsFile *os.File,
) {
	ttrCsvFile, writer, err := framework.NewCSVResultsWriter(resultsDir, "ttr.csv", resultHeaders...)
	Expect(err).ToNot(HaveOccurred())

	startTime := time.Now().Unix()

	preReqManifestFiles, err := framework.GetYamlFileList(framework.GetPrereqDirName(resultsDir))
	Expect(err).ToNot(HaveOccurred())

	Expect(resourceManager.ApplyFromFiles(preReqManifestFiles, ns.Name)).To(Succeed())

	// Expect(resourceManager.WaitForAppsToBeReady(ns.Name)).To(Succeed())

	manifestsDir := filepath.Join(resultsDir, "manifests")
	for i := 0; i < numIterations; i++ {

		manifestFile := []string{filepath.Join(manifestsDir, fmt.Sprintf("manifest-%d.yaml", i))}

		Expect(resourceManager.ApplyFromFiles(manifestFile, ns.Name)).To(Succeed())

		url := fmt.Sprintf("%s://%d.example.com", protocol, i)

		ttr, err := framework.WaitForResponseForHost(url, address, timeoutConfig.RequestTimeout)

		seconds := ttr.Seconds()
		record := []string{strconv.Itoa(i + 1), strconv.FormatFloat(seconds, 'f', -1, 64)}
		if err != nil {
			record = append(record, err.Error())
		}

		Expect(writer.Write(record)).To(Succeed())

		time.Sleep(*scaleTestsDelay)
	}

	endTime := time.Now().Unix()

	writer.Flush()
	Expect(ttrCsvFile.Close()).To(Succeed())

	output, err := framework.GenerateTTRPNG(resultsDir, ttrCsvFile.Name(), fmt.Sprintf("%s.png", "TTR"))
	Expect(err).ToNot(HaveOccurred(), string(output))

	resourceManifestFiles, err := framework.GetYamlFileList(manifestsDir)
	Expect(resourceManager.DeleteFromFiles(resourceManifestFiles, ns.Name)).To(Succeed())
	Expect(os.RemoveAll(manifestsDir)).To(Succeed())

	gatherScaleTestResults(resultsDir, testName, ngfPodName, startTime, endTime, resultsFile)
}

func gatherScaleTestResults(resultsDir, testName, podName string, startTime, endTime int64, resultsFile *os.File) {
	// Get CPU usage data
	resultHeaders := []string{"Timestamp", "CPU Usage (core seconds)"}
	cpuCSVFile, cpuCSVWriter, err := framework.NewCSVResultsWriter(resultsDir, "CPU.csv", resultHeaders...)
	Expect(err).ToNot(HaveOccurred())
	Expect(framework.WriteGKEMetricsData(podName, framework.CpuMetricName, startTime, endTime, cpuCSVWriter)).To(Succeed())
	cpuCSVWriter.Flush()
	Expect(cpuCSVFile.Close()).To(Succeed())
	framework.GenerateCPUPNG(resultsDir, cpuCSVFile.Name(), fmt.Sprintf("%s.png", "CPU"))
	// Expect(err).ToNot(HaveOccurred(), string(output))

	// Get Memory usage data
	resultHeaders = []string{"Timestamp", "Memory Usage (bytes)"}
	memCSVFile, memCSVWriter, err := framework.NewCSVResultsWriter(resultsDir, "Memory.csv", resultHeaders...)
	Expect(err).ToNot(HaveOccurred())
	Expect(framework.WriteGKEMetricsData(podName, framework.MemoryMetricName, startTime, endTime, memCSVWriter)).To(Succeed())
	memCSVWriter.Flush()
	Expect(memCSVFile.Close()).To(Succeed())
	framework.GenerateMemoryPNG(resultsDir, memCSVFile.Name(), fmt.Sprintf("%s.png", "Memory"))
	// Expect(err).ToNot(HaveOccurred(), string(output))

	// Write output to results file:
	// 	duration := endTime.Sub(startTime)
	// Get Prometheus data using
	// endUnix + 10*Seconds
	// TODO: populate with real data

	var reloadCount int
	var reloadErrsCount int
	var reloadAvgTime int
	var reloadsUnder500ms string

	var eventsCount int
	var eventsErrsCount int
	var eventsAvgTime int
	var eventsUnder500ms string

	var nginxErrs string
	var ngfErrs string
	var podRestarts string

	text := fmt.Sprintf("\n### Test %s: \n\n", testName)
	_, err = fmt.Fprint(resultsFile, text)
	Expect(err).ToNot(HaveOccurred())

	reloadsText := fmt.Sprintf(reloadFmtString, reloadCount, reloadErrsCount, reloadAvgTime, reloadsUnder500ms)
	eventsText := fmt.Sprintf(eventsFmtString, eventsCount, eventsErrsCount, eventsAvgTime, eventsUnder500ms)
	ngxErrsText := fmt.Sprintf(nginxErrsFmtString, nginxErrs)
	ngfErrsText := fmt.Sprintf(ngfErrsFmtString, ngfErrs)
	podRestartsText := fmt.Sprintf(podRestartsFmtString, podRestarts)
	cpuText := fmt.Sprintf(cpuFmtString, version, testName)
	memoryText := fmt.Sprintf(memoryFmtString, version, testName)
	ttrText := fmt.Sprintf(ttrFmtString, version, testName)

	results := fmt.Sprintf("%v\n\n%v\n%v\n%v\n%v\n%v\n%v\n%v\n", reloadsText, eventsText, ngxErrsText, ngfErrsText, podRestartsText, cpuText, memoryText, ttrText)
	_, err = fmt.Fprint(resultsFile, results)
	Expect(err).ToNot(HaveOccurred())
}

var (
	reloadFmtString = `
Reloads:

| Total | Total Errors | Ave Time (ms)      | <= 500ms |
|-------|--------------|--------------------|----------|
| %d   | %d            | %d | %v     |
`

	eventsFmtString = `
Event Batch Processing:

| Total | Ave Time (ms)      | <= 500ms | <= 1000ms |
|-------|--------------------|----------|-----------|
| %d   | %d | %v   | %v    |
`
	nginxErrsFmtString   = "**NGINX Errors**: %v."
	ngfErrsFmtString     = "**NGF Errors**: %v."
	podRestartsFmtString = "**Pod Restarts**: %v."
	cpuFmtString         = "**CPU**: ![CPU.png](/tests/results/scale/%s/%s/CPU.png)."
	memoryFmtString      = "**Memory**: ![Memory.png](/tests/results/scale/%s/%s/Memory.png)."
	ttrFmtString         = "**Time To Ready**: ![TTR.png](/tests/results/scale/%s/%s/TTR.png)."
)
