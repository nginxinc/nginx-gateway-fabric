package framework

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	prometheusNamespace   = "prom"
	prometheusReleaseName = "prom"
)

var defaultPrometheusQueryTimeout = 2 * time.Second

// PrometheusConfig is the configuration for installing Prometheus.
type PrometheusConfig struct {
	// ScrapeInterval is the interval at which Prometheus scrapes metrics.
	ScrapeInterval time.Duration
	// QueryTimeout is the timeout for Prometheus queries.
	// Default is 2s.
	QueryTimeout time.Duration
}

// InstallPrometheus installs Prometheus in the cluster.
// It waits for Prometheus pods to be ready before returning.
func InstallPrometheus(
	rm ResourceManager,
	cfg PrometheusConfig,
) (PrometheusInstance, error) {
	output, err := exec.Command(
		"helm",
		"repo",
		"add",
		"prometheus-community",
		"https://prometheus-community.github.io/helm-charts",
	).CombinedOutput()
	if err != nil {
		return PrometheusInstance{}, fmt.Errorf("failed to add Prometheus helm repo: %w; output: %s", err, string(output))
	}

	output, err = exec.Command(
		"helm",
		"repo",
		"update",
	).CombinedOutput()
	if err != nil {
		return PrometheusInstance{}, fmt.Errorf("failed to update helm repos: %w; output: %s", err, string(output))
	}

	scrapeInterval := fmt.Sprintf("%ds", int(cfg.ScrapeInterval.Seconds()))

	//nolint:gosec
	output, err = exec.Command(
		"helm",
		"install",
		prometheusReleaseName,
		"prometheus-community/prometheus",
		"--create-namespace",
		"--namespace", prometheusNamespace,
		"--set", fmt.Sprintf("server.global.scrape_interval=%s", scrapeInterval),
		"--wait",
	).CombinedOutput()
	if err != nil {
		return PrometheusInstance{}, fmt.Errorf("failed to install Prometheus: %w; output: %s", err, string(output))
	}

	pods, err := rm.GetPods(prometheusNamespace, client.MatchingLabels{
		"app.kubernetes.io/name": "prometheus",
	})
	if err != nil {
		return PrometheusInstance{}, fmt.Errorf("failed to get Prometheus pods: %w", err)
	}

	if len(pods) != 1 {
		return PrometheusInstance{}, fmt.Errorf("expected one Prometheus pod, found %d", len(pods))
	}

	pod := pods[0]

	if pod.Status.PodIP == "" {
		return PrometheusInstance{}, errors.New("the Prometheus pod has no IP")
	}

	var queryTimeout time.Duration
	if cfg.QueryTimeout == 0 {
		queryTimeout = defaultPrometheusQueryTimeout
	} else {
		queryTimeout = cfg.QueryTimeout
	}

	return PrometheusInstance{
		podIP:        pod.Status.PodIP,
		podName:      pod.Name,
		podNamespace: pod.Namespace,
		queryTimeout: queryTimeout,
	}, nil
}

// UninstallPrometheus uninstalls Prometheus from the cluster.
func UninstallPrometheus(rm ResourceManager) error {
	output, err := exec.Command(
		"helm",
		"uninstall",
		prometheusReleaseName,
		"-n", prometheusNamespace,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall Prometheus: %w; output: %s", err, string(output))
	}

	if err := rm.DeleteNamespace(prometheusNamespace); err != nil {
		return fmt.Errorf("failed to delete Prometheus namespace: %w", err)
	}

	return nil
}

const (
	// PrometheusPortForwardPort is the local port that will forward to the Prometheus API.
	PrometheusPortForwardPort = 9090
	prometheusAPIPort         = 9090
)

// PrometheusInstance represents a Prometheus instance in the cluster.
type PrometheusInstance struct {
	apiClient    v1.API
	podIP        string
	podName      string
	podNamespace string
	queryTimeout time.Duration
	portForward  bool
}

