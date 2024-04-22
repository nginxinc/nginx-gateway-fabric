package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteTelemetry(t *testing.T) {
	conf := dataplane.Configuration{
		Telemetry: dataplane.Telemetry{
			Endpoint:    "1.2.3.4:123",
			ServiceName: "ngf:gw-ns:gw-name:my-name",
			Interval:    "5s",
			BatchSize:   512,
			BatchCount:  4,
			SpanAttributes: []ngfAPI.SpanAttribute{
				{
					Key:   "key1",
					Value: "val1",
				},
				{
					Key:   "key2",
					Value: "val2",
				},
			},
		},
	}

	g := NewWithT(t)
	expSubStrings := map[string]int{
		"endpoint 1.2.3.4:123;":                        1,
		"otel_service_name ngf:gw-ns:gw-name:my-name;": 1,
		"interval 5s;":                                 1,
		"batch_size 512;":                              1,
		"batch_count 4;":                               1,
		"otel_span_attr":                               2,
	}

	for expSubStr, expCount := range expSubStrings {
		g.Expect(expCount).To(Equal(strings.Count(string(executeTelemetry(conf)), expSubStr)))
	}
}

func TestExecuteTelemetryNil(t *testing.T) {
	conf := dataplane.Configuration{
		Telemetry: dataplane.Telemetry{},
	}

	g := NewWithT(t)

	g.Expect(string(executeTelemetry(conf))).To(BeEmpty())
}
