package config

var versionTemplateText = `server {
    listen unix:/var/run/nginx/nginx-config-version.sock;
	access_log off;

    location /configVersion {
        return 200 {{.Version}};
    }
}
map $http_x_expected_config_version $config_version_mismatch {
	"{{.Version}}" "";
	default "mismatch";
}`
