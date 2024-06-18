package config

const versionTemplateText = `
server {
    listen unix:/var/run/nginx/nginx-config-version.sock;
    access_log off;

    location /version {
        return 200 {{.}};
    }
}
`
