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
        {{- range $r := $l.Rewrites }}
        rewrite {{ $r }};
        {{- end }}
		
		{{- if $l.MirrorPath }}
		mirror $l.MirrorPath;
		{{- end }}

        {{- if $l.Return }}
        return {{ $l.Return.Code }} "{{ $l.Return.Body }}";
        {{- end }}

        {{- if $l.HTTPMatchVar }}
        set $http_matches {{ $l.HTTPMatchVar | printf "%q" }};
        js_content httpmatches.redirect;
        {{- end }}

        {{- if $l.ProxyPass -}}
            {{ range $h := $l.ProxySetHeaders }}
        proxy_set_header {{ $h.Name }} "{{ $h.Value }}";
            {{- end }}
        proxy_http_version 1.1;
        proxy_pass {{ $l.ProxyPass }};
            {{- if $l.ProxySSLVerify }}
        proxy_ssl_verify on;
        proxy_ssl_name {{ $l.ProxySSLVerify.Name }};
        proxy_ssl_trusted_certificate {{ $l.ProxySSLVerify.TrustedCertificate }};
            {{- end }}
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
