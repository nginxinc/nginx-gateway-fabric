package config

import (
	"strings"
	"testing"

	. "github.com/onsi/gomega"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

func TestExecuteMainIncludesConfig(t *testing.T) {
	t.Parallel()

	completeConfiguration := dataplane.Configuration{
		Telemetry: dataplane.Telemetry{
			Endpoint:    "1.2.3.4:123",
			ServiceName: "ngf:gw-ns:gw-name:my-name",
			Interval:    "5s",
			BatchSize:   512,
			BatchCount:  4,
		},
		Logging: dataplane.Logging{
			ErrorLevel: "info",
		},
	}

	missingTelemetryEndpoint := dataplane.Configuration{
		Logging: dataplane.Logging{
			ErrorLevel: "info",
		},
	}

	// Configuration.Logging will always be set, so no need to test if it is missing
	tests := []struct {
		name                      string
		conf                      dataplane.Configuration
		expTelemetryEndpointCount int
	}{
		{
			name:                      "complete configuration",
			conf:                      completeConfiguration,
			expTelemetryEndpointCount: 1,
		},
		{
			name:                      "missing telemetry endpoint",
			conf:                      missingTelemetryEndpoint,
			expTelemetryEndpointCount: 0,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			g := NewWithT(t)

			res := executeMainIncludesConfig(test.conf)

			g.Expect(res).To(HaveLen(1))

			g.Expect(strings.Count(
				string(res[0].data),
				"load_module modules/ngx_otel_module.so;"),
			).To(Equal(test.expTelemetryEndpointCount))

			g.Expect(strings.Count(
				string(res[0].data),
				"error_log stderr "+test.conf.Logging.ErrorLevel+";",
			)).To(Equal(1))
		})
	}
}
