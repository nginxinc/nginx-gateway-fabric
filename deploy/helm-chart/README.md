# NGINX Gateway Fabric Helm Chart

- [NGINX Gateway Fabric Helm Chart](#nginx-gateway-fabric-helm-chart)
  - [Introduction](#introduction)
  - [Prerequisites](#prerequisites)
    - [Installing the Gateway API resources](#installing-the-gateway-api-resources)
  - [Installing the Chart](#installing-the-chart)
    - [Installing the Chart from the OCI Registry](#installing-the-chart-from-the-oci-registry)
    - [Installing the Chart via Sources](#installing-the-chart-via-sources)
      - [Pulling the Chart](#pulling-the-chart)
      - [Installing the Chart](#installing-the-chart-1)
  - [Upgrading the Chart](#upgrading-the-chart)
    - [Upgrading the Gateway Resources](#upgrading-the-gateway-resources)
    - [Upgrading the CRDs](#upgrading-the-crds)
    - [Upgrading the Chart from the OCI Registry](#upgrading-the-chart-from-the-oci-registry)
    - [Upgrading the Chart from the Sources](#upgrading-the-chart-from-the-sources)
    - [Configure Delayed Termination for Zero Downtime Upgrades](#configure-delayed-termination-for-zero-downtime-upgrades)
  - [Uninstalling the Chart](#uninstalling-the-chart)
    - [Uninstalling the Gateway Resources](#uninstalling-the-gateway-resources)
  - [Configuration](#configuration)

## Introduction

This chart deploys the NGINX Gateway Fabric in your Kubernetes cluster.

## Prerequisites

- [Helm 3.0+](https://helm.sh/docs/intro/install/)
- [kubectl](https://kubernetes.io/docs/tasks/tools/)

### Installing the Gateway API resources

> Note: The Gateway API resources from the standard channel (the CRDs and the validating webhook) must be installed
> before deploying NGINX Gateway Fabric. If they are already installed in your cluster, please ensure they are
> the correct version as supported by the NGINX Gateway Fabric -
> [see the Technical Specifications](https://github.com/nginxinc/nginx-gateway-fabric/blob/main/README.md#technical-specifications).

To install the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

```shell
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
```

## Installing the Chart

### Installing the Chart from the OCI Registry

To install the chart with the release name `my-release` (`my-release` is the name that you choose) into the
nginx-gateway namespace (with optional `--create-namespace` flag - you can omit if the namespace already exists):

```shell
helm install my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric  --create-namespace --wait -n nginx-gateway
```

This will install the latest stable release. To install the latest version from the `main` branch, specify the
`--version 0.0.0-edge` flag when installing.

### Installing the Chart via Sources

#### Pulling the Chart

```shell
helm pull oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric --untar
cd nginx-gateway-fabric
```

This will pull the latest stable release. To pull the latest version from the `main` branch, specify the
`--version 0.0.0-edge` flag when pulling.

#### Installing the Chart

To install the chart with the release name `my-release` (`my-release` is the name that you choose) into the
nginx-gateway namespace (with optional `--create-namespace` flag - you can omit if the namespace already exists):

```shell
helm install my-release . --create-namespace --wait -n nginx-gateway
```

## Upgrading the Chart

> **Note**
> See [below](#configure-delayed-termination-for-zero-downtime-upgrades) for instructions on how to configure delayed
> termination if required for zero downtime upgrades in your environment.

### Upgrading the Gateway Resources

Before you upgrade a release, ensure the Gateway API resources are the correct version as supported by the NGINX
Gateway Fabric - [see the Technical Specifications](../../README.md#technical-specifications).:

To upgrade the Gateway resources from [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

```shell
kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
```

### Upgrading the CRDs

Helm does not upgrade the NGINX Gateway Fabric CRDs during a release upgrade. Before you upgrade a release, you
must [pull the chart](#pulling-the-chart) from GitHub and run the following command to upgrade the CRDs:

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
helm upgrade my-release oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric -n nginx-gateway
```

This will upgrade to the latest stable release. To upgrade to the latest version from the `main` branch, specify
the `--version 0.0.0-edge` flag when upgrading.

### Upgrading the Chart from the Sources

Pull the chart sources as described in [Pulling the Chart](#pulling-the-chart), if not already present. Then, to upgrade
the release `my-release`, run:

```shell
helm upgrade my-release . -n nginx-gateway
```

### Configure Delayed Termination for Zero Downtime Upgrades

To achieve zero downtime upgrades (meaning clients will not see any interruption in traffic while a rolling upgrade is
being performed on NGF), you may need to configure delayed termination on the NGF Pod, depending on your environment.

> **Note**
> When proxying Websocket or any long-lived connections, NGINX will not terminate until that connection is closed
> by either the client or the backend. This means that unless all those connections are closed by clients/backends
> before or during an upgrade, NGINX will not terminate, which means Kubernetes will kill NGINX. As a result, the
> clients will see the connections abruptly closed and thus experience downtime.

1. Add `lifecycle` to both the nginx and the nginx-gateway container definition. To do so, update your `values.yaml`
   file to include the following (update the `sleep` values to what is required in your environment):

   ```yaml
    nginxGateway:
    <...>
    lifecycle:
        preStop:
        exec:
            command:
            - /usr/bin/gateway
            - sleep
            - --duration=40s # This flag is optional, the default is 30s

    nginx:
    <...>
    lifecycle:
        preStop:
        exec:
            command:
            - /bin/sleep
            - "40"
   ```

2. Ensure the `terminationGracePeriodSeconds` matches or exceeds the `sleep` value from the `preStopHook` (the default
   is 30). This is to ensure Kubernetes does not terminate the Pod before the `preStopHook` is complete. To do so,
   update your `values.yaml` file to include the following (update the value to what is required in your environment):

   ```yaml
   terminationGracePeriodSeconds: 50
   ```

> **Note**
> More information on container lifecycle hooks can be found
> [here](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks) and a detailed
> description of Pod termination behavior can be found in
> [Termination of Pods](https://kubernetes.io/docs/concepts/workloads/Pods/Pod-lifecycle/#Pod-termination).

## Uninstalling the Chart

To uninstall/delete the release `my-release`:

```shell
helm uninstall my-release -n nginx-gateway
kubectl delete ns nginx-gateway
kubectl delete crd nginxgateways.gateway.nginx.org nginxproxies.gateway.nginx.org
```

These commands remove all the Kubernetes components associated with the release and deletes the release.

### Uninstalling the Gateway Resources

> **Warning: This command will delete all the corresponding custom resources in your cluster across all namespaces!
Please ensure there are no custom resources that you want to keep and there are no other Gateway API implementations
running in the cluster!**

To delete the Gateway resources using [the Gateway API repo](https://github.com/kubernetes-sigs/gateway-api), run:

```shell
kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
```

## Configuration

The following tables lists the configurable parameters of the NGINX Gateway Fabric chart and their default values.

| Parameter                                         | Description                                                                                                                                                                                                  | Default Value                                                                                                   |
|---------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-----------------------------------------------------------------------------------------------------------------|
| `nginxGateway.image.repository`                   | The repository for the NGINX Gateway Fabric image.                                                                                                                                                       | ghcr.io/nginxinc/nginx-gateway-fabric                                                                       |
| `nginxGateway.image.tag`                          | The tag for the NGINX Gateway Fabric image.                                                                                                                                                              | edge                                                                                                            |
| `nginxGateway.image.pullPolicy`                   | The `imagePullPolicy` for the NGINX Gateway Fabric image.                                                                                                                                                | Always                                                                                                          |
| `nginxGateway.lifecycle`                          | The `lifecycle` of the nginx-gateway container.                                                                                                                                              | {}                                                                                                              |
| `nginxGateway.extraVolumeMounts`                  | Extra `volumeMounts` for the nginxGateway container.                                                                                                                                                         | {}                                                                                                              |
| `nginxGateway.gatewayClassName`                   | The name of the GatewayClass for the NGINX Gateway Fabric deployment.                                                                                                                                    | nginx                                                                                                           |
| `nginxGateway.gatewayControllerName`              | The name of the Gateway controller. The controller name must be of the form: DOMAIN/PATH. The controller's domain is gateway.nginx.org.                                                                      | gateway.nginx.org/nginx-gateway-controller                                                                      |
| `nginxGateway.kind`                               | The kind of the NGINX Gateway Fabric installation - currently, only Deployment is supported.                                                                                                             | deployment                                                                                                      |
| `nginxGateway.config`                             | The dynamic configuration for the control plane that is contained in the NginxGateway resource                                                                                                               | [See nginxGateway.config section](values.yaml)                                                                  |
| `nginxGateway.readinessProbe.enable`              | Enable the /readyz endpoint on the control plane.                                                                                                                                                            | true                                                                                                            |
| `nginxGateway.readinessProbe.port`                | Port in which the readiness endpoint is exposed.                                                                                                                                                             | 8081                                                                                                            |
| `nginxGateway.readinessProbe.initialDelaySeconds` | The number of seconds after the Pod has started before the readiness probes are initiated.                                                                                                                   | 3                                                                                                               |
| `nginxGateway.replicaCount`                       | The number of replicas of the NGINX Gateway Fabric Deployment.                                                                                                                                           | 1                                                                                                               |
| `nginxGateway.leaderElection.enable`              | Enable leader election. Leader election is used to avoid multiple replicas of the NGINX Gateway Fabric reporting the status of the Gateway API resources.                                                | true                                                                                                            |
| `nginxGateway.leaderElection.lockName`            | The name of the leader election lock. A Lease object with this name will be created in the same Namespace as the controller.                                                                                 | Autogenerated                                                                                                   |
| `nginx.image.repository`                          | The repository for the NGINX image.                                                                                                                                                                          | ghcr.io/nginxinc/nginx-gateway-fabric/nginx                                                                 |
| `nginx.image.tag`                                 | The tag for the NGINX image.                                                                                                                                                                                 | edge                                                                                                            |
| `nginx.image.pullPolicy`                          | The `imagePullPolicy` for the NGINX image.                                                                                                                                                                   | Always                                                                                                          |
| `nginx.config`                                    | The configuration for the data plane that is contained in the NginxProxy resource. | [See nginx.config section](values.yaml) |
| `nginx.lifecycle`                                 | The `lifecycle` of the nginx container.                                                                                                                                                            | {}                                                                                                              |
| `nginx.extraVolumeMounts`                         | Extra `volumeMounts` for the nginx container.                                                                                                                                                            | {}                                                                                                              |
| `terminationGracePeriodSeconds`                   | The termination grace period of the NGINX Gateway Fabric pod.                                                                                                                                                            | 30                                                                                                              |
| `tolerations`                                     | The `tolerations` of the NGINX Gateway Fabric pod.                                                                                                                                                           | []                                                                                                              |
| `affinity`                                        | The `affinity` of the NGINX Gateway Fabric pod.                                                                                                                                                            | {}                                                                                                              |
| `serviceAccount.annotations`                      | The `annotations` for the ServiceAccount used by the NGINX Gateway Fabric deployment.                                                                                                                    | {}                                                                                                              |
| `serviceAccount.name`                             | Name of the ServiceAccount used by the NGINX Gateway Fabric deployment.                                                                                                                                  | Autogenerated                                                                                                   |
| `service.create`                                  | Creates a service to expose the NGINX Gateway Fabric pods.                                                                                                                                               | true                                                                                                            |
| `service.type`                                    | The type of service to create for the NGINX Gateway Fabric.                                                                                                                                              | Loadbalancer                                                                                                    |
| `service.externalTrafficPolicy`                   | The `externalTrafficPolicy` of the service. The value `Local` preserves the client source IP.                                                                                                                | Local                                                                                                           |
| `service.annotations`                             | The `annotations` of the NGINX Gateway Fabric service.                                                                                                                                                   | {}                                                                                                              |
| `service.ports`                                   | A list of ports to expose through the NGINX Gateway Fabric service. Update it to match the listener ports from your Gateway resource. Follows the conventional Kubernetes yaml syntax for service ports. | [ port: 80, targetPort: 80, protocol: TCP, name: http; port: 443, targetPort: 443, protocol: TCP, name: https ] |
| `metrics.disable`                                 | Disable exposing metrics in the Prometheus format.                                                                                                                                                           | false                                                                                                           |
| `metrics.port`                                    | Set the port where the Prometheus metrics are exposed. Format: [1024 - 65535]                                                                                                                                | 9113                                                                                                            |
| `metrics.secure`                                  | Enable serving metrics via https. By default metrics are served via http. Please note that this endpoint will be secured with a self-signed certificate.                                                     | false                                                                                                           |
| `extraVolumes`                                    | Extra `volumes` for the NGINX Gateway Fabric pod.                                                                                                                                                            | []                                                                                                              |
