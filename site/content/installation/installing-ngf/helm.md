---
title: "Installation with Helm"
description: "Learn how to install, upgrade, and uninstall NGINX Gateway Fabric in a Kubernetes cluster with Helm."
weight: 100
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Prerequisites

To complete this guide, you'll need to install:

- [kubectl](https://kubernetes.io/docs/tasks/tools/), a command-line tool for managing Kubernetes clusters.
- [Helm 3.0 or later](https://helm.sh/docs/intro/install/), for deploying and managing applications on Kubernetes.


## Deploy NGINX Gateway Fabric

### Install from the OCI registry

- To install the latest stable release of NGINX Gateway Fabric in the **nginx-gateway** namespace, run the following command:

   ```shell
   helm install <my-release> oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric --create-namespace --wait -n nginx-gateway
   ```

   Change `<my-release>` to the name you want for your release. If the namespace already exists, you can omit the optional `--create-namespace` flag. If you want the latest version from the **main** branch, add `--version 0.0.0-edge` to your install command.

### Install from sources {#install-from-sources}

1. {{<include "installation/helm/pulling-the-chart.md" >}}

2. To install the chart into the **nginx-gateway** namespace, run the following command.

   ```shell
   helm install <my-release> . --create-namespace --wait -n nginx-gateway
   ```

   Change `<my-release>` to the name you want for your release. If the namespace already exists, you can omit the optional `--create-namespace` flag.

## Upgrade NGINX Gateway Fabric

{{<tip>}}For guidance on zero downtime upgrades, see the [Delay Pod Termination](#configure-delayed-pod-termination-for-zero-downtime-upgrades) section below.{{</tip>}}

To upgrade NGINX Gateway Fabric and get the latest features and improvements, take the following steps:

### Upgrade Gateway resources

To upgrade your Gateway API resources, take the following steps:

- Verify the Gateway API resources are compatible with your NGINX Gateway Fabric version. Refer to the [Technical Specifications]({{< relref "reference/technical-specifications.md" >}}) for details.
- Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v1.0.0) for any important upgrade-specific information.
- To upgrade the Gateway API resources, run:

   ```shell
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v1.0.0/standard-install.yaml
   ```

### Upgrade NGINX Gateway Fabric CRDs

Helm's upgrade process does not automatically upgrade the NGINX Gateway Fabric CRDs (Custom Resource Definitions).

To upgrade the CRDs, take the following steps:

1. {{<include "installation/helm/pulling-the-chart.md" >}}

2. Upgrade the CRDs:

      ```shell
      kubectl apply -f crds/
      ```

      {{<note>}}Ignore the following warning, as it is expected.{{</note>}}

      ``` text
      Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply.
      ```

### Upgrade NGINX Gateway Fabric release

#### Upgrade from the OCI registry

- To upgrade to the latest stable release of NGINX Gateway Fabric, run:

   ```shell
   helm upgrade <my-release> oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric -n nginx-gateway
   ```

   Replace `<my-release>` with your chosen release name.

#### Upgrade from sources

1. {{<include "installation/helm/pulling-the-chart.md" >}}

1. To upgrade, run: the following command:

   ```shell
   helm upgrade <my-release> . -n nginx-gateway
   ```

   Replace `<my-release>` with your chosen release name.

## Delay pod termination for zero downtime upgrades {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

{{< include "installation/delay-pod-termination/delay-pod-termination-overview.md" >}}

Follow these steps to configure delayed pod termination:

1. Open the `values.yaml` for editing.

1. **Add delayed shutdown hooks**:

   - In the `values.yaml` file, add `lifecycle: preStop` hooks to both the `nginx` and `nginx-gateway` container definitions. These hooks instruct the containers to delay their shutdown process, allowing time for connections to close gracefully. Update the `sleep` value to what works for your environment.

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

1. **Set the termination grace period**:

   - {{<include "installation/delay-pod-termination/termination-grace-period.md">}}

1. Save the changes.

{{<see-also>}} 
For additional information on configuring and understanding the behavior of containers and pods during their lifecycle, refer to the following Kubernetes documentation:
- [Container Lifecycle Hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks)
- [Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/Pods/Pod-lifecycle/#Pod-termination)
{{</see-also>}} 


## Uninstall NGINX Gateway Fabric

Follow these steps to uninstall NGINX Gateway Fabric and Gateway API from your Kubernetes cluster:

1. **Uninstall NGINX Gateway Fabric:**

   - To uninstall NGINX Gateway Fabric, run:

      ```shell
      helm uninstall <my-release> -n nginx-gateway
      ```

      Replace `<my-release>` with your chosen release name.

2. **Remove namespace and CRDs:**

   - To remove the **nginx-gateway** namespace and its custom resource definitions (CRDs), run:

      ```shell
      kubectl delete ns nginx-gateway
      kubectl delete crd nginxgateways.gateway.nginx.org
      ```

3. **Remove the Gateway API resources:**

   - {{<include "installation/helm/uninstall-gateway-api-resources.md" >}}

## Next steps

### Expose NGINX Gateway Fabric

{{<include "installation/next-step-expose-fabric.md">}}

