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

const mgmtConfigTemplateText = `
mgmt {
	license_token /etc/nginx/license/license.jwt;
	{{- if .Endpoint }}
	usage_report endpoint={{ .Endpoint }};
	{{- end }}
	{{- if .Resolver }}
	resolver {{ .Resolver }};
	{{- end }}
	{{- if .DeploymentCtxFile }}
	deployment_context {{ .DeploymentCtxFile }};
	{{- end }}
	{{- if .SkipVerify }}
	ssl_verify off;
	{{- end }}
	{{- if .CACertExists }}
	ssl_trusted_certificate /etc/nginx/usage-certs/ca/ca.crt;
	{{- end }}
	{{- if .ClientSSLCertExists }}
	ssl_certificate /etc/nginx/usage-certs/client/tls.crt;
	ssl_certificate_key /etc/nginx/usage-certs/client/tls.key;
	{{- end }}
}
`
