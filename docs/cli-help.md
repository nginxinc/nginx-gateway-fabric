# Command-line Help

This document describes the commands available in the `gateway` binary of `nginx-kubernetes-gateway` container.

## Control Plane 

This command starts the control plane.

Usage:

```
  gateway control-plane [flags]
```

Flags:

| Name | Type | Description |
|-|-|-|
| `gateway-ctlr-name` | `string` |  The name of the Gateway controller. The controller name must be of the form: `DOMAIN/PATH`. The controller's domain is `k8s-gateway.nginx.org`. |
| `gatewayclass`      | `string` | The name of the GatewayClass resource. Every NGINX Gateway must have a unique corresponding GatewayClass resource. |
