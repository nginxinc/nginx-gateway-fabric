package config

const baseHTTPTemplateText = `
{{- if .HTTP2 }}http2 on;{{ end }}
`
