package framework

import (
	"errors"
	"fmt"
	"time"

	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// TODO: having gomega and regular queries here is bad

func getFirstValueOfVector(query string, promInstance PrometheusInstance) (float64, error) {
	result, err := promInstance.Query(query)
	if err != nil {
		return 0, err
	}

	val, err := GetFirstValueOfPrometheusVector(result)
	if err != nil {
		return 0, err
	}

	return val, nil
}

func getBuckets(query string, promInstance PrometheusInstance) ([]Bucket, error) {
	result, err := promInstance.Query(query)
	if err != nil {
		return nil, err
	}

	res, ok := result.(model.Vector)
	if !ok {
		return nil, errors.New("could not convert result to vector")
	}

	buckets := make([]Bucket, 0, len(res))

	for _, sample := range res {
		le := sample.Metric["le"]
		val := float64(sample.Value)
		bucket := Bucket{
			Le:  string(le),
			Val: int(val),
		}
		buckets = append(buckets, bucket)
	}

	return buckets, nil
}

func GetReloadCount(promInstance PrometheusInstance, ngfPodName string) (float64, error) {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_nginx_reloads_total{pod="%[1]s"}`,
			ngfPodName,
		),
		promInstance,
	)
}

//func getReloadErrsCount(promInstance PrometheusInstance, ngfPodName string, startTime time.Time) float64 {
//	return getFirstValueOfVector(
//		fmt.Sprintf(
//			`nginx_gateway_fabric_nginx_reload_errors_total{pod="%[1]s"}`+
//				` - `+
//				`nginx_gateway_fabric_nginx_reload_errors_total{pod="%[1]s"} @ %d`,
//			ngfPodName,
//			startTime.Unix(),
//		),
//		promInstance,
//	)
//}

func GetReloadAvgTime(promInstance PrometheusInstance, ngfPodName string) (float64, error) {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_nginx_reloads_milliseconds_sum{pod="%[1]s"}`+
				` / `+
				`nginx_gateway_fabric_nginx_reloads_total{pod="%[1]s"}`,
			ngfPodName,
		),
		promInstance,
	)
}

func GetReloadBuckets(promInstance PrometheusInstance, ngfPodName string) ([]Bucket, error) {
	return getBuckets(
		fmt.Sprintf(
			`nginx_gateway_fabric_nginx_reloads_milliseconds_bucket{pod="%[1]s"}`,
			ngfPodName,
		),
		promInstance,
	)
}

func GetEventsCount(promInstance PrometheusInstance, ngfPodName string) (float64, error) {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_event_batch_processing_milliseconds_count{pod="%[1]s"}`,
			ngfPodName,
		),
		promInstance,
	)
}

func GetEventsAvgTime(promInstance PrometheusInstance, ngfPodName string) (float64, error) {
	return getFirstValueOfVector(
		fmt.Sprintf(
			`nginx_gateway_fabric_event_batch_processing_milliseconds_sum{pod="%[1]s"}`+
				` / `+
				`nginx_gateway_fabric_event_batch_processing_milliseconds_count{pod="%[1]s"}`,
			ngfPodName,
		),
		promInstance,
	)
}

func GetEventsBuckets(promInstance PrometheusInstance, ngfPodName string) ([]Bucket, error) {
	return getBuckets(
		fmt.Sprintf(
			`nginx_gateway_fabric_event_batch_processing_milliseconds_bucket{pod="%[1]s"}`,
			ngfPodName,
		),
		promInstance,
	)
}

func CreateMetricExistChecker(
	promInstance PrometheusInstance,
	query string,
	getTime func() time.Time,
	modifyTime func(),
) func() error {
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

func CreateEndTimeFinder(
	promInstance PrometheusInstance,
	query string,
	startTime time.Time,
	t *time.Time,
	queryRangeStep time.Duration,
) func() error {
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

//func createResponseChecker(url, address string, requestTimeout time.Duration) func() error {
//	return func() error {
//		status, _, err := Get(url, address, requestTimeout)
//		if err != nil {
//			return fmt.Errorf("bad response: %w", err)
//		}
//
//		if status != 200 {
//			return fmt.Errorf("unexpected status code: %d", status)
//		}
//
//		return nil
//	}
//}

type Bucket struct {
	Le  string
	Val int
}
