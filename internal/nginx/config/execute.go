package config

import (
	"bytes"
	"text/template"
)

// executes the template with the given data.
func execute(template *template.Template, data interface{}) []byte {
	var buf bytes.Buffer

	err := template.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
