package config

import (
	gotemplate "text/template"

	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
)

var versionTemplate = gotemplate.Must(gotemplate.New("version").Parse(versionTemplateText))

func executeVersion(version int) []byte {
	return helpers.MustExecuteTemplate(versionTemplate, version)
}
