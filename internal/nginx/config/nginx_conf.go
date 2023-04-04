package config

import (
	"text/template"
)

var nginxConfTemplate = template.Must(template.New("nginx-conf").Parse(nginxConfTemplateText))

func executeNginxConf(configGeneration int) []byte {
	return execute(nginxConfTemplate, configGeneration)
}
