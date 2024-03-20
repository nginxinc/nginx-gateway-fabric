//go:build scale
// +build scale

package scale

import (
	"context"
	"crypto/tls"
	"encoding/csv"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/util/wait"
)

// testing flags
var (
	numIterations = flag.Int("i", 1, "number of times to scale the resource")
	delay         = flag.Duration("delay", 0, "delay between each scaling iteration")
	version       = flag.String("version", "1.2.0", "version of NGF under test")
	plus          = flag.Bool("plus", false, "nginx-plus enabled")
)

func TestScale_Listeners(t *testing.T) {
	ip := getIP(t)
	url := fmt.Sprintf("http://%s/", ip)

	runScaleTest(
		t,
		[]string{"# Listeners", "Time to Ready (s)", "Error"},
		func(dir string) error {
			return generateScaleListenerManifests(*numIterations, dir, false /*non-tls*/)
		},
		url,
	)
}

func TestScale_HTTPSListeners(t *testing.T) {
	ip := getIP(t)
	url := fmt.Sprintf("https://%s/", ip)

	runScaleTest(
		t,
		[]string{"# HTTPS Listeners", "Time to Ready (s)", "Error"},
		func(dir string) error {
			return generateScaleListenerManifests(*numIterations, dir, true /*tls*/)
		},
		url,
	)
}

func TestScale_HTTPRoutes(t *testing.T) {
	ip := getIP(t)
	url := fmt.Sprintf("http://%s/", ip)

	runScaleTest(
		t,
		[]string{"# HTTPRoutes", "Time to Ready (s)", "Error"},
		func(dir string) error {
			return generateScaleHTTPRouteManifests(*numIterations, dir)
		},
		url,
	)
}

func runScaleTest(
	t *testing.T,
	resultHeaders []string,
	generateManifests func(dir string) error,
	url string,
) {
	t.Helper()
	manifestDir := t.Name()

	writer := newResultsWriter(t, t.Name(), resultHeaders...)

	if err := generateManifests(manifestDir); err != nil {
		t.Fatalf("failed to generate manifests: %s", err)
	}

	startTime := time.Now()
	startUnix := fmt.Sprintf("%d", startTime.Unix())

	if err := kubectlApply(getPrereqDirName(manifestDir)); err != nil {
		t.Fatalf("failed to apply prerequisite resources: %s", err)
	}

	t.Log("Waiting for all Pods to be Ready")
	if err := kubectlWaitAllPodsReady(); err != nil {
		t.Fatalf("failed to wait for all Pods to be Ready: %s", err)
	}

	for i := 0; i < *numIterations; i++ {
		t.Logf("Scaling up to %d resources", i)

		manifestFile := filepath.Join(manifestDir, fmt.Sprintf("manifest-%d.yaml", i))

		if err := kubectlApply(manifestFile); err != nil {
			t.Errorf("failed to scale up: %s", err)
		}

		host := fmt.Sprintf("%d.example.com", i)

		t.Logf("Sending request to url %s with host %s...", url, host)

		ttr, err := waitForResponseForHost(url, host)

		seconds := ttr.Seconds()
		record := []string{strconv.Itoa(i + 1), strconv.FormatFloat(seconds, 'f', -1, 64)}
		if err != nil {
			record = append(record, err.Error())
		}

		if err = writer.Write(record); err != nil {
			t.Fatalf("failed to write time to ready to csv file: %s", err)
		}

		time.Sleep(*delay)
	}

	endTime := time.Now()
	endUnix := fmt.Sprintf("%d", endTime.Unix())

	// This accounts for prometheus 10s scraping window
	endUnixPlusTen := fmt.Sprintf("%d", endTime.Add(10*time.Second).Unix())

	records := [][]string{
		{"Test Start", "Test End", "Test End + 10s", "Duration"},
		{startUnix, endUnix, endUnixPlusTen, endTime.Sub(startTime).String()},
	}

	if err := writer.WriteAll(records); err != nil {
		t.Logf("failed to write records to csv")
	}
}

func getIP(t *testing.T) string {
	t.Helper()

	ip := os.Getenv("NGF_IP")
	if ip == "" {
		t.Fatalf("NGF_IP env var not set")
	}

	return ip
}

func newResultsWriter(t *testing.T, testName string, resultHeaders ...string) *csv.Writer {
	t.Helper()

	versionDir := filepath.Join("results", *version)
	if err := os.Mkdir(versionDir, 0o750); err != nil && !os.IsExist(err) {
		t.Fatalf("failed to create results version directory: %s", err)
	}

	testDirName := testName
	if *plus {
		testDirName += "_Plus"
	}

	dir := filepath.Join(versionDir, testDirName)
	if err := os.Mkdir(dir, 0o750); err != nil {
		t.Fatalf("failed to create results test directory: %s", err)
	}

	file, err := os.Create(filepath.Join(dir, "results.csv"))
	if err != nil {
		t.Fatalf("failed to create results csv file: %s", err)
	}

	writer := csv.NewWriter(file)

	if err = writer.Write(resultHeaders); err != nil {
		t.Fatalf("failed to write headers to csv file: %s", err)
	}

	t.Cleanup(func() {
		writer.Flush()
		_ = file.Close()
	})

	return writer
}

func kubectlApply(filename string) error {
	if err := kubectlExec("apply", "-f", filename); err != nil {
		return fmt.Errorf("error applying %s: %w", filename, err)
	}

	return nil
}

func kubectlWaitAllPodsReady() error {
	if err := kubectlExec("wait", "pod", "--all", "--for=condition=Ready"); err != nil {
		return fmt.Errorf("error waiting for all pods to be ready:%w", err)
	}

	return nil
}

func kubectlExec(arg ...string) error {
	cmd := exec.Command("kubectl", arg...)

	return cmd.Run()
}

func waitForResponseForHost(url, host string) (time.Duration, error) {
	client := &http.Client{}

	if strings.HasPrefix(url, "https") {
		customTransport := http.DefaultTransport.(*http.Transport)
		customTransport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true, // nolint: gosec
			ServerName:         host,
		}
		client.Transport = customTransport
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Host = host

	start := time.Now()

	err = wait.PollUntilContextCancel(
		ctx,
		200*time.Millisecond,
		true,
		func(ctx context.Context) (done bool, err error) {
			resp, err := client.Do(req)
			if err != nil {
				fmt.Println("Retrying GET request", "error", err)
				return false, err
			}

			if resp.StatusCode == http.StatusOK {
				return true, nil
			}

			fmt.Println("Retrying GET request", "host", host, "status", resp.Status)
			return false, nil
		})

	return time.Since(start), err
}
