package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/state/dataplane"
)

var mainIncludesTemplate = gotemplate.Must(gotemplate.New("mainIncludes").Parse(mainIncludesTemplateText))

func executeMainIncludesConfig(conf dataplane.Configuration) executeResult {
	result := executeResult{
		dest: mainIncludesConfigFile,
		data: helpers.MustExecuteTemplate(mainIncludesTemplate, conf),
	}

	return result
}
