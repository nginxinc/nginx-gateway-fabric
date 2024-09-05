---
title: "Command-line reference guide"
weight: 100
toc: true
docs: "DOCS-1435"
---

## Overview

Learn about the commands available for the executable file of the NGINX Gateway Fabric container.

## Static mode

This command configures NGINX for a single NGINX Gateway Fabric resource.

*Usage*:

```shell
  gateway static-mode [flags]
```

### Flags

{{< bootstrap-table "table table-bordered table-striped table-responsive" >}}
| Name                                | Type     | Description                                                                                                                                                                                                                                                                                                                                                                              |
|-------------------------------------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| _gateway-ctlr-name_                 | _string_ | The name of the Gateway controller. The controller name must be in the form: `DOMAIN/PATH`. The controller's domain is `gateway.nginx.org`.                                                                                                                                                                                                                                              |
| _gatewayclass_                      | _string_ | The name of the GatewayClass resource. Every NGINX Gateway Fabric must have a unique corresponding GatewayClass resource.                                                                                                                                                                                                                                                                |
| _gateway_                           | _string_ | The namespaced name of the Gateway resource to use. Must be of the form: `NAMESPACE/NAME`. If not specified, the control plane will process all Gateways for the configured GatewayClass. Among them, it will choose the oldest resource by creation timestamp. If the timestamps are equal, it will choose the resource that appears first in alphabetical order by {namespace}/{name}. |
| _nginx-plus_                        | _bool_   | Enable support for NGINX Plus.                                                                                                                                                                                                                                                                                                                                                           |
| _gateway-api-experimental-features_ | _bool_   | Enable the experimental features of Gateway API which are supported by NGINX Gateway Fabric. Requires the Gateway APIs installed from the experimental channel.                                                                                                                                                                                                                          |
| _config_                            | _string_ | The name of the NginxGateway resource to be used for this controller's dynamic configuration. Lives in the same namespace as the controller.                                                                                                                                                                                                                                             |
| _service_                           | _string_ | The name of the service that fronts this NGINX Gateway Fabric pod. Lives in the same namespace as the controller.                                                                                                                                                                                                                                                                        |
| _metrics-disable_                   | _bool_   | Disable exposing metrics in the Prometheus format (Default: `false`).                                                                                                                                                                                                                                                                                                                    |
| _metrics-listen-port_               | _int_    | Sets the port where the Prometheus metrics are exposed. An integer between 1024 - 65535 (Default: `9113`)                                                                                                                                                                                                                                                                                |
| _metrics-secure-serving_            | _bool_   | Configures if the metrics endpoint should be secured using https. Note that this endpoint will be secured with a self-signed certificate (Default `false`).                                                                                                                                                                                                                              |
| _update-gatewayclass-status_        | _bool_   | Update the status of the GatewayClass resource (Default: `true`).                                                                                                                                                                                                                                                                                                                        |
| _health-disable_                    | _bool_   | Disable running the health probe server (Default: `false`).                                                                                                                                                                                                                                                                                                                              |
| _health-port_                       | _int_    | Set the port where the health probe server is exposed. An integer between 1024 - 65535 (Default: `8081`).                                                                                                                                                                                                                                                                                |
| _leader-election-disable_           | _bool_   | Disable leader election, which is used to avoid multiple replicas of the NGINX Gateway Fabric reporting the status of the Gateway API resources. If disabled, all replicas of NGINX Gateway Fabric will update the statuses of the Gateway API resources (Default: `false`).                                                                                                             |
| _leader-election-lock-name_         | _string_ | The name of the leader election lock. A lease object with this name will be created in the same namespace as the controller (Default: `"nginx-gateway-leader-election-lock"`).                                                                                                                                                                                                           |
| _product-telemetry-disable_         | _bool_   | Disable the collection of product telemetry (Default: `false`).                                                                                                                                                                                                                                                                                                                          |
| _usage-report-secret_               | _string_ | The namespace/name of the Secret containing the credentials for NGINX Plus usage reporting.                                                                                                                                                                                                                                                                                              |
| _usage-report-server-url_           | _string_ | The base server URL of the NGINX Plus usage reporting server.                                                                                                                                                                                                                                                                                                                            |
| _usage-report-cluster-name_         | _string_ | The display name of the Kubernetes cluster in the NGINX Plus usage reporting server.                                                                                                                                                                                                                                                                                                     |
| _usage-report-skip-verify_          | _bool_   | Disable client verification of the NGINX Plus usage reporting server certificate.                                                                                                                                                                                                                                                                                                        |
| _snippets-filters_           | _bool_   | Enable SnippetsFilters feature. SnippetsFilters allow inserting NGINX configuration into the generated NGINX config for HTTPRoute and GRPCRoute resources.                                                                                                                                                                                                                               |
{{% /bootstrap-table %}}

## Sleep

This command sleeps for specified duration, then exits.

_Usage_:

```shell
  gateway sleep [flags]
```

{{< bootstrap-table "table table-bordered table-striped table-responsive" >}}
| Name     | Type            | Description                                                                                                                   |
| -------- | --------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| duration | `time.Duration` | Set the duration of sleep. Must be parsable by [`time.ParseDuration`](https://pkg.go.dev/time#ParseDuration). (default `30s`) |
{{% /bootstrap-table %}}
