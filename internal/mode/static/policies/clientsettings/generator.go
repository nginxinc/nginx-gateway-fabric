package clientsettings

import (
	"bytes"
	"fmt"
	"text/template"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
)

// NewClientSettingsGeneratorFunc returns a function that generates configuration as []byte for a ClientSettingsPolicy.
func NewClientSettingsGeneratorFunc() func(policy policies.Policy) []byte {
	return func(policy policies.Policy) []byte {
		csp, ok := policy.(*ngfAPI.ClientSettingsPolicy)
		if !ok {
			panic(fmt.Sprintf("expected ClientSettingsPolicy, got: %T", policy))
		}

		tmpl := template.Must(template.New("client settings policy").Parse(clientSettingsTemplate))

		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, csp.Spec); err != nil {
			panic(fmt.Errorf("failed to execute template for client settings policy: %w", err))
		}

		return buf.Bytes()
	}
}

var clientSettingsTemplate = `
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
