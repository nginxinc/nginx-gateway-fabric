package framework

import (
	"errors"
	"fmt"
	"time"

	. "github.com/onsi/gomega"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

func getFirstValueOfVector(query string, promInstance PrometheusInstance) float64 {
	result, err := promInstance.Query(query)
	Expect(err).ToNot(HaveOccurred())

	val, err := GetFirstValueOfPrometheusVector(result)
	Expect(err).ToNot(HaveOccurred())

	return val
}

func getBuckets(query string, promInstance PrometheusInstance) []bucket {
	result, err := promInstance.Query(query)
	Expect(err).ToNot(HaveOccurred())

	res, ok := result.(model.Vector)
	Expect(ok).To(BeTrue())

	buckets := make([]bucket, 0, len(res))

	for _, sample := range res {
		le := sample.Metric["le"]
		val := float64(sample.Value)
		bucket := bucket{
			Le:  string(le),
			Val: int(val),
		}
		buckets = append(buckets, bucket)
	}

	return buckets
}

func GetReloadCount(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) float64 {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_nginx_reloads_total{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_nginx_reloads_total{pod="%[1]s"} @ %d`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)

	//return getFirstValueOfVector(
	//	fmt.Sprintf(
	//		`nginx_gateway_fabric_nginx_reloads_milliseconds_count{pod="%[1]s"}`+
	//			` - `+
	//			`nginx_gateway_fabric_nginx_reloads_milliseconds_count{pod="%[1]s"} @ %d`,
	//		ngfPodName,
	//		startTime.Unix(),
	//	),
	//	promInstance,
	//)
}

func getReloadErrsCount(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) float64 {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_nginx_reload_errors_total{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_nginx_reload_errors_total{pod="%[1]s"} @ %d`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)
}

func getReloadAvgTime(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) float64 {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`(nginx_gateway_fabric_nginx_reloads_milliseconds_sum{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_nginx_reloads_milliseconds_sum{pod="%[1]s"} @ %[2]d)`+
				` / `+
				`(nginx_gateway_fabric_nginx_reloads_total{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_nginx_reloads_total{pod="%[1]s"} @ %[2]d)`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)
}

func getReloadBuckets(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) []bucket {
	return getBuckets(
		fmt.Sprintf(
			`nginx_gateway_fabric_nginx_reloads_milliseconds_bucket{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_nginx_reloads_milliseconds_bucket{pod="%[1]s"} @ %d`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)
}

func getEventsCount(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) float64 {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_event_batch_processing_milliseconds_count{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_event_batch_processing_milliseconds_count{pod="%[1]s"} @ %d`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)
}

func getEventsAvtTime(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) float64 {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`(nginx_gateway_fabric_event_batch_processing_milliseconds_sum{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_event_batch_processing_milliseconds_sum{pod="%[1]s"} @ %[2]d)`+
				` / `+
				`(nginx_gateway_fabric_event_batch_processing_milliseconds_count{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_event_batch_processing_milliseconds_count{pod="%[1]s"} @ %[2]d)`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)
}

func getEventsBuckets(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) []bucket {
	return getBuckets(
		fmt.Sprintf(
			`nginx_gateway_fabric_event_batch_processing_milliseconds_bucket{pod="%[1]s"}`+
				` - `+
				`nginx_gateway_fabric_event_batch_processing_milliseconds_bucket{pod="%[1]s"} @ %d`,
			ngfPodName,
			startTime.Unix(),
		),
		promInstance,
	)
}

func CreateMetricExistChecker(promInstance PrometheusInstance, query string, getTime func() time.Time, modifyTime func()) func() error {
	return func() error {
		queryWithTimestamp := fmt.Sprintf("%s @ %d", query, getTime().Unix())

		result, err := promInstance.Query(queryWithTimestamp)
		if err != nil {
			return fmt.Errorf("failed to query Prometheus: %w", err)
		}

		if result.String() == "" {
			modifyTime()
			return errors.New("empty result")
		}

		return nil
	}
}

func CreateEndTimeFinder(promInstance PrometheusInstance, query string, startTime time.Time, t *time.Time, queryRangeStep time.Duration) func() error {
	return func() error {
		result, err := promInstance.QueryRange(query, v1.Range{
			Start: startTime,
			End:   *t,
			Step:  queryRangeStep,
		})
		if err != nil {
			return fmt.Errorf("failed to query Prometheus: %w", err)
		}

		if result.String() == "" {
			*t = time.Now()
			return errors.New("empty result")
		}

		return nil
	}
}

func createResponseChecker(url, address string, requestTimeout time.Duration) func() error {
	return func() error {
		status, _, err := Get(url, address, requestTimeout)
		if err != nil {
			return fmt.Errorf("bad response: %w", err)
		}

		if status != 200 {
			return fmt.Errorf("unexpected status code: %d", status)
		}

		return nil
	}
}

type bucket struct {
	Le  string
	Val int
}
