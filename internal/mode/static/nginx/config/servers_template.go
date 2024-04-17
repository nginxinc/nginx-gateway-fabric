package config

const serversTemplateText = `
js_preload_object matches from /etc/nginx/conf.d/matches.json;
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

        {{- if $l.Return }}
        return {{ $l.Return.Code }} "{{ $l.Return.Body }}";
        {{- end }}

        {{- if $l.HTTPMatchKey }}
        set $match_key {{ $l.HTTPMatchKey }};
        js_content httpmatches.redirect;
        {{- end }}

        {{ $proxyOrGRPC := "proxy" }}{{ if $l.IsGRPC }}{{ $proxyOrGRPC = "grpc" }}{{ end }}

        {{- if $l.IsGRPC }}
        error_page 400 = @grpc_internal;
        error_page 401 = @grpc_unauthenticated;
        error_page 403 = @grpc_permission_denied;
        error_page 404 = @grpc_unimplemented;
        error_page 429 = @grpc_unavailable;
        error_page 502 = @grpc_unavailable;
        error_page 503 = @grpc_unavailable;
        error_page 504 = @grpc_unavailable;
        error_page 405 = @grpc_internal;
        error_page 408 = @grpc_deadline_exceeded;
        error_page 413 = @grpc_resource_exhausted;
        error_page 414 = @grpc_resource_exhausted;
        error_page 415 = @grpc_internal;
        error_page 426 = @grpc_internal;
        error_page 495 = @grpc_unauthenticated;
        error_page 496 = @grpc_unauthenticated;
        error_page 497 = @grpc_internal;
        error_page 500 = @grpc_internal;
        error_page 501 = @grpc_internal;
        {{- end }}

        {{- if $l.ProxyPass -}}
            {{ range $h := $l.ProxySetHeaders }}
        {{ $proxyOrGRPC }}_set_header {{ $h.Name }} "{{ $h.Value }}";
            {{- end }}
        {{ $proxyOrGRPC }}_pass {{ $l.ProxyPass }};
        proxy_http_version 1.1;
            {{- if $l.ProxySSLVerify }}
        {{ $proxyOrGRPC }}_ssl_verify on;
        {{ $proxyOrGRPC }}_ssl_name {{ $l.ProxySSLVerify.Name }};
        {{ $proxyOrGRPC }}_ssl_trusted_certificate {{ $l.ProxySSLVerify.TrustedCertificate }};
            {{- end }}
        {{- end }}
    }
        {{ end }}

        {{- if $s.HTTP2 }}
    location @grpc_deadline_exceeded {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 4;
        add_header grpc-message 'deadline exceeded';
        return 204;
    }

    location @grpc_permission_denied {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 7;
        add_header grpc-message 'permission denied';
        return 204;
    }

    location @grpc_resource_exhausted {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 8;
        add_header grpc-message 'resource exhausted';
        return 204;
    }

    location @grpc_unimplemented {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 12;
        add_header grpc-message unimplemented;
        return 204;
    }

    location @grpc_internal {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 13;
        add_header grpc-message 'internal error';
        return 204;
    }

    location @grpc_unavailable {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 14;
        add_header grpc-message unavailable;
        return 204;
    }

    location @grpc_unauthenticated {
        default_type application/grpc;
        add_header content-type application/grpc;
        add_header grpc-status 16;
        add_header grpc-message unauthenticated;
        return 204;
    }
    {{- end }}
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
