# Control Plane Configuration

This document describes how to dynamically update the NGINX Kubernetes Gateway control plane configuration.

## Overview

NGINX Kubernetes Gateway offers a way to update the control plane configuration dynamically without the need for a
restart. These configuration options include:

| Option        | Available values   | Default value |
|---------------|--------------------|---------------|
| Logging Level | info, debug, error | info          |


The control plane configuration is stored in the NginxGateway custom resource. This resource is created during the
installation of NGINX Kubernetes Gateway. The default name of the resource is `nginx-gateway-config` and is deployed
in the same Namespace as the controller (`nginx-gateway`).

The control plane only watches this single instance of the custom resource. If the resource is deleted or invalid, an
error is emitted and the default values will be used by the control plane for its configuration.

## Viewing and Updating the Configuration

To view the current configuration:

```shell
kubectl -n nginx-gateway get nginxgateways nginx-gateway-config -o yaml
```

To update the configuration:

```shell
kubectl -n nginx-gateway edit nginxgateways nginx-gateway-config
```

This will open the configuration in your default editor. You can then update and save the configuration, which is
applied automatically to the control plane.