// PortForward starts port forwarding to the Prometheus instance.
func (ins *PrometheusInstance) PortForward(config *rest.Config, stopCh <-chan struct{}) error {
	if ins.portForward {
		panic("port forwarding already started")
	}

	ins.portForward = true

	ports := []string{fmt.Sprintf("%d:%d", PrometheusPortForwardPort, prometheusAPIPort)}
	return PortForward(config, ins.podNamespace, ins.podName, ports, stopCh)
}

func (ins *PrometheusInstance) getAPIClient() (v1.API, error) {
	var endpoint string
	if ins.portForward {
		endpoint = fmt.Sprintf("http://localhost:%d", PrometheusPortForwardPort)
	} else {
		// on GKE, test runner VM can access the pod directly
		endpoint = fmt.Sprintf("http://%s:%d", ins.podIP, prometheusAPIPort)
	}

	cfg := api.Config{
		Address: endpoint,
	}

	c, err := api.NewClient(cfg)
	if err != nil {
		return nil, err
	}

	return v1.NewAPI(c), nil
}

func (ins *PrometheusInstance) ensureAPIClient() error {
	if ins.apiClient == nil {
		api, err := ins.getAPIClient()
		if err != nil {
			return fmt.Errorf("failed to get Prometheus API client: %w", err)
		}
		ins.apiClient = api
	}

	return nil
}

// Query sends a query to Prometheus.
func (ins *PrometheusInstance) Query(query string) (model.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ins.queryTimeout)
	defer cancel()

	return ins.QueryWithCtx(ctx, query)
}

// QueryWithCtx sends a query to Prometheus with the specified context.
func (ins *PrometheusInstance) QueryWithCtx(ctx context.Context, query string) (model.Value, error) {
	if err := ins.ensureAPIClient(); err != nil {
		return nil, err
	}

	result, warnings, err := ins.apiClient.Query(ctx, query, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}

	if len(warnings) > 0 {
		slog.Info(
			"Prometheus query returned warnings",
			"query", query,
			"warnings", warnings,
		)
	}

	return result, nil
}

// QueryRange sends a range query to Prometheus.
func (ins *PrometheusInstance) QueryRange(query string, promRange v1.Range) (model.Value, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ins.queryTimeout)
	defer cancel()

	return ins.QueryRangeWithCtx(ctx, query, promRange)
}

// QueryRangeWithCtx sends a range query to Prometheus with the specified context.
func (ins *PrometheusInstance) QueryRangeWithCtx(ctx context.Context,
	query string, promRange v1.Range,
) (model.Value, error) {
	if err := ins.ensureAPIClient(); err != nil {
		return nil, err
	}

	result, warnings, err := ins.apiClient.QueryRange(ctx, query, promRange)
	if err != nil {
		return nil, fmt.Errorf("failed to query Prometheus: %w", err)
	}

	if len(warnings) > 0 {
		slog.Info(
			"Prometheus range query returned warnings",
			"query", query,
			"range", promRange,
			"warnings", warnings,
		)
	}

	return result, nil
}

// GetFirstValueOfPrometheusVector returns the first value of a Prometheus vector.
func GetFirstValueOfPrometheusVector(val model.Value) (float64, error) {
	res, ok := val.(model.Vector)
	if !ok {
		return 0, fmt.Errorf("expected a vector, got %T", val)
	}

	if len(res) == 0 {
		return 0, errors.New("empty vector")
	}

	return float64(res[0].Value), nil
}

// WritePrometheusMatrixToCSVFile writes a Prometheus matrix to a CSV file.
func WritePrometheusMatrixToCSVFile(fileName string, value model.Value) error {
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer file.Close()

	csvWriter := csv.NewWriter(file)

	matrix, ok := value.(model.Matrix)
	if !ok {
		return fmt.Errorf("expected a matrix, got %T", value)
	}

	for _, sample := range matrix {
		for _, pair := range sample.Values {
			record := []string{fmt.Sprint(pair.Timestamp.Unix()), pair.Value.String()}
			if err := csvWriter.Write(record); err != nil {
				return err
			}
		}
	}

	csvWriter.Flush()

	return nil
}
