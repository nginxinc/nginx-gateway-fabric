package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var versionTemplate = gotemplate.Must(gotemplate.New("version").Parse(versionTemplateText))

func executeVersion(conf dataplane.Configuration) []executeResult {
	result := executeResult{
		dest: configVersionFile,
		data: helpers.MustExecuteTemplate(versionTemplate, conf.Version),
	}

	return []executeResult{result}
}
