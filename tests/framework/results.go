package framework

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	vegeta "github.com/tsenart/vegeta/v12/lib"
)

// CreateResultsDir creates and returns the name of the results directory for a test.
func CreateResultsDir(testName, version string) (string, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	dirName := filepath.Join(filepath.Dir(pwd), "results", testName, version)

	return dirName, os.MkdirAll(dirName, 0o777)
}

// CreateResultsFile creates and returns the results file for a test.
func CreateResultsFile(filename string) (*os.File, error) {
	outFile, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o777)
	if err != nil {
		return nil, err
	}

	return outFile, nil
}

// WriteSystemInfoToFile writes the cluster system info to the given file.
func WriteSystemInfoToFile(file *os.File, ci ClusterInfo) error {
	clusterType := "Local"
	if ci.IsGKE {
		clusterType = "GKE"
	}
	text := fmt.Sprintf(
		//nolint:lll
		"# Results\n\n## Test environment\n\n%s Cluster:\n\n- Node count: %d\n- k8s version: %s\n- vCPUs per node: %d\n- RAM per node: %s\n- Max pods per node: %d\n",
		clusterType, ci.NodeCount, ci.K8sVersion, ci.CPUCountPerNode, ci.MemoryPerNode, ci.MaxPodsPerNode,
	)
	if _, err := fmt.Fprint(file, text); err != nil {
		return err
	}
	if ci.IsGKE {
		if _, err := fmt.Fprintf(file, "- Zone: %s\n- Instance Type: %s\n", ci.GkeZone, ci.GkeInstanceType); err != nil {
			return err
		}
	}
	return nil
}

// GeneratePNG generates a PNG using gnuplot.
func GeneratePNG(resultsDir, inputFilename, outputFilename string) ([]byte, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	gnuplotCfg := filepath.Join(filepath.Dir(pwd), "scripts", "requests-plot.gp")

	files := fmt.Sprintf("inputfile='%s';outputfile='%s'", inputFilename, outputFilename)
	cmd := exec.Command("gnuplot", "-e", files, "-c", gnuplotCfg)
	cmd.Dir = resultsDir

	return cmd.CombinedOutput()
}

// WriteResults writes the vegeta metrics results to the results file in text format.
func WriteResults(resultsFile *os.File, metrics *vegeta.Metrics) error {
	reporter := vegeta.NewTextReporter(metrics)

	return reporter.Report(resultsFile)
}
