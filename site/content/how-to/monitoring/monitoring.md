---
title: "Monitoring NGINX Gateway Fabric"
description: "Learn how to monitor your NGINX Gateway Fabric effectively. This guide provides easy steps for configuring monitoring settings and understanding key performance metrics."
weight: 100
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Overview


NGINX Gateway Fabric metrics are displayed in [Prometheus](https://prometheus.io/) format, simplifying monitoring. You can track NGINX and controller-runtime metrics through a metrics server orchestrated by the controller-runtime package. These metrics are enabled by default and can be accessed on HTTP port `9113`.


{{<call-out "important" "Security note for metrics">}}
Metrics are served over HTTP by default. Enabling HTTPS will secure the metrics endpoint with a self-signed certificate. When using HTTPS, adjust the Prometheus Pod scrape settings by adding the `insecure_skip_verify` flag to handle the self-signed certificate. For further details, refer to the [Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#tls_config).
{{</call-out>}}

## How to change the default metrics configuration

Configuring NGINX Gateway Fabric for monitoring is straightforward. You can change metric settings using Helm or Kubernetes manifests, depending on your setup.

### Using Helm

If you're setting up NGINX Gateway Fabric with Helm, you can adjust the `metrics.*` parameters to fit your needs. For detailed options and instructions, see the [Helm README](/deploy/helm-chart/README.md).

### Using Kubernetes manifests

For setups using Kubernetes manifests, change the metrics configuration by editing the [NGINX Gateway manifest](/deploy/manifests/nginx-gateway.yaml).

#### Disabling metrics

If you need to disable metrics:

1. Set the `-metrics-disable` [command-line argument]({{< relref "reference/cli-help.md">}}) to `true` in the NGINX Gateway Fabric Pod's configuration. Remove any other `-metrics-*` arguments.
2. In the Pod template for NGINX Gateway Fabric, delete the metrics port entry from the container ports list:

    ```yaml
    - name: metrics
      containerPort: 9113
    ```

3. Also, remove the following annotations from the NGINX Gateway Fabric Pod template:

    ```yaml
    annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9113"
    ```

#### Changing the default port

To change the default port for metrics:

1. Update the `-metrics-port` [command-line argument]({{< relref "reference/cli-help.md">}}) in the NGINX Gateway Fabric Pod's configuration to your chosen port number.
2. In the Pod template, change the metrics port entry to reflect the new port:

    ```yaml
    - name: metrics
      containerPort: <new-port>
    ```

3. Modify the `prometheus.io/port` annotation in the Pod template to match the new port:

    ```yaml
    annotations:
        <...>
        prometheus.io/port: "<new-port>"
        <...>
    ```

#### Enabling HTTPS for metrics

For enhanced security with HTTPS:

1. Enable HTTPS security by setting the `-metrics-secure-serving` [command-line argument]({{< relref "reference/cli-help.md">}}) to `true` in the NGINX Gateway Fabric Pod's configuration.

2. Add an HTTPS scheme annotation to the Pod template:

    ```yaml
    annotations:
        <...>
        prometheus.io/scheme: "https"
        <...>
    ```

## Available metrics in NGINX Gateway Fabric

NGINX Gateway Fabric provides a variety of metrics to assist in monitoring and analyzing performance. These metrics are categorized as follows:

### NGINX/ NGINX Plus metrics

NGINX metrics, essential for monitoring specific NGINX operations, include details like the total number of accepted client connections. For a complete list of available NGINX/ NGINX Plus metrics, refer to the [NGINX Prometheus Exporter developer docs](https://github.com/nginxinc/nginx-prometheus-exporter#exported-metrics).

These metrics use  the `nginx_gateway_fabric` namespace and include the `class` label, indicating the NGINX Gateway class. For example, `nginx_gateway_fabric_connections_accepted{class="nginx"}`.

### NGINX Gateway Fabric metrics

Metrics specific to the NGINX Gateway Fabric include:

- `nginx_reloads_total`: Counts successful NGINX reloads.
- `nginx_reload_errors_total`: Counts NGINX reload failures.
- `nginx_stale_config`: Indicates if NGINX Gateway Fabric couldn't update NGINX with the latest configuration, resulting in a stale version.
- `nginx_last_reload_milliseconds`: Time in milliseconds for NGINX reloads.
- `event_batch_processing_milliseconds`: Time in milliseconds to process batches of Kubernetes events.

All these metrics are under the `nginx_gateway_fabric` namespace and include a `class` label set to the Gateway class of NGINX Gateway Fabric. For example, `nginx_gateway_fabric_nginx_reloads_total{class="nginx"}`.

### Controller-runtime metrics

Provided by the [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) library, these metrics cover a range of aspects:

- General resource usage like CPU and memory.
- Go runtime metrics such as the number of Go routines, garbage collection duration, and Go version.
- Controller-specific metrics, including reconciliation errors per controller, length of the reconcile queue, and reconciliation latency.
