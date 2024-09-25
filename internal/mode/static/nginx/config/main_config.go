package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/nginx/config/shared"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var mainConfigTemplate = gotemplate.Must(gotemplate.New("main").Parse(mainConfigTemplateText))

type mainConfig struct {
	Includes         []shared.Include
	TelemetryEnabled bool
}

func executeMainConfig(conf dataplane.Configuration) []executeResult {
	includes := createIncludesFromSnippets(conf.MainSnippets)

	mc := mainConfig{
		TelemetryEnabled: conf.Telemetry.Endpoint != "",
		Includes:         includes,
	}

	results := make([]executeResult, 0, len(includes)+1)
	results = append(
		results, executeResult{
			dest: mainIncludeFile,
			data: helpers.MustExecuteTemplate(mainConfigTemplate, mc),
		},
	)
	results = append(results, createIncludeExecuteResults(includes)...)

	return results
}