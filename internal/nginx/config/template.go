package config

import (
	"bytes"
	"fmt"
	"text/template"
)

var serverTemplate = `server {
	server_name {{ .ServerName }};

	{{ range $l := .Locations }}
	location {{ $l.Path }} {
		proxy_set_header Host $host;
		proxy_pass {{ $l.ProxyPass }};
	}
	{{ end }}
}
`

// templateExecutor generates NGINX configuration using a template.
// Template parsing or executing errors can only occur if there is a bug in the template, so they are handled with panics.
// For now, we only generate configuration with NGINX server, but in the future we will also need to generate
// the main NGINX configuration file, upstreams.
type templateExecutor struct {
	serverTemplate *template.Template
}

func newTemplateExecutor() *templateExecutor {
	t, err := template.New("server").Parse(serverTemplate)
	if err != nil {
		panic(fmt.Errorf("failed to parse server template: %w", err))
	}

	return &templateExecutor{serverTemplate: t}
}

func (e *templateExecutor) ExecuteForServer(s server) []byte {
	var buf bytes.Buffer

	err := e.serverTemplate.Execute(&buf, s)
	if err != nil {
		panic(fmt.Errorf("failed to execute server template: %w", err))
	}

	return buf.Bytes()
}
