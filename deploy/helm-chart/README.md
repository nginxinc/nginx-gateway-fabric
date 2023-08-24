# NGINX Kubernetes Gateway Helm Chart

## Introduction

This chart deploys the NGINX Kubernetes Gateway in your Kubernetes cluster.

## Prerequisites

- [Helm 3.0+](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Installing the Gateway API resources

> Note: The Gateway API resources from the standard channel (the CRDs and the validating webhook) must be installed
before deploying NGINX Kubernetes Gateway. If they are already installed in your cluster, please ensure they are the
correct version as supported by the NGINX Kubernetes Gateway -
[see the Technical Specifications](../../README.md#technical-specifications).

To install the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

```shell
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.0/standard-install.yaml
```

## Installing the Chart

### Installing the Chart from the OCI Registry

To install the chart with the release name `my-release` (`my-release` is the name that you choose) into the
nginx-gateway namespace (with optional `--create-namespace` flag - you can omit if the namespace already exists):

```shell
helm install my-release oci://ghcr.io/nginxinc/charts/nginx-kubernetes-gateway --version 0.0.0-edge --create-namespace --wait -n nginx-gateway
```

### Installing the Chart via Sources

#### Pulling the Chart

```shell
helm pull oci://ghcr.io/nginxinc/charts/nginx-kubernetes-gateway --untar --version 0.0.0-edge
cd nginx-kubernetes-gateway
```

#### Installing the Chart

To install the chart with the release name `my-release` (`my-release` is the name that you choose) into the
nginx-gateway namespace (with optional `--create-namespace` flag - you can omit if the namespace already exists):

```shell
helm install my-release . --create-namespace --wait -n nginx-gateway
```

## Upgrading the Chart
### Upgrading the Gateway Resources

Before you upgrade a release, ensure the Gateway API resources are the correct version as supported by the NGINX
Kubernetes Gateway - [see the Technical Specifications](../../README.md#technical-specifications).:

To upgrade the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

```shell
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.0/standard-install.yaml
```

### Upgrading the CRDs

Helm does not upgrade the NGINX Kubernetes Gateway CRDs during a release upgrade. Before you upgrade a release,
you must [pull the chart](#pulling-the-chart) from GitHub and run the following command to upgrade the CRDs:

```shell
kubectl apply -f crds/
```

The following warning is expected and can be ignored:

```text
Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply.
```

### Upgrading the Chart from the OCI Registry

To upgrade the release `my-release`, run:

```shell
helm upgrade my-release oci://ghcr.io/nginxinc/charts/nginx-kubernetes-gateway --version 0.0.0-edge -n nginx-gateway
```

### Upgrading the Chart from the Sources

Pull the chart sources as described in [Pulling the Chart](#pulling-the-chart), if not already present. Then, to upgrade
the release `my-release`, run:

```shell
helm upgrade my-release . -n nginx-gateway
```

## Uninstalling the Chart

To uninstall/delete the release `my-release`:

```shell
helm uninstall my-release -n nginx-gateway
kubectl delete ns nginx-gateway
kubectl delete crd nginxgateways.gateway.nginx.org
```

These commands remove all the Kubernetes components associated with the release and deletes the release.

### Uninstalling the Gateway Resources

>**Warning: This command will delete all the corresponding custom resources in your cluster across all namespaces!
Please ensure there are no custom resources that you want to keep and there are no other Gateway API implementations
running in the cluster!**

To delete the Gateway resources using [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

```shell
kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.0/standard-install.yaml
```

## Configuration

The following tables lists the configurable parameters of the NGINX Kubernetes Gateway chart and their default values.

| Parameter                            | Description                                                                                                                                                                                                  | Default Value                                                                                                   |
|--------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `nginxGateway.image.repository`      | The repository for the NGINX Kubernetes Gateway image.                                                                                                                                                       | ghcr.io/nginxinc/nginx-kubernetes-gateway                                                                       |
| `nginxGateway.image.tag`             | The tag for the NGINX Kubernetes Gateway image.                                                                                                                                                              | edge                                                                                                            |
| `nginxGateway.image.pullPolicy`      | The `imagePullPolicy` for the NGINX Kubernetes Gateway image.                                                                                                                                                | Always                                                                                                          |
| `nginxGateway.gatewayClassName`      | The name of the GatewayClass for the NGINX Kubernetes Gateway deployment.                                                                                                                                    | nginx                                                                                                           |
| `nginxGateway.gatewayControllerName` | The name of the Gateway controller. The controller name must be of the form: DOMAIN/PATH. The controller's domain is gateway.nginx.org.                                                                  | gateway.nginx.org/nginx-gateway-controller                                                                  |
| `nginxGateway.kind`                  | The kind of the NGINX Kubernetes Gateway installation - currently, only Deployment is supported.                                                                                                             | deployment                                                                                                      |
|`nginxGateway.config`                 | The dynamic configuration for the control plane that is contained in the NginxGateway resource                                                                                                               | [See nginxGateway.config section](values.yaml) |
| `nginx.image.repository`             | The repository for the NGINX image.                                                                                                                                                                          | ghcr.io/nginxinc/nginx-kubernetes-gateway/nginx                                                                 |
| `nginx.image.tag`                    | The tag for the NGINX image.                                                                                                                                                                                 | edge                                                                                                            |
| `nginx.image.pullPolicy`             | The `imagePullPolicy` for the NGINX image.                                                                                                                                                                   | Always                                                                                                          |
| `serviceAccount.annotations`         | The `annotations` for the ServiceAccount used by the NGINX Kubernetes Gateway deployment.                                                                                                                    | {}                                                                                                              |
| `serviceAccount.name`                | Name of the ServiceAccount used by the NGINX Kubernetes Gateway deployment.                                                                                                                                  | Autogenerated                                                                                                   |
| `service.create`                     | Creates a service to expose the NGINX Kubernetes Gateway pods.                                                                                                                                               | true                                                                                                            |
| `service.type`                       | The type of service to create for the NGINX Kubernetes Gateway.                                                                                                                                              | Loadbalancer                                                                                                    |
| `service.externalTrafficPolicy`      | The `externalTrafficPolicy` of the service. The value `Local` preserves the client source IP.                                                                                                                | Local                                                                                                           |
| `service.annotations`                | The `annotations` of the NGINX Kubernetes Gateway service.                                                                                                                                                   | {}                                                                                                              |
| `service.ports`                      | A list of ports to expose through the NGINX Kubernetes Gateway service. Update it to match the listener ports from your Gateway resource. Follows the conventional Kubernetes yaml syntax for service ports. | [ port: 80, targetPort: 80, protocol: TCP, name: http; port: 443, targetPort: 443, protocol: TCP, name: https ] |
| `metrics.enable`                     | Enable exposing metrics in the Prometheus format.                                                                                                                                                   |false                                                                                                              |
| `metrics.port`                       | Sets the port where the Prometheus metrics are exposed. Format: [1024 - 65535]                                                                                                                                  |9113                                                                                                              |
| `metrics.secure`                     | Configures if the metrics endpoint should be secured using https. Please note that this endpoint will be secured with a self-signed certificate.                                                                                                              |false                                                                                                              |
