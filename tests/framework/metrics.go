package framework

import (
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"

	"google.golang.org/api/iterator"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	CpuMetricName    = "kubernetes.io/container/cpu/core_usage_time"
	MemoryMetricName = "kubernetes.io/container/memory/used_bytes"
	PromReloads      = "prometheus.googleapis.com/nginx_gateway_fabric_nginx_reloads_total/counter"
)

// NGFPrometheusMetrics are the relevant NGFPrometheusMetrics for scale tests
type NGFPrometheusMetrics struct {
	ReloadCount       string
	ReloadErrsCount   string
	ReloadAvgTime     string
	ReloadsUnder500ms string
	EventsCount       string
	EventsErrsCount   string
	EventsAvgTime     string
	EventsUnder500ms  string
}

// WriteGKEMetricsData gathers the requested timeseries metrics data and writes it to the given CSV file.
func WriteGKEMetricsData(podName, metricType string, startTime, endTime int64, csvWriter *csv.Writer) error {
	// Create a context and specify the desired project ID.
	ctx := context.Background()
	// projectID := os.Getenv("GKE_PROJECT")
	// TODO: Remove hardcoded project ID
	projectID := ""

	// Create a MetricsClient for the Google Cloud Monitoring API.
	metricsClient, err := monitoring.NewMetricClient(ctx)
	if err != nil {
		return fmt.Errorf("failed to create MetricsClient: %v", err)
	}
	defer metricsClient.Close()

	// Create a filter string to retrieve the required metric for the specific pod.
	metricValueName := strings.Split(metricType, "/")[3]
	valueAggregate := fmt.Sprintf("value_aggregate: aggregate(value.%s)", metricValueName)
	filter := fmt.Sprintf(`resource.type = "k8s_container" AND resource.label.pod_name = "%s" AND metric.type = "%s"`, podName, metricType)
	aggregation := &monitoringpb.Aggregation{
		GroupByFields: []string{valueAggregate, "max(value_aggregate)"},
	}

	fmt.Println("Filter: ", filter)
	// Prepare the request to retrieve time series data.
	req := &monitoringpb.ListTimeSeriesRequest{
		Name:        "projects/" + projectID,
		Filter:      filter,
		Aggregation: aggregation,
		Interval: &monitoringpb.TimeInterval{
			StartTime: &timestamppb.Timestamp{Seconds: startTime},
			EndTime:   &timestamppb.Timestamp{Seconds: endTime},
		},
	}

	fmt.Println("start-time", startTime, "end-time", endTime)

	dataFound, err := writeMetricsData(ctx, metricsClient, req, csvWriter)

	if err != nil {
		return err
	}

	if !dataFound {
		// no data yet, wait and try again with new timestamp
		time.Sleep(1 * time.Minute)
		req.Interval.EndTime = &timestamppb.Timestamp{Seconds: time.Now().Unix()}
		dataFound, err = writeMetricsData(ctx, metricsClient, req, csvWriter)
		if err != nil {
			return err
		}
		if !dataFound {
			fmt.Println("Waited for 1 minute but still no data for time period, exiting...")
		}
	}
	return nil
}

func writeMetricsData(ctx context.Context, metricsClient *monitoring.MetricClient, req *monitoringpb.ListTimeSeriesRequest, csvWriter *csv.Writer) (bool, error) {
	var dataFound bool
	// Retrieve time series data.
	it := metricsClient.ListTimeSeries(ctx, req)

	for {
		ts, err := it.Next()
		if errors.Is(err, iterator.Done) {
			fmt.Println("no more entries in iterator")
			break
		}
		if err != nil {
			return false, fmt.Errorf("failed to iterate time series data: %v", err)
		}
		fmt.Println("Points in time series: ", len(ts.Points))
		for _, point := range ts.Points {
			timestamp := point.Interval.EndTime.Seconds
			fmt.Println()
			data := fmt.Sprintf("%.2f", point.GetValue().GetDoubleValue())
			if data != "0.00" {
				dataFound = true
				err = csvWriter.Write([]string{strconv.FormatInt(timestamp, 10), data})
				fmt.Println("Timestamp: ", strconv.FormatInt(timestamp, 10), timestamp)
				if err != nil {
					return false, fmt.Errorf("failed to write data to CSV file: %v", err)
				}
			} else {
				fmt.Println("No data yet, try again...")
			}
		}
	}
	return dataFound, nil
}

//func getPrometheusMetricsData() (error, NGFPrometheusMetrics) {
//	// Create a context and specify the desired project ID.
//	ctx := context.Background()
//	// projectID := os.Getenv("GKE_PROJECT")
//	// TODO: Remove hardcoded project ID
//	projectID := "f5-gcs-7899-ptg-ingrss-ctlr"
//
//	var ngfPromMetrics NGFPrometheusMetrics
//
//	// Create a MetricsClient for the Google Cloud Monitoring API.
//	metricsClient, err := monitoring.NewMetricClient(ctx)
//	if err != nil {
//		return fmt.Errorf("failed to create MetricsClient: %v", err), ngfPromMetrics
//	}
//	defer metricsClient.Close()
//
//}
