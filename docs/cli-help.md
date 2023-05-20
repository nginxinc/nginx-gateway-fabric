# Command-line Help

This document describes the commands available in the `gateway` binary of `nginx-kubernetes-gateway` container.

## Static Mode

This command configures NGINX in the scope of a single Gateway resource. In case of multiple Gateway resources created
in the cluster, NGINX Kubernetes Gateway will use a deterministic conflict resolution strategy: it will choose the
oldest resource by creation timestamp. If the timestamps are equal, NGINX Kubernetes Gateway will choose the resource
that appears first in alphabetical order by “{namespace}/{name}”. We might support multiple Gateway resources. Please
share your use case with us if you're interested in that support.

Usage:

```
  gateway static-mode [flags]
```

Flags:

| Name | Type | Description |
|-|-|-|
| `gateway-ctlr-name` | `string` |  The name of the Gateway controller. The controller name must be of the form: `DOMAIN/PATH`. The controller's domain is `k8s-gateway.nginx.org`. |
| `gatewayclass`      | `string` | The name of the GatewayClass resource. Every NGINX Gateway must have a unique corresponding GatewayClass resource. |
