package config

import (
	"bytes"
	"fmt"
	"text/template"
)

var httpTemplate = `{{ range $u := .Upstreams }}
upstream {{ $u.Name }} {
	{{ range $server := $u.Servers }} 
	server {{ $server.Address }};
	{{ end }}
}
{{ end }}
{{ range $s := .Servers }}
	{{ if $s.IsDefaultSSL }}
server {
	listen 443 ssl default_server;

	ssl_reject_handshake on;
}
	{{ else if $s.IsDefaultHTTP }}
server {
	listen 80 default_server;
	
	default_type text/html;
	return 404;
}
	{{ else }}
server {
		{{ if $s.SSL }}
	listen 443 ssl;
	ssl_certificate {{ $s.SSL.Certificate }};
	ssl_certificate_key {{ $s.SSL.CertificateKey }};

	if ($ssl_server_name != $host) {
		return 421;
	}
		{{ end }}

	server_name {{ $s.ServerName }};

		{{ range $l := $s.Locations }}
	location {{ $l.Path }} {
		{{ if $l.Internal }}
		internal;
		{{ end }}
		
		proxy_set_header Host $host;

		{{ if $l.HTTPMatchVar }}
		set $http_matches {{ $l.HTTPMatchVar | printf "%q" }};
		js_content httpmatches.redirect;
		{{ end }}

		{{ if $l.ProxyPass }}
		proxy_pass {{ $l.ProxyPass }}$request_uri;
		{{ end }}
	}
		{{ end }}
}
	{{ end }}
{{ end }}
`

// templateExecutor generates NGINX configuration using a template.
// Template parsing or executing errors can only occur if there is a bug in the template, so they are handled with panics.
// For now, we only generate configuration with NGINX http servers and upstreams, but in the future we will also need to generate
// the main NGINX configuration file and stream servers.
type templateExecutor struct {
	httpTemplate *template.Template
}

func newTemplateExecutor() *templateExecutor {
	t, err := template.New("server").Parse(httpTemplate)
	if err != nil {
		panic(fmt.Errorf("failed to parse http template: %w", err))
	}

	return &templateExecutor{httpTemplate: t}
}

func (e *templateExecutor) ExecuteForHTTP(http http) []byte {
	var buf bytes.Buffer

	err := e.httpTemplate.Execute(&buf, http)
	if err != nil {
		panic(fmt.Errorf("failed to execute http template: %w", err))
	}

	return buf.Bytes()
}
