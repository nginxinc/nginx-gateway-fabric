package config

const otelTemplateText = `
otel_exporter {
	endpoint {{ .Endpoint }};
	{{- if .Interval }}
	interval {{ .Interval }};
	{{- end }}
	{{- if .BatchSize }}
	batch_size {{ .BatchSize }};
	{{- end }}
	{{- if .BatchCount }}
	batch_count {{ .BatchCount }};
	{{- end }}
}

otel_service_name {{ .ServiceName }};

{{- range $attr := .SpanAttributes }}
otel_span_attr "{{ $attr.Key }}" "{{ $attr.Value }}";
{{- end }}
`
