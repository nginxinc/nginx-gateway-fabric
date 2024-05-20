package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var baseHTTPTemplate = gotemplate.Must(gotemplate.New("baseHttp").Parse(baseHTTPTemplateText))

func executeBaseHTTPConfig(conf dataplane.Configuration) []executeResult {
	result := executeResult{
		dest: httpConfigFile,
		data: execute(baseHTTPTemplate, conf.BaseHTTPConfig),
	}

	return []executeResult{result}
}
