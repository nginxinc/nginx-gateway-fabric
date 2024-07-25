---
title: "Control plane configuration"
weight: 300
toc: true
docs: "DOCS-1416"
---

Learn how to dynamically update the Gateway Fabric control plane configuration.

## Overview

NGINX Gateway Fabric can dynamically update the control plane configuration without restarting. The control plane configuration is stored in the NginxGateway custom resource, created during the installation of NGINX Gateway Fabric.

NginxGateway is deployed in the same namespace as the controller (Default: `nginx-gateway`). The resource's default name is based on your [installation method]({{<relref "/installation/installing-ngf">}}):

- Helm: `<release-name>-config`
- Manifests: `nginx-gateway-config`

The control plane only watches this single instance of the custom resource.

If the resource is invalid to the OpenAPI schema, the Kubernetes API server will reject the changes. If the resource is deleted or deemed invalid by NGINX Gateway Fabric, a warning event is created in the `nginx-gateway` namespace, and the default values will be used by the control plane for its configuration.

Additionally, the control plane updates the status of the resource (if it exists) to reflect whether it is valid or not.

### Spec

{{< bootstrap-table "table table-striped table-bordered" >}}
| name    | description                                                     | type                     | required |
|---------|-----------------------------------------------------------------|--------------------------|----------|
| logging | Logging defines logging related settings for the control plane. | [logging](#speclogging) | no       |
{{< /bootstrap-table >}}

### Spec.Logging

{{< bootstrap-table "table table-striped table-bordered" >}}
| name  | description                                                            | type   | required |
|-------|------------------------------------------------------------------------|--------|----------|
| level | Level defines the logging level. Supported values: info, debug, error. | string | no       |
{{< /bootstrap-table >}}

## Viewing and Updating the Configuration

{{< note >}} For the following examples, the name `nginx-gateway-config` should be updated to the name of the resource created for your installation. {{< /note >}}

To view the current configuration:

```shell
kubectl -n nginx-gateway get nginxgateways nginx-gateway-config -o yaml
```

To update the configuration:

```shell
kubectl -n nginx-gateway edit nginxgateways nginx-gateway-config
```

This will open the configuration in your default editor. You can then update and save the configuration, which is applied automatically to the control plane.

To view the status of the configuration:

```shell
kubectl -n nginx-gateway describe nginxgateways nginx-gateway-config
```
