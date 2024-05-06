package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var otelTemplate = gotemplate.Must(gotemplate.New("otel").Parse(otelTemplateText))

func executeTelemetry(conf dataplane.Configuration) []executeResult {
	if conf.Telemetry.Endpoint != "" {
		result := executeResult{
			dest: httpConfigFile,
			data: execute(otelTemplate, conf.Telemetry),
		}

		return []executeResult{result}
	}

	return nil
}
