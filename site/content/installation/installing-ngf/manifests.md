---
title: "Installation with Manifests"
description: "Learn how to install, upgrade, and uninstall NGINX Gateway Fabric using Kubernetes manifests."
weight: 100
toc: true
docs: "DOCS-000"
---

{{<custom-styles>}}

## Prerequisites

To complete this guide, you'll need to:

- Install [kubectl](https://kubernetes.io/docs/tasks/tools/), a command-line interface for managing Kubernetes clusters.


## Deploy NGINX Gateway Fabric from Manifests

Deploying NGINX Gateway Fabric with Kubernetes manifests takes only a few steps. With manifests, you can configure your deployment exactly how you want. Manifests also make it easy to replicate deployments across environments or clusters, ensuring consistency.

{{<note>}}By default, NGINX Gateway Fabric is installed in the **nginx-gateway** namespace. You can deploy in another namespace by modifying the manifest files.{{</note>}}

1. **Install the Gateway API Resources:**
   - Start by installing the Gateway API resources, including the CRDs and the validating webhook:
     ```shell
     kubectl apply -f https://github.com/kubernetes-sigs/gateway-api/releases/download/v0.8.1/standard-install.yaml
     ```

2. **Deploy the NGINX Gateway Fabric CRDs:**
   - Next, deploy the NGINX Gateway Fabric CRDs:
     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/crds.yaml
     ```

3. **Deploy NGINX Gateway Fabric:**
   - Then, deploy NGINX Gateway Fabric:
     ```shell
     kubectl apply -f https://github.com/nginxinc/nginx-gateway-fabric/releases/download/v1.0.0/nginx-gateway.yaml
     ```

4. **Verify the Deployment:**
   - To confirm that NGINX Gateway Fabric is running, check the pods in the `nginx-gateway` namespace:
     ```shell
     kubectl get pods -n nginx-gateway
     ```
     The output should look similar to this (note that the pod name will include a unique string):
     ```text
     NAME                             READY   STATUS    RESTARTS   AGE
     nginx-gateway-5d4f4c7db7-xk2kq   2/2     Running   0          112s
     ```


## Upgrade NGINX Gateway Fabric from Manifests

{{<tip>}}For guidance on zero downtime upgrades, see the [Delay Pod Termination](#configure-delayed-pod-termination-for-zero-downtime-upgrades) section below.{{</tip>}}

To upgrade NGINX Gateway Fabric and get the latest features and improvements, take the following steps:

1. **Upgrade Gateway Resources:**

    - Verify that your NGINX Gateway Fabric version is compatible with the Gateway API resources. Refer to the [Technical Specifications]({{< relref "reference/technical-specifications.md" >}}) for details.
   - Review the [release notes](https://github.com/kubernetes-sigs/gateway-api/releases/tag/v0.8.1) for any important upgrade-specific information.
   - To upgrade the Gateway API resources, run:

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

## Delay Pod Termination for Zero Downtime Upgrades {#configure-delayed-pod-termination-for-zero-downtime-upgrades}

To avoid client service interruptions when upgrading NGINX Gateway Fabric, you can configure [`PreStop` hooks](https://kubernetes.io/docs/concepts/containers/container-lifecycle-hooks/) to delay terminating the NGINX Gateway Fabric pod, allowing the pod to complete certain actions before shutting down. This ensures a smooth upgrade without any downtime, also known as a zero downtime upgrade. 

For an in-depth explanation of how Kubernetes handles pod termination, see the [Termination of Pods](https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/#pod-termination) topic on their official website.

{{<note>}}Keep in mind that NGINX won't shut down while WebSocket or other long-lived connections are open. NGINX will only stop when these connections are closed by the client or the backend. If these connections stay open during an upgrade, Kubernetes might need to shut down NGINX forcefully. This sudden shutdown could interrupt service for clients.{{</note>}}

Follow these steps to configure delayed pod termination:

1. Open the `nginx-gateway.yaml` for editing.

2. **Add delayed shutdown hooks**:

   In the `nginx-gateway.yaml` file, add `lifecycle: preStop` hooks to both the `nginx` and `nginx-gateway` container definitions. These hooks instruct the containers to delay their shutdown process, allowing time for connections to close gracefully. Update the `sleep` value to what works for your environment.

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

3. **Set the termination grace period**:

   Set `terminationGracePeriodSeconds` to a value that is equal to or greater than the `sleep` duration specified in the `preStop` hook (default is `30`). This setting prevents Kubernetes from terminating the pod before before the `preStop` hook has completed running.

   ```yaml
   terminationGracePeriodSeconds: 50
   ```

4. Save the changes.

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

   - {{<include "installation/helm/uninstall-gateway-api-resources.md" >}}

## Expose NGINX Gateway Fabric

Once NGINX Gateway Fabric is installed, the next step is to make it accessible. Refer to the following instructions for guidance on configuring access and creating the necessary services:

- [Expose the NGINX Gateway Fabric]({{< relref "installation/expose-nginx-gateway-fabric.md" >}}).

