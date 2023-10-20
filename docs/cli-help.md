# Command-line Help

This document describes the commands available in the `gateway` binary of the `nginx-gateway` container.

## Static Mode

This command configures NGINX in the scope of a single Gateway resource.

Usage:

```text
  gateway static-mode [flags]
```

Flags:

| Name                         | Type     | Description                                                                                                                                                                                                                                                                                                                                                                                       |
|------------------------------|----------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `gateway-ctlr-name`          | `string` | The name of the Gateway controller. The controller name must be of the form: `DOMAIN/PATH`. The controller's domain is `gateway.nginx.org`.                                                                                                                                                                                                                                                       |
| `gatewayclass`               | `string` | The name of the GatewayClass resource. Every NGINX Gateway Fabric must have a unique corresponding GatewayClass resource.                                                                                                                                                                                                                                                                                |
| `gateway`                    | `string` | The namespaced name of the Gateway resource to use. Must be of the form: `NAMESPACE/NAME`. If not specified, the control plane will process all Gateways for the configured GatewayClass. However, among them, it will choose the oldest resource by creation timestamp. If the timestamps are equal, it will choose the resource that appears first in alphabetical order by {namespace}/{name}. |
| `config`                     | `string` | The name of the NginxGateway resource to be used for this controller's dynamic configuration. Lives in the same Namespace as the controller.                                                                                                                                                                                                                                                      |
| `service`                    | `string` | The name of the Service that fronts this NGINX Gateway Fabric Pod. Lives in the same Namespace as the controller. |
| `metrics-disable`            | `bool`   | Disable exposing metrics in the Prometheus format. (default false)                                                                                                                                                                                                                                                                                                                                |
| `metrics-listen-port`        | `int`    | Sets the port where the Prometheus metrics are exposed. Format: `[1024 - 65535]` (default `9113`)                                                                                                                                                                                                                                                                                                 |
| `metrics-secure-serving`     | `bool`   | Configures if the metrics endpoint should be secured using https. Please note that this endpoint will be secured with a self-signed certificate. (default false)                                                                                                                                                                                                                                  |
| `update-gatewayclass-status` | `bool`   | Update the status of the GatewayClass resource. (default true)                                                                                                                                                                                                                                                                                                                                    |
| `health-disable`             | `bool`   | Disable running the health probe server. (default false)                                                                                                                                                                                                                                                                                                                                          |
| `health-port`                | `int`    | Set the port where the health probe server is exposed. Format: `[1024 - 65535]` (default `8081`)                                                                                                                                                                                                                                                                                                  |
| `leader-election-disable`    | `bool`   | Disable leader election. Leader election is used to avoid multiple replicas of the NGINX Gateway Fabric reporting the status of the Gateway API resources. If disabled, all replicas of NGINX Gateway Fabric will update the statuses of the Gateway API resources. (default false)                                                                                                       |
| `leader-election-lock-name`  | `string` | The name of the leader election lock. A Lease object with this name will be created in the same Namespace as the controller. (default "nginx-gateway-leader-election-lock")                                                                                                                                                                                                                       |

## Sleep

This command sleeps for specified duration and exits.

Usage:

```text
Usage:
  gateway sleep [flags]
```

| Name     | Type            | Description                                                                                           |
|----------|-----------------|-------------------------------------------------------------------------------------------------------|
| duration | `time.Duration` | Set the duration of sleep. Must be parsable by [`time.ParseDuration`][parseDuration]. (default `30s`) |


[parseDuration]:https://pkg.go.dev/time#ParseDuration
