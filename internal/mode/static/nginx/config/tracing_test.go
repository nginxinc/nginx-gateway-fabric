package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteTracing(t *testing.T) {
	conf := dataplane.Configuration{
		Tracing: dataplane.Tracing{
			ExporterEndpoint: "1.2.3.4:123",
			Enabled:          true,
			ServiceName:      "ngf:my-name",
			Interval:         "5s",
			BatchSize:        512,
			BatchCount:       4,
		},
	}

	g := NewWithT(t)
	expSubStrings := map[string]int{
		"endpoint 1.2.3.4:123;":          1,
		"otel_trace on;":                 1,
		"otel_service_name ngf:my-name;": 1,
		"interval 5s;":                   1,
		"batch_size  512;":               1,
		"batch_count 4;":                 1,
	}

	maps := string(executeTracing(conf))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(expCount).To(Equal(strings.Count(maps, expSubStr)))
	}
}

func TestExecuteTracingNil(t *testing.T) {
	conf := dataplane.Configuration{
		Tracing: dataplane.Tracing{},
	}

	g := NewWithT(t)
	expSubStrings := map[string]int{
		"endpoint ;": 0,
	}

	maps := string(executeTracing(conf))
	for expSubStr, expCount := range expSubStrings {
		g.Expect(expCount).To(Equal(strings.Count(maps, expSubStr)))
	}
}
