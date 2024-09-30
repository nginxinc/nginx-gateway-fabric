package config

const mainConfigTemplateText = `
{{ if .TelemetryEnabled -}}
load_module modules/ngx_otel_module.so;
{{ end -}}

{{ range $i := .Includes -}}
include {{ $i.Name }};
{{ end -}}
`
