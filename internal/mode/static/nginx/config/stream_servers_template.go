package config

const streamServersTemplateText = `
{{ $proxyProtocol := "" }}
{{ if $.RewriteClientIP.ProxyProtocol }}{{ $proxyProtocol = " proxy_protocol" }}{{ end }}
{{- range $s := .Servers }}
server {
	{{- if and ($.IPFamily.IPv4) (not $s.IsSocket) }}
    listen {{ $s.Listen }};
	{{- else if and ($.IPFamily.IPv4) ( $s.IsSocket)}}
	listen {{ $s.Listen }}{{ $proxyProtocol }};
	{{- end }}
	{{- if and ($.IPFamily.IPv6) (not $s.IsSocket) }}
    listen [::]:{{ $s.Listen }};
	{{- end }}

	{{- if and ($s.IsSocket) ($.RewriteClientIP.ProxyProtocol) }}
		{{- range $cidr := $.RewriteClientIP.RealIPFrom }}
    set_real_ip_from {{ $cidr }};
        {{- end}}
	{{- end }}
	{{- if $.Plus }}
    status_zone {{ $s.StatusZone }};
    {{- end }}

	{{- if $s.ProxyPass }}
    proxy_pass {{ $s.ProxyPass }};
	{{- end }}
	{{- if $s.Pass }}
    pass {{ $s.Pass }};
	{{- end }}
	{{- if $s.SSLPreread }}
    ssl_preread on;
	{{- end }}
}
{{- end }}

server {
    listen unix:/var/run/nginx/connection-closed-server.sock;
    return "";
}
`
