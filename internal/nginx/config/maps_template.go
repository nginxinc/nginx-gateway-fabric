package config

var mapsTemplateText = `
{{ range $m := . }}
map {{ $m.Source }} {{ $m.Variable }} {
    {{ range $p := $m.Parameters }}
    {{ $p.Value }} {{ $p.Result }};
    {{ end }}
}
{{- end }}
`
