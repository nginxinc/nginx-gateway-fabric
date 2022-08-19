# Command-line Arguments

The table belows describes the command-line arguments of the `gateway` binary from the `nginx-kubernetes-gateway` container.

| Name | Type | Description |
|-|-|-|
|`gateway-ctlr-name` | `string` |  The name of the Gateway controller. The controller name must be of the form: `DOMAIN/NAMESPACE/NAME`. The controller's domain is `k8s-gateway.nginx.org`; the namespace is `nginx-ingress`. |
|`gatewayclass`| `string` | The name of the GatewayClass resource. Every NGINX Gateway must have a unique corresponding GatewayClass resource. |
