# Monitoring

The NGINX Kubernetes Gateway exposes a number of metrics in the [Prometheus](https://prometheus.io/) format. Those
include NGINX and the controller-runtime metrics. These are delivered using a metrics server orchestrated by the
controller-runtime package.

## Enabling Metrics

### Using Helm

If you're using *Helm* to install the NGINX Kubernetes Gateway, to enable Prometheus metrics, configure the `metrics.*`
parameters of the Helm chart. See the [Helm README](/deploy/helm-chart/README.md). Please note that if `metrics.secure`
is set to true, as the metrics server is using a self-signed certificate, the Prometheus pod scrape configuration will
also require the `insecure_skip_verify` flag set. See [the Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#tls_config).

### Using Manifests

If you're using *Kubernetes manifests* to install the NGINX Kubernetes Gateway, modify the [manifest](/deploy/manifests/nginx-gateway.yaml)
to enable Prometheus metrics to enable the following:

1. Run the NGINX Kubernetes Gateway with the `-enable-metrics` [command-line argument](/docs/cli-help.md).
   As a result, the NGINX Kubernetes Gateway will expose NGINX metrics in the Prometheus format via the path `/metrics`
   on port `9113` (customizable via the `-metrics-listen-port` [command-line argument](/docs/cli-help.md)) using http
   (customizable via the `-metrics-secure-serving` [command-line argument](/docs/cli-help.md)).
2. Add the Prometheus port to the list of the ports of the NGINX Kubernetes Gateway container in the template of the
   NGINX Kubernetes Gateway pod:

    ```yaml
    - name: prometheus
      containerPort: 9113
    ```

3. Make Prometheus aware of the NGINX Kubernetes Gateway targets by adding the following annotations to the template of
   the NGINX Kubernetes Gateway pod (note: this assumes your Prometheus is configured to discover targets by analyzing the
   annotations of pods):

    ```yaml
    annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "9113"
    ```

  If the `-metrics-secure-serving` flag is enabled, the following annotation will also needed to be added to the pod spec:

  ```yaml
  prometheus.io/scheme: "https"
  ```

  As the metrics server is using a self-signed certificate, the Prometheus pod scrape configuration will also require the
  `insecure_skip_verify` flag set. See [the Prometheus documentation](https://prometheus.io/docs/prometheus/latest/configuration/configuration/#tls_config)

## Available Metrics

The NGINX Kubernetes Gateway exports the following metrics:

- NGINX metrics:
  - Exported by NGINX. Refer to the [NGINX Prometheus Exporter developer docs](https://github.com/nginxinc/nginx-prometheus-exporter#metrics-for-nginx-oss)

- [controller-runtime](https://github.com/kubernetes-sigs/controller-runtime) metrics. These include:
  - Total number of reconcilation errors per controller
  - Length of reconcile queue per controller
  - Reconcilation latency
  - Usual resource metrics such as CPU, memory usage, file descriptor usage
  - Go runtime metrics such as number of Go routines, GC duration, and Go version information
