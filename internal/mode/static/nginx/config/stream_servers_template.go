package config

const streamServersTemplateText = `
{{- range $s := . }}
server {
	listen {{ $s.Listen }};

	{{- if $s.ProxyPass }}
	proxy_pass {{ $s.Destination }};
	{{- else }}
	pass {{ $s.Destination }};
	{{- end }}

	{{- if $s.SSLPreread }}
	ssl_preread on;
	{{- end }}
}
{{- end }}
`
