# Command-line Arguments

The table below describes the command-line arguments of the `gateway` binary from the `nginx-kubernetes-gateway` container.

| Name | Type | Description |
|-|-|-|
|`gateway-ctlr-name` | `string` |  The name of the Gateway controller. The controller name must be of the form: `DOMAIN/PATH`. The controller's domain is `k8s-gateway.nginx.org`. |
|`gatewayclass`| `string` | The name of the GatewayClass resource. Every NGINX Gateway must have a unique corresponding GatewayClass resource. |
