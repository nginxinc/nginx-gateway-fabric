package config

const streamServersTemplateText = `
{{- range $s := . }}
server {
	listen {{ $s.Listen }};

	{{- if $s.ProxyPass }}
	proxy_pass {{ $s.ProxyPass }};
	{{- end }}

	{{- if $s.Pass }}
	pass {{ $s.Pass }};
	{{- end }}

	{{- if $s.SSLPreread }}
	ssl_preread on;
	{{- end }}
}
{{- end }}
`
