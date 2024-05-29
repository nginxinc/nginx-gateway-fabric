package clientsettings

import (
	"text/template"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
)

var tmpl = template.Must(template.New("client settings policy").Parse(clientSettingsTemplate))

const clientSettingsTemplate = `
{{- if .Body }}
	{{- if .Body.MaxSize }}
client_max_body_size {{ .Body.MaxSize }};
	{{- end }}
	{{- if .Body.Timeout }}
client_body_timeout {{ .Body.Timeout }};
	{{- end }}
{{- end }}
{{- if .KeepAlive }}
	{{- if .KeepAlive.Requests }}
keepalive_requests {{ .KeepAlive.Requests }};
	{{- end }}
	{{- if .KeepAlive.Time }}
keepalive_time {{ .KeepAlive.Time }};
	{{- end }}
    {{- if .KeepAlive.Timeout }}
        {{- if and .KeepAlive.Timeout.Server .KeepAlive.Timeout.Header }}
keepalive_timeout {{ .KeepAlive.Timeout.Server }} {{ .KeepAlive.Timeout.Header }};
        {{- else if .KeepAlive.Timeout.Server }}
keepalive_timeout {{ .KeepAlive.Timeout.Server }};
        {{- end }}
    {{- end }}
{{- end }}
`

// Generate generates configuration as []byte for a ClientSettingsPolicy.
func Generate(policy policies.Policy, _ *policies.GlobalSettings) []byte {
	csp := helpers.MustCastObject[*ngfAPI.ClientSettingsPolicy](policy)

	return helpers.MustExecuteTemplate(tmpl, csp.Spec)
}
