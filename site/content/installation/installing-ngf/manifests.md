---
title: "Installation with Manifests"
description: "Learn how to install, upgrade, and uninstall NGINX Gateway Fabric using manifest deployments in a Kubernetes cluster. This guide offers clear, step-by-step instructions to get you started."
weight: 100
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Prerequisites

In order to complete the steps in this guide, you must first:

- Install [kubectl](https://kubernetes.io/docs/tasks/tools/), a command-line interface for managing Kubernetes clusters.


## Deploy NGINX Gateway Fabric from Manifests

Deploying NGINX Gateway Fabric using Kubernetes manifests is a straightforward process that involves setting up necessary resources and deploying NGINX Gateway Fabric components within your cluster. This method allows for a detailed and controlled deployment, suitable for environments where customization and precise configuration are required.

{{<note>}}NGINX Gateway Fabric installs into the **nginx-gateway** namespace by default. To run NGINX Gateway Fabric in a different namespace, modify the installation manifests.{{</note>}}

1. Install the Gateway API resources from the standard channel (the CRDs and validating webhook):

   ```shell
   kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
   ```

2. Deploy the NGINX Gateway Fabric CRDs:

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
   ```

3. Deploy NGINX Gateway Fabric:

   ```shell
   kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
   ```

4. Confirm the NGINX Gateway Fabric is running in the `nginx-gateway` namespace:

   ```shell
   kubectl get pods -n nginx-gateway
   ```

   Expected output (note that `5d4f4c7db7-xk2kq` is a randomly generated string and will vary):

   ```text
   NAME                             READY   STATUS    RESTARTS   AGE
   nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
   ```

## Upgrade NGINX Gateway Fabric from Manifests

This section provides guidelines for upgrading your NGINX Gateway Fabric deployment to ensure you are using the latest features and improvements.

{{<tip>}}For zero-downtime upgrades, follow the instructions to [configure a delayed pod termination](#configure-delayed-pod-termination-for-zero-downtime-upgrades) for the NGINX Gateway Fabric pod.{{</tip>}}

Upgrading NGINX Gateway Fabric from manifests involves several steps to ensure all components are updated to their latest versions.

1. **Upgrade Gateway Resources:**
   - Check that the **Gateway API** resources are compatible with your version of NGINX Gateway Fabric ([Technical Specifications](/README.md#technical-specifications)).
   - Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.1) for any important upgrade-specific information.
   - To upgrade the gateway resources, run:

     ```shell
     kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
     ```

2. **Upgrade NGINX Gateway Fabric CRDs:**
   - To upgrade the Custom Resource Definitions (CRDs), run:

     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
     ```

3. **Upgrade NGINX Gateway Fabric Deployment:**
   - To upgrade the deployment, run:

     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
     ```



### Configure Delayed Pod Termination for Zero-Downtime Upgrades {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

In order to achieve zero-downtime upgrades and maintain continuous service availability, it's important to configure delayed pod termination. This setup is especially critical in environments that manage persistent or long-lived connections.

{{<note>}}NGINX won't shut down until all websocket or long-lived connections are closed. Keeping these connections open during an upgrade can lead to Kubernetes forcibly shutting down NGINX, potentially causing downtime for clients.{{</note>}}

1. Open the `nginx-gateway.yaml` for editing.

1. **Add delayed shutdown hooks**:

   In the `nginx-gateway.yaml` file, add `lifecycle: preStop` hooks to both the `nginx` and `nginx-gateway` container definitions. These hooks instruct the containers to delay their shutdown process, allowing time for connections to close gracefully.

   ```yaml
   <...>
   name: nginx-gateway
   <...>
   lifecycle:
     preStop:
       exec:
         command:
         - /usr/bin/gateway
         - sleep
         - --duration=40s # This flag is optional, the default is 30s
   <...>
   name: nginx
   <...>
   lifecycle:
     preStop:
       exec:
         command:
         - /bin/sleep
         - "40"
   <...>
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


## Uninstall NGINX Gateway Fabric from Manifests

Uninstalling NGINX Gateway Fabric from your Kubernetes cluster involves removing the deployed NGINX Gateway Fabric components and the Gateway API resources. This procedure should be followed carefully to ensure that all relevant components are cleanly removed from your system.

1. **Uninstall NGINX Gateway Fabric:**

   - Run the following commands to remove the NGINX Gateway Fabric and its Custom Resource Definitions (CRDs):

     ```shell
     kubectl delete -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
     ```

     ```shell
     kubectl delete -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
     ```

2. **Remove the Gateway API resources:**

   - To uninstall the Gateway API resources, including the CRDs and the validating webhook, use the command below. Ensure no custom resources you wish to keep or other Gateway API implementations are running in the cluster before proceeding:

     {{<warning>}}This command will remove all corresponding custom resources in your cluster across all namespaces!{{</warning>}}

     ```shell
     kubectl delete -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
     ```
