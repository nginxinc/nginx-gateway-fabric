---
title: "Installation with Helm"
description: "Learn how to install, upgrade, and uninstall NGINX Gateway Fabric in a Kubernetes cluster with using Helm."
weight: 200
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Prerequisites

To complete this guide, you'll need to install:

- [kubectl](https://kubernetes.io/docs/tasks/tools/), a command-line tool for managing Kubernetes clusters.
- [Helm 3.0 or later](https://helm.sh/docs/intro/install/), for deploying and managing applications on Kubernetes.


## Deploy NGINX Gateway Fabric with Helm

### Install from the OCI Registry

- To install the latest stable release of the NGINX Gateway Fabric in the **nginx-gateway** namespace, run the following command. Change `<my-release>` to the name you want for your release. If the namespace already exists, you can omit the optional `--create-namespace` flag.

   ```shell
   helm install <my-release> oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric --create-namespace --wait -n nginx-gateway
   ```

   If you want the latest version from the **main** branch, add `--version 0.0.0-edge` to your install command.

### Install from Sources{#install-from-sources}

1. {{<include "installation/helm/pulling-the-chart.md" >}}

2. To install the chart into the **nginx-gateway** namespace, run the following command. Change `<my-release>` to the name you want for your release. If the namespace already exists, you can omit the optional `--create-namespace` flag.

   ```shell
   helm install <my-release> . --create-namespace --wait -n nginx-gateway
   ```




## Upgrade NGINX Gateway Fabric Using Helm

Upgrading your NGINX Gateway Fabric deployment is crucial to take advantage of the latest features, security updates, and performance improvements. This section guides you through the upgrade process, ensuring a smooth transition to the latest version of NGINX Gateway Fabric. Pay attention to the compatibility of Gateway API resources and follow each step carefully for a successful upgrade.

{{<tip>}}For guidance on zero-downtime upgrades (ensuring service continuity without interruptions during upgrades), see [Configure Delayed Pod Termination](#configure-delayed-pod-termination-for-zero-downtime-upgrades).{{</tip>}}


### Upgrade Gateway Resources

Keeping your Gateway API resources up-to-date is a key part of maintaining NGINX Gateway Fabric. This subsection focuses on verifying and upgrading these resources to ensure they are compatible with the latest version of your NGINX Gateway Fabric.

To upgrade your Gateway API resources, take the following steps:

- Verify the Gateway API resources are compatible with your NGINX Gateway Fabric version. Refer to the [Technical Specifications]({{< relref "reference/technical-specifications.md" >}}) for details.
- Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.1) for any important upgrade-specific information.
- To upgrade the Gateway API resources, run:

   ```shell
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

### Upgrade NGINX Gateway Fabric CRDs

Helm's upgrade process does not automatically upgrade the NGINX Gateway Fabric CRDs (Custom Resource Definitions). This subsection provides the necessary steps to manually upgrade your CRDs.


1. {{<include "installation/helm/pulling-the-chart.md" >}}

2. To upgrade the Custom Resource Definitions (CRDs), run:

      ```shell
      kubectl apply -f crds/
      ```

      {{<note>}}Ignore the following warning, as it is expected.{{</note>}}

      ```
      Warning: kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply.
      ```

### Upgrade NGINX Gateway Fabric Release

Upgrading your NGINX Gateway Fabric release is a crucial step to keep your deployment current. This subsection covers two methods for upgrading: 

- Upgrade from the OCI Registry
- Upgrade from Sources

Choose the method that best fits your setup and follow the corresponding steps to ensure a successful upgrade.

#### Upgrade from the OCI Registry

- To upgrade to the latest stable release, run the following command, replacing `<my-release>` with your chosen release name:

   ```shell
   helm upgrade <my-release> oci://ghcr.io/nginxinc/charts/nginx-gateway-fabric -n nginx-gateway
   ```

#### Upgrade from Sources

1. {{<include "installation/helm/pulling-the-chart.md" >}}

2. To upgrade your release, run the following command, replacing `<my-release>` with your chosen release name:

   ```shell
   helm upgrade <my-release> . -n nginx-gateway
   ```

## Configure Delayed Pod Termination for Zero-Downtime Upgrades {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

In order to achieve zero-downtime upgrades and maintain continuous service availability, it's important to configure delayed pod termination. This setup is especially critical in environments that manage persistent or long-lived connections.

{{<note>}}NGINX won't shut down until all websocket or long-lived connections are closed. Keeping these connections open during an upgrade can lead to Kubernetes forcibly shutting down NGINX, potentially causing downtime for clients.{{</note>}}

1. Open the `values.yaml` for editing.

1. **Add delayed shutdown hooks**:

   In the `values.yaml` file, add `lifecycle: preStop` hooks to both the `nginx` and `nginxGateway` container definitions. These hooks instruct the containers to delay their shutdown process, allowing time for connections to close gracefully.

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

   Set `terminationGracePeriodSeconds` to a value that is equal to or greater than the `sleep` duration specified in the `preStop` hook (default is `30`). This setting prevents Kubernetes from terminating the pod before before the `preStop` hook has completed running.

   ```yaml
   terminationGracePeriodSeconds: 50
   ```

1. Save the changes.

{{<see-also>}} 
For additional information on configuring and understanding the behavior of containers and pods during their lifecycle, refer to the following Kubernetes documentation:
- [Container Lifecycle Hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/#container-hooks)
- [Pod Lifecycle](https://kubernetes.io/docs/concepts/workloads/Pods/Pod-lifecycle/#Pod-termination)
{{</see-also>}} 


## Uninstall NGINX Gateway Fabric Using Helm

Uninstalling NGINX Gateway Fabric is a straightforward process but requires careful attention to remove all related components. Follow these steps to uninstall the release and clean up resources:

1. **Uninstall NGINX Gateway Fabric:**

   - Run the following command uninstall the NGINX Gateway Fabric release, replacing `<my-release>` with your chosen release name:

      ```shell
      helm uninstall <my-release> -n nginx-gateway
      ```

2. **Remove namespace and CRDs:**

   - Remove the **nginx-gateway** namespace and CRDs to completely clean up all associated resources:

      ```shell
      kubectl delete ns nginx-gateway
      kubectl delete crd nginxgateways.gateway.nginx.org
      ```

      These commands will remove all Kubernetes components associated with the release and delete the release itself.

3. **Remove the Gateway API resources:**

   - {{<include "installation/helm/uninstall-gateway-api-resources.md" >}}



## Expose NGINX Gateway Fabric

Once NGINX Gateway Fabric is installed, the next step is to make it accessible. Refer to the following instructions for guidance on configuring access and creating the necessary services:

- [Expose the NGINX Gateway Fabric]({{< relref "installation/expose-nginx-gateway-fabric.md" >}}).

