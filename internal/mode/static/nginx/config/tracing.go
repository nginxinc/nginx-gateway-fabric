package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var otelTemplate = gotemplate.Must(gotemplate.New("otel").Parse(otelTemplateText))

func executeTracing(conf dataplane.Configuration) []byte {
	if conf.Tracing.ExporterEndpoint != "" {
		return execute(otelTemplate, conf.Tracing)
	}
	return nil
}
