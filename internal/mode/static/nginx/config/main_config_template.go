package config

const mainConfigTemplateText = `
{{ if .Conf.Telemetry.Endpoint -}}
load_module modules/ngx_otel_module.so;
{{ end -}}

error_log stderr {{ .Conf.Logging.ErrorLevel }};

{{ range $i := .Includes -}}
include {{ $i.Name }};
{{ end -}}
`
