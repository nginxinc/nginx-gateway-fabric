package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var otelTemplate = gotemplate.Must(gotemplate.New("otel").Parse(otelTemplateText))

func executeTelemetry(conf dataplane.Configuration) []byte {
	if conf.Telemetry.Endpoint != "" {
		return execute(otelTemplate, conf.Telemetry)
	}

	return nil
}
