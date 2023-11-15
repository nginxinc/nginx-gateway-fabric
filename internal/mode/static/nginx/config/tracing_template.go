package config

var otelTemplateText = `
otel_exporter {
    endpoint {{ .ExporterEndpoint }};
    interval {{ .Interval }};
    batch_size  {{ .BatchSize }};
    batch_count {{ .BatchCount }};
}
otel_trace {{ if .Enabled }}on{{ else }}off{{ end }};
otel_service_name {{ .ServiceName }};
`
