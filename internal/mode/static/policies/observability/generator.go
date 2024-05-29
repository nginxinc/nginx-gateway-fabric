package observability

import (
	"fmt"
	"text/template"

	ngfAPI "github.com/nginxinc/nginx-gateway-fabric/apis/v1alpha1"
	"github.com/nginxinc/nginx-gateway-fabric/internal/framework/helpers"
	"github.com/nginxinc/nginx-gateway-fabric/internal/mode/static/policies"
)

var tmpl = template.Must(template.New("observability policy").Parse(observabilityTemplate))

const observabilityTemplate = `
{{- if .Tracing }}
otel_trace {{ .Strategy }};
  {{- if .Tracing.Context }}
otel_trace_context {{ .Tracing.Context }};
  {{- end }}
  {{- if .Tracing.SpanName }}
otel_span_name "{{ .Tracing.SpanName }}";
  {{- end }}
  {{- range $attr := .Tracing.SpanAttributes }}
otel_span_attr "{{ $attr.Key }}" "{{ $attr.Value }}";
  {{- end }}
  {{- range $attr := .GlobalSpanAttributes }}
otel_span_attr "{{ $attr.Key }}" "{{ $attr.Value }}";
  {{- end }}
{{- end }}
`

// Generate generates configuration as []byte for an ObservabilityPolicy.
func Generate(policy policies.Policy, globalSettings *policies.GlobalSettings) []byte {
	obs := helpers.MustCastObject[*ngfAPI.ObservabilityPolicy](policy)

	var strategy string
	if obs.Spec.Tracing != nil {
		switch obs.Spec.Tracing.Strategy {
		case ngfAPI.TraceStrategyParent:
			strategy = "$otel_parent_sampled"
		case ngfAPI.TraceStrategyRatio:
			strategy = "on"
			if obs.Spec.Tracing.Ratio != nil {
				if *obs.Spec.Tracing.Ratio > 0 {
					strategy = CreateRatioVarName(obs)
				} else {
					strategy = "off"
				}
			}
		default:
			strategy = "off"
		}
	}

	var spanAttributes []ngfAPI.SpanAttribute
	if globalSettings != nil {
		spanAttributes = globalSettings.TracingSpanAttributes
	}

	fields := map[string]interface{}{
		"Tracing":              obs.Spec.Tracing,
		"Strategy":             strategy,
		"GlobalSpanAttributes": spanAttributes,
	}

	return helpers.MustExecuteTemplate(tmpl, fields)
}

// CreateRatioVarName builds a variable name for an ObservabilityPolicy to be used with
// ratio-based trace sampling.
func CreateRatioVarName(policy *ngfAPI.ObservabilityPolicy) string {
	return fmt.Sprintf("$otel_ratio_%d", *policy.Spec.Tracing.Ratio)
}
