package config

import (
	gotemplate "text/template"
)

var versionTemplate = gotemplate.Must(gotemplate.New("version").Parse(versionTemplateText))

func executeVersion(version int) []byte {
	data := make(map[string]int, 1)
	data["Version"] = version
	return execute(versionTemplate, data)
}
