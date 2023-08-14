package config

var serversTemplateText = `
{{- range $s := . -}}
    {{ if $s.IsDefaultSSL -}}
server {
    listen {{ $s.Port }} ssl default_server;

    ssl_reject_handshake on;
}
    {{- else if $s.IsDefaultHTTP }}
server {
    listen {{ $s.Port }} default_server;

    default_type text/html;
    return 404;
}
    {{- else }}
server {
        {{- if $s.SSL }}
    listen {{ $s.Port }} ssl;
    ssl_certificate {{ $s.SSL.Certificate }};
    ssl_certificate_key {{ $s.SSL.CertificateKey }};

    if ($ssl_server_name != $host) {
        return 421;
    }
        {{- else }}
    listen {{ $s.Port }};
        {{- end }}

    server_name {{ $s.ServerName }};

        {{ range $l := $s.Locations }}
    location {{ $l.Path }} {
        {{ if $l.Internal -}}
        internal;
        {{ end }}

        {{- if $l.Return -}}
        return {{ $l.Return.Code }} "{{ $l.Return.Body }}";
        {{ end }}

        {{- if $l.HTTPMatchVar -}}
        set $http_matches {{ $l.HTTPMatchVar | printf "%q" }};
        js_content httpmatches.redirect;
        {{ end }}

        {{- if $l.ProxyPass -}}
            {{ range $h := $l.ProxySetHeaders }}
        proxy_set_header {{ $h.Name }} "{{ $h.Value }}";
            {{- end }}
        proxy_set_header Host $gw_api_compliant_host;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection $connection_upgrade;
        proxy_pass {{ $l.ProxyPass }}$request_uri;
        {{- end }}
    }
        {{ end }}
}
    {{- end }}
{{ end }}
server {
    listen unix:/var/lib/nginx/nginx-502-server.sock;
    access_log off;

    return 502;
}

server {
    listen unix:/var/lib/nginx/nginx-500-server.sock;
    access_log off;

    return 500;
}
`
