package config

// FIXME(kate-osborn): Dynamically calculate upstream zone size based on the number of upstreams.
var upstreamsTemplateText = `
{{ range $u := . }}
upstream {{ $u.Name }} {
    random two least_conn;
    zone {{ $u.Name }} 512k;
    {{ range $server := $u.Servers }} 
    server {{ $server.Address }};
    {{- end }}
}
{{ end }}
`
