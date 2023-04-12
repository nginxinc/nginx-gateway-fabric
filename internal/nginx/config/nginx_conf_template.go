package config

var nginxConfTemplateText = `# config version: {{ . }}
load_module /usr/lib/nginx/modules/ngx_http_js_module.so;

events {}

pid /etc/nginx/nginx.pid;

error_log /var/log/nginx/error.log debug;

http {
    include /etc/nginx/conf.d/*.conf;
    js_import /usr/lib/nginx/modules/njs/httpmatches.js;

    log_format  main  '$remote_addr - $remote_user [$time_local] "$request" '
                      '$status $body_bytes_sent "$http_referer" '
                      '"$http_user_agent" "$http_x_forwarded_for" ';

    access_log /var/log/nginx/access.log  main;

    # stub status API
    # needed by the agent in order to collect metrics
    server {
        listen 127.0.0.1:8082;
        location /api {
            stub_status;
            allow 127.0.0.1;
            deny all;
        }
    }

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
}`
