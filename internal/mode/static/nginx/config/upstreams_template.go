package config

// FIXME(kate-osborn): Dynamically calculate upstream zone size based on the number of upstreams.
// 512k will support up to 648 http upstream servers for OSS.
// NGINX Plus needs 1m to support roughly the same amount of http servers (556 upstream servers).
// For stream upstream servers, 512k will support 576 in OSS and 1m will support 991 in NGINX Plus
// https://github.com/nginxinc/nginx-gateway-fabric/issues/483
const upstreamsTemplateText = `
{{ range $u := . }}
upstream {{ $u.Name }} {
    # if the keepalive directive us present, it is necessary to activate the load balancing method before the directive
    random two least_conn;
    {{ if $u.ZoneSize -}}
    zone {{ $u.Name }} {{ $u.ZoneSize }};
    {{ end -}}
    {{ range $server := $u.Servers }}
    server {{ $server.Address }};
    {{- end }}
    {{ if $u.KeepAliveConnections -}}
    keepalive {{ $u.KeepAliveConnections }};
    {{- end }}
    {{ if $u.KeepAliveRequests -}}
    keepalive_requests {{ $u.KeepAliveRequests }};
    {{- end }}
    {{ if $u.KeepAliveTime -}}
    keepalive_time {{ $u.KeepAliveTime }};
    {{- end }}
    {{ if $u.KeepAliveTimeout -}}
    keepalive_timeout {{ $u.KeepAliveTimeout }};
    {{- end }}
}
{{ end -}}
`

const streamUpstreamsTemplateText = `
{{ range $u := . }}
upstream {{ $u.Name }} {
    random two least_conn;
    {{ if $u.ZoneSize -}}
    zone {{ $u.Name }} {{ $u.ZoneSize }};
    {{ end -}}
    {{ range $server := $u.Servers }}
    server {{ $server.Address }};
    {{- end }}
}
{{ end -}}
`
