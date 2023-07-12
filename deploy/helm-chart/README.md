# NGINX Kubernetes Gateway Helm Chart

## Introduction

This chart deploys the NGINX Kubernetes Gateway in your Kubernetes cluster.

## Prerequisites

- [Helm 3.0+](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

> Note: NGINX Kubernetes Gateway can only run in the `nginx-gateway` namespace. This limitation will be addressed in
the future releases.

## Installing the Chart

> Note: The Gateway API resources from the standard channel (the CRDs and the validating webhook) must be installed
before deploying NGINX Kubernetes Gateway. By default, they will be installed by the Chart if not already
present in the cluster. If they are already installed in your cluster, please ensure they are the correct version as
supported by the NGINX Kubernetes Gateway - [see the Technical Specifications](../../README.md#technical-specifications).
Helm will not upgrade CRDs - to do so manually, see
[Upgrading the Gateway API resources](#upgrading-the-gateway-resources).

### Installing the Chart from the OCI registry

To install the chart with the release name `my-release` (`my-release` is the name that you choose) into the
nginx-gateway namespace (with optional `--create-namespace` flag - you can omit if the namespace already exists), and
the Gateway API resources from the standard channel (if not already present):

```
helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway --version 0.0.0-edge --create-namespace --wait-for-jobs -n nginx-gateway
```

### Installing the Chart via Sources

#### Pulling the Chart

```
helm pull oci://ghcr.io/nginxinc/charts/nginx-gateway --untar --version 0.0.0-edge
cd nginx-gateway
```

#### Installing the Chart

To install the chart with the release name `my-release` (`my-release` is the name that you choose) into the
nginx-gateway namespace (with optional `--create-namespace` flag - you can omit if the namespace already exists), and
the Gateway API resources from the standard channel (if not already present):

```
helm install my-release . --create-namespace --wait-for-jobs -n nginx-gateway
```

## Upgrading the Chart
### Upgrading the Gateway resources
Helm does not upgrade CRDs during a release upgrade, or on an install if they are already present in the cluster.
Before you upgrade a release, ensure the Gateway API resources are up to date by doing one of the following:

1. To upgrade the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:
   ```
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
   ```
1. To upgrade the Gateway resources from the NGINX Kubernetes Gateway Chart sources, pull the chart sources as described
   in [Pulling the Chart](#pulling-the-chart) and then run:
   ```
   kubectl apply -f crds/
   ```

>Note: The following warning is expected and can be ignored: `Warning: kubectl apply should be used on resource created
by either kubectl create --save-config or kubectl apply`.

### Upgrading the Chart from the OCI registry
To upgrade the release `my-release`, run:

```
helm upgrade my-release oci://ghcr.io/nginxinc/charts/nginx-ingress --version 0.0.0-edge -n nginx-gateway
```

### Upgrading the Chart from the sources

Pull the chart sources as described in [Pulling the Chart](#pulling-the-chart), if not already present. Then, to upgrade
the release `my-release`, run:
```
helm upgrade my-release . -n nginx-gateway
```

## Uninstalling the Chart

To uninstall/delete the release `my-release`:

```
helm uninstall my-release -n nginx-gateway
```

The command removes all the Kubernetes components associated with the release and deletes the release.

### Uninstalling the Gateway resources
Uninstalling the release does NOT uninstall the CRDs or validating webhook. To clean up these resources, do one of the
following:

>**Warning: These commands will delete all the corresponding custom resources in your cluster across all namespaces!
Please ensure there are no custom resources that you want to keep and there are no other Gateway API implementations
running in the cluster!**

1. To delete the Gateway resources using [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:
   ```
   kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.7.1/standard-install.yaml
   ```
1. To delete the Gateway resources using the NGINX Kubernetes Gateway Chart sources, pull the chart sources as described
   in [Pulling the Chart](#pulling-the-chart) and then run:
   ```
   kubectl delete -f crds/
   ```

## Configuration

The following tables lists the configurable parameters of the NGINX Kubernetes Gateway chart and their default values.

|Parameter | Description | Default |
| --- | --- | --- |
|`nginxGateway.image.repository` | The repository for the NGINX Kubernetes Gateway image. | ghcr.io/nginxinc/nginx-kubernetes-gateway |
|`nginxGateway.image.tag` | The tag for the NGINX Kubernetes Gateway image. | edge |
|`nginxGateway.imagePullPolicy` | The imagePullPolicy for the NGINX Kubernetes Gateway image. | Always |
|`nginxGateway.gatewayClass` | The GatewayClass for the NGINX Kubernetes Gateway deployment. | nginx |
|`nginx.image.repository` | The repository for the NGINX image. | nginx |
|`nginx.image.tag` | The tag for the NGINX image. | 1.25 |
|`nginx.imagePullPolicy` | The imagePullPolicy for the NGINX image. | Always |
|`initContainer.image.repository` | The repository for the initContainer image. | busybox |
|`initContainer.image.tag` | The tag for the initContainer image. | 1.36 |
|`serviceAccount.annotations` | Annotations for the ServiceAccount used by the NGINX Kubernetes Gateway deployment. | {} |
|`serviceAccount.name` | Name of the ServiceAccount used by the NGINX Kubernetes Gateway deployment. | Autogenerated |
|`service.create` | Creates a service to expose the NGINX Kubernetes Gateway pods. | true |
|`service.type` | The type of service to create for the NGINX Kubernetes Gateway. | Loadbalancer |
|`service.externalTrafficPolicy` | The externalTrafficPolicy of the service. The value Local preserves the client source IP. | Local |
|`service.annotations` | The annotations of the NGINX Kubernetes Gateway service. | true |
|`service.ports` | A list of ports to expose through the NGINX Kubernetes Gateway service. Follows the conventional Kubernetes yaml syntax for service ports. | [ port: 80, targetPort: 80, protocol: TCP, name: http; port: 443, targetPort: 443, protocol: TCP, name: https ] |
