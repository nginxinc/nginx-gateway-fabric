# Command-line Help

This document describes the commands available in the `gateway` binary of `nginx-kubernetes-gateway` container.

## Static Mode

This command configures NGINX in the scope of a single Gateway resource.

Usage:

```text
  gateway static-mode [flags]
```

Flags:

| Name                         | Type     | Description                                                                                                                                                                                                                                                                                                                                                                                       |
|------------------------------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `gateway-ctlr-name`          | `string` | The name of the Gateway controller. The controller name must be of the form: `DOMAIN/PATH`. The controller's domain is `gateway.nginx.org`. |
| `gatewayclass`               | `string` | The name of the GatewayClass resource. Every NGINX Gateway must have a unique corresponding GatewayClass resource. |
| `gateway`                    | `string` | The namespaced name of the Gateway resource to use. Must be of the form: `NAMESPACE/NAME`. If not specified, the control plane will process all Gateways for the configured GatewayClass. However, among them, it will choose the oldest resource by creation timestamp. If the timestamps are equal, it will choose the resource that appears first in alphabetical order by {namespace}/{name}. |
| `config`                     | `string` | The name of the NginxGateway resource to be used for this controller's dynamic configuration. Lives in the same Namespace as the controller. |
| `enable-metrics`             | `bool`   | Enable metrics reporting in the Prometheus format. (default false) |
| `metrics-listen-port`        | `int`    | Sets the port where the Prometheus metrics are exposed. Format: `[1024 - 65535]` (default `9113`) |
| `metrics-secure-serving`     | `bool`   | Configures if the metrics endpoint should be secured using https. Please note that this endpoint will be secured with a self-signed certificate. (default false) |
