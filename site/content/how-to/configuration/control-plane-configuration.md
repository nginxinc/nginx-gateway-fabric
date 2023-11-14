---
title: "Control Plane Configuration"
description: "Learn how to dynamically update the NGINX Gateway Fabric control plane configuration."
weight: 100
toc: true
docs: "DOCS-000"
---

## Overview

NGINX Gateway Fabric offers a way to update the control plane configuration dynamically without the need for a
restart. The control plane configuration is stored in the NginxGateway custom resource. This resource is created
during the installation of NGINX Gateway Fabric.

If using manifests, the default name of the resource is `nginx-gateway-config`. If using Helm, the default name
of the resource is `<release-name>-config`. It is deployed in the same Namespace as the controller
(default `nginx-gateway`).

The control plane only watches this single instance of the custom resource. If the resource is invalid per the OpenAPI
schema, the Kubernetes API server will reject the changes. If the resource is deleted or deemed invalid by NGINX
Gateway Fabric, a warning Event is created in the `nginx-gateway` Namespace, and the default values will be used by
the control plane for its configuration. Additionally, the control plane updates the status of the resource (if it exists)
to reflect whether it is valid or not.

### Spec

| name    | description                                                     | type                     | required |
|---------|-----------------------------------------------------------------|--------------------------|----------|
| logging | Logging defines logging related settings for the control plane. | [logging](#speclogging) | no       |

### Spec.Logging

| name  | description                                                            | type   | required |
|-------|------------------------------------------------------------------------|--------|----------|
| level | Level defines the logging level. Supported values: info, debug, error. | string | no       |

## Viewing and Updating the Configuration

> For the following examples, the name `nginx-gateway-config` should be updated to the name of the resource that
> was created by your installation.

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

To view the status of the configuration:

```shell
kubectl -n nginx-gateway describe nginxgateways nginx-gateway-config
```
