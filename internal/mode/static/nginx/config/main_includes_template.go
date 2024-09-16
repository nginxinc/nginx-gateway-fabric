package config

const mainIncludesTemplateText = `
{{- if .Telemetry.Endpoint }}load_module modules/ngx_otel_module.so;{{ end -}}

error_log stderr {{ .Logging.ErrorLevel }};
`
