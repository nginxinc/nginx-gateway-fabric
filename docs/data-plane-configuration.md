# Data Plane Configuration

This document describes how to enable or customize various NGINX features.

## Overview

NGINX Gateway Fabric supports enabling or customizing various NGINX features via NginxProxy resource.
The resource is created during the installation of NGINX Gateway Fabric. An update to the resource will make
NGINX Gateway Fabric regenerate NGINX configuration and reload NGINX.

If using manifests, the default name of the resource is `nginx-proxy-config`. If using Helm, the default name
of the resource is `<release-name>-proxy-config`. It is deployed in the same Namespace as the controller
(default `nginx-gateway`).

The control plane only watches this single instance of the custom resource. If the resource is invalid per the OpenAPI
schema, the Kubernetes API server will reject the changes. If the resource is deleted or deemed invalid by NGINX
Gateway Fabric, a warning Event is created in the `nginx-gateway` Namespace, and the default values will be used by
the data plane for its configuration. Additionally, the control plane updates the status of the resource (if it exists)
to reflect whether it is valid or not.

### Spec

| name | description                                        | type              | required |
| ---- | -------------------------------------------------- | ----------------- | -------- |
| http | HTTP defines the NGINX HTTP context configuration. | [http](#spechttp) | no       |

### Spec.HTTP

| name      | description                                    | type                        | required |
| --------- | ---------------------------------------------- | --------------------------- | -------- |
| telemetry | Telemetry defines the telemetry configuration. | [telemetry](#spechttptelemetry) | no       |

### Spec.HTTP.Telemetry

| name    | description                                | type                    | required |
| ------- | ------------------------------------------ | ----------------------- | -------- |
| tracing | Tracing defines the tracing configuration. | [tracing](#spechttptelemetrytracing) | no       |

### Spec.HTTP.Telemetry.Tracing

| name       | description                                                                                                         | type   | required |
| ---------- | ------------------------------------------------------------------------------------------------------------------- | ------ | -------- |
| enable    | Enabled enables or disables OpenTelemetry tracing at the HTTP context. Default is false.                            | bool   | no       |
| endpoint   | Endpoint specifies the address of OTLP/gRPC endpoint that will accept telemetry data.                               | bool   | yes      |
| interval   | Interval specifies the tracing interval. Default is 5s.                                                             | string | no       |
| batchSize  | BatchSize specifies the maximum number of spans to be sent in one batch per worker. Default is 512.                 | int    | no       |
| batchCount | BatchCount specifies the number of pending batches per worker, spans exceeding the limit are dropped. Default is 4. | int    | no       |

## Viewing and Updating the Configuration

> For the following examples, the name `nginx-proxy-config` should be updated to the name of the resource that
> was created by your installation.

To view the current configuration:

```shell
kubectl -n nginx-gateway get nginxproxies nginx-proxy-config -o yaml
```

To update the configuration:

```shell
kubectl -n nginx-gateway edit nginxproxies nginx-proxy-config
```

This will open the configuration in your default editor. You can then update and save the configuration, which is
applied automatically to the data plane.

To view the status of the configuration:

```shell
kubectl -n nginx-gateway describe nginxproxies nginx-proxy-config
```
