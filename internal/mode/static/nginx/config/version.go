package config

import (
	gotemplate "text/template"
)

var versionTemplate = gotemplate.Must(gotemplate.New("version").Parse(versionTemplateText))

func executeVersion(version int) []byte {
	return execute(versionTemplate, version)
}
